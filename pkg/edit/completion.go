package edit

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/entynetproject/gyrux/pkg/cli"
	"github.com/entynetproject/gyrux/pkg/cli/addons/completion"
	"github.com/entynetproject/gyrux/pkg/edit/complete"
	"github.com/entynetproject/gyrux/pkg/eval"
	"github.com/entynetproject/gyrux/pkg/eval/vals"
	"github.com/entynetproject/gyrux/pkg/parse"
	"github.com/entynetproject/gyrux/pkg/util"
	"github.com/xiaq/persistent/hash"
)

type complexCandidateOpts struct {
	CodeSuffix    string
	DisplaySuffix string
	Display       string
}

func (*complexCandidateOpts) SetDefaultOptions() {}

func complexCandidate(opts complexCandidateOpts, stem string) complexItem {
	display := opts.Display
	if display == "" {
		// TODO(#898): Deprecate DisplaySuffix and remove this branch.
		display = stem + opts.DisplaySuffix
	}
	return complexItem{
		Stem:       stem,
		CodeSuffix: opts.CodeSuffix,
		Display:    display,
	}
}

func completionStart(app cli.App, binding cli.Handler, cfg complete.Config, smart bool) {
	buf := app.CodeArea().CopyState().Buffer
	result, err := complete.Complete(
		complete.CodeBuffer{Content: buf.Content, Dot: buf.Dot}, cfg)
	if err != nil {
		app.Notify(err.Error())
		return
	}
	if smart {
		prefix := ""
		for i, item := range result.Items {
			if i == 0 {
				prefix = item.ToInsert
				continue
			}
			prefix = commonPrefix(prefix, item.ToInsert)
			if prefix == "" {
				break
			}
		}
		if prefix != "" {
			insertedPrefix := false
			app.CodeArea().MutateState(func(s *cli.CodeAreaState) {
				rep := s.Buffer.Content[result.Replace.From:result.Replace.To]
				if len(prefix) > len(rep) && strings.HasPrefix(prefix, rep) {
					s.Pending = cli.PendingCode{
						Content: prefix,
						From:    result.Replace.From, To: result.Replace.To}
					s.ApplyPending()
					insertedPrefix = true
				}
			})
			if insertedPrefix {
				return
			}
		}
	}
	completion.Start(app, completion.Config{
		Name: result.Name, Replace: result.Replace, Items: result.Items,
		Binding: binding})
}

func initCompletion(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	matcherMapVar := newMapVar(vals.EmptyMap)
	argGeneratorMapVar := newMapVar(vals.EmptyMap)
	cfg := func() complete.Config {
		return complete.Config{
			PureEvaler: pureEvaler{ev},
			Filterer: adaptMatcherMap(
				app, ev, matcherMapVar.Get().(vals.Map)),
			ArgGenerator: adaptArgGeneratorMap(
				ev, argGeneratorMapVar.Get().(vals.Map)),
		}
	}
	generateForSudo := func(args []string) ([]complete.RawItem, error) {
		return complete.GenerateForSudo(cfg(), args)
	}
	ns.AddGoFns("<edit>", map[string]interface{}{
		"complete-filename": wrapArgGenerator(complete.GenerateFileNames),
		"complete-getopt":   completeGetopt,
		"complete-sudo":     wrapArgGenerator(generateForSudo),
		"complex-candidate": complexCandidate,
		"match-prefix":      wrapMatcher(strings.HasPrefix),
		"match-subseq":      wrapMatcher(util.HasSubseq),
		"match-substr":      wrapMatcher(strings.Contains),
	})
	ns.AddNs("completion",
		eval.Ns{
			"arg-completer": argGeneratorMapVar,
			"binding":       bindingVar,
			"matcher":       matcherMapVar,
		}.AddGoFns("<edit:completion>:", map[string]interface{}{
			"accept":      func() { listingAccept(app) },
			"smart-start": func() { completionStart(app, binding, cfg(), true) },
			"start":       func() { completionStart(app, binding, cfg(), false) },
			"close":       func() { completion.Close(app) },
			"up":          func() { listingUp(app) },
			"down":        func() { listingDown(app) },
			"up-cycle":    func() { listingUpCycle(app) },
			"down-cycle":  func() { listingDownCycle(app) },
			"left":        func() { listingLeft(app) },
			"right":       func() { listingRight(app) },
		}))
}

// A wrapper type implementing Gyrux value methods.
type complexItem complete.ComplexItem

func (c complexItem) Index(k interface{}) (interface{}, bool) {
	switch k {
	case "stem":
		return c.Stem, true
	case "code-suffix":
		return c.CodeSuffix, true
	case "display":
		return c.Display, true
	}
	return nil, false
}

func (c complexItem) IterateKeys(f func(interface{}) bool) {
	util.Feed(f, "stem", "code-suffix", "display")
}

func (c complexItem) Kind() string { return "map" }

func (c complexItem) Equal(a interface{}) bool {
	rhs, ok := a.(complexItem)
	return ok && c.Stem == rhs.Stem &&
		c.CodeSuffix == rhs.CodeSuffix && c.Display == rhs.Display
}

func (c complexItem) Hash() uint32 {
	h := hash.DJBInit
	h = hash.DJBCombine(h, hash.String(c.Stem))
	h = hash.DJBCombine(h, hash.String(c.CodeSuffix))
	h = hash.DJBCombine(h, hash.String(c.Display))
	return h
}

func (c complexItem) Repr(indent int) string {
	// TODO(gyrux): Pretty-print when indent >= 0
	return fmt.Sprintf("(edit:complex-candidate %s &code-suffix=%s &display=%s)",
		parse.Quote(c.Stem), parse.Quote(c.CodeSuffix), parse.Quote(c.Display))
}

type wrappedArgGenerator func(*eval.Frame, ...string) error

// Wraps an ArgGenerator into a function that can be then passed to
// eval.NewGoFn.
func wrapArgGenerator(gen complete.ArgGenerator) wrappedArgGenerator {
	return func(fm *eval.Frame, args ...string) error {
		rawItems, err := gen(args)
		if err != nil {
			return err
		}
		ch := fm.OutputChan()
		for _, rawItem := range rawItems {
			switch rawItem := rawItem.(type) {
			case complete.ComplexItem:
				ch <- complexItem(rawItem)
			case complete.PlainItem:
				ch <- string(rawItem)
			default:
				ch <- rawItem
			}
		}
		return nil
	}
}

func commonPrefix(s1, s2 string) string {
	for i, r := range s1 {
		if s2 == "" {
			break
		}
		r2, n2 := utf8.DecodeRuneInString(s2)
		if r2 != r {
			return s1[:i]
		}
		s2 = s2[n2:]
	}
	return s1
}

// The type for a native Go matcher. This is not equivalent to the Gyrux
// counterpart, which streams input and output. This is because we can actually
// afford calling a Go function for each item, so omitting the streaming
// behavior makes the implementation simpler.
//
// Native Go matchers are wrapped into Gyrux matchers, but never the other way
// around.
//
// This type is satisfied by strings.Contains and strings.HasPrefix; they are
// wrapped into match-substr and match-prefix respectively.
type matcher func(text, seed string) bool

type matcherOpts struct {
	IgnoreCase bool
	SmartCase  bool
}

func (*matcherOpts) SetDefaultOptions() {}

type wrappedMatcher func(fm *eval.Frame, opts matcherOpts, seed string, inputs eval.Inputs)

func wrapMatcher(m matcher) wrappedMatcher {
	return func(fm *eval.Frame, opts matcherOpts, seed string, inputs eval.Inputs) {
		out := fm.OutputChan()
		if opts.IgnoreCase || (opts.SmartCase && seed == strings.ToLower(seed)) {
			if opts.IgnoreCase {
				seed = strings.ToLower(seed)
			}
			inputs(func(v interface{}) {
				out <- m(strings.ToLower(vals.ToString(v)), seed)
			})
		} else {
			inputs(func(v interface{}) {
				out <- m(vals.ToString(v), seed)
			})
		}
	}
}

// Adapts $edit:completion:matcher into a Filterer.
func adaptMatcherMap(nt notifier, ev *eval.Evaler, m vals.Map) complete.Filterer {
	return func(ctxName, seed string, rawItems []complete.RawItem) []complete.RawItem {
		matcher, ok := lookupFn(m, ctxName)
		if !ok {
			nt.Notify(fmt.Sprintf(
				"matcher for %s not a function, falling back to prefix matching", ctxName))
		}
		if matcher == nil {
			return complete.FilterPrefix(ctxName, seed, rawItems)
		}
		input := make(chan interface{})
		stopInputFeeder := make(chan struct{})
		defer close(stopInputFeeder)
		// Feed a string representing all raw candidates to the input channel.
		go func() {
			defer close(input)
			for _, rawItem := range rawItems {
				select {
				case input <- rawItem.String():
				case <-stopInputFeeder:
					return
				}
			}
		}()
		ports := []*eval.Port{
			{Chan: input, File: eval.DevNull},
			{}, // Will be replaced in CaptureOutput
			{File: os.Stderr},
		}
		fm := eval.NewTopFrame(ev, eval.NewInternalGoSource("[editor matcher]"), ports)
		outputs, err := fm.CaptureOutput(matcher, []interface{}{seed}, eval.NoOpts)
		if err != nil {
			nt.Notify(fmt.Sprintf("[matcher error] %s", err))
			// Continue with whatever values have been output
		}
		if len(outputs) != len(rawItems) {
			nt.Notify(fmt.Sprintf(
				"matcher has output %v values, not equal to %v inputs",
				len(outputs), len(rawItems)))
		}
		filtered := []complete.RawItem{}
		for i := 0; i < len(rawItems) && i < len(outputs); i++ {
			if vals.Bool(outputs[i]) {
				filtered = append(filtered, rawItems[i])
			}
		}
		return filtered
	}
}

func adaptArgGeneratorMap(ev *eval.Evaler, m vals.Map) complete.ArgGenerator {
	return func(args []string) ([]complete.RawItem, error) {
		gen, ok := lookupFn(m, args[0])
		if !ok {
			return nil, fmt.Errorf("arg completer for %s not a function", args[0])
		}
		if gen == nil {
			return complete.GenerateFileNames(args)
		}
		argValues := make([]interface{}, len(args))
		for i, arg := range args {
			argValues[i] = arg
		}
		ports := []*eval.Port{
			eval.DevNullClosedChan,
			{}, // Will be replaced in CaptureOutput
			{File: os.Stderr},
		}
		var output []complete.RawItem
		var outputMutex sync.Mutex
		collect := func(item complete.RawItem) {
			outputMutex.Lock()
			defer outputMutex.Unlock()
			output = append(output, item)
		}
		valueCb := func(ch <-chan interface{}) {
			for v := range ch {
				switch v := v.(type) {
				case string:
					collect(complete.PlainItem(v))
				case complexItem:
					collect(complete.ComplexItem(v))
				default:
					collect(complete.PlainItem(vals.ToString(v)))
				}
			}
		}
		bytesCb := func(r *os.File) {
			buffered := bufio.NewReader(r)
			for {
				line, err := buffered.ReadString('\n')
				if line != "" {
					collect(complete.PlainItem(strings.TrimSuffix(line, "\n")))
				}
				if err != nil {
					break
				}
			}
		}
		fm := eval.NewTopFrame(ev, eval.NewInternalGoSource("[editor arg generator]"), ports)
		err := fm.CallWithOutputCallback(gen, argValues, eval.NoOpts, valueCb, bytesCb)
		return output, err
	}
}

func lookupFn(m vals.Map, ctxName string) (eval.Callable, bool) {
	val, ok := m.Index(ctxName)
	if !ok {
		val, ok = m.Index("")
	}
	if !ok {
		// No matcher, but not an error either
		return nil, true
	}
	fn, ok := val.(eval.Callable)
	if !ok {
		return nil, false
	}
	return fn, true
}

type pureEvaler struct{ ev *eval.Evaler }

func (pureEvaler) EachExternal(f func(string)) { eval.EachExternal(f) }

func (pureEvaler) EachSpecial(f func(string)) {
	for name := range eval.IsBuiltinSpecial {
		f(name)
	}
}

func (pe pureEvaler) EachNs(f func(string)) { pe.ev.EachNsInTop(f) }

func (pe pureEvaler) EachVariableInNs(ns string, f func(string)) {
	pe.ev.EachVariableInTop(ns, f)
}

func (pe pureEvaler) PurelyEvalPrimary(pn *parse.Primary) interface{} {
	return pe.ev.PurelyEvalPrimary(pn)
}

func (pe pureEvaler) PurelyEvalCompound(cn *parse.Compound) (string, error) {
	return pe.ev.PurelyEvalCompound(cn)
}

func (pe pureEvaler) PurelyEvalPartialCompound(cn *parse.Compound, in *parse.Indexing) (string, error) {
	return pe.ev.PurelyEvalPartialCompound(cn, in)
}
