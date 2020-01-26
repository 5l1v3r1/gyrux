package edit

import (
	"fmt"
	"os"

	"github.com/entynetproject/gyrux/pkg/cli"
	"github.com/entynetproject/gyrux/pkg/diag"
	"github.com/entynetproject/gyrux/pkg/eval"
	"github.com/entynetproject/gyrux/pkg/eval/vals"
	"github.com/entynetproject/gyrux/pkg/eval/vars"
)

func initConfigAPI(appSpec *cli.AppSpec, ev *eval.Evaler, ns eval.Ns) {
	initMaxHeight(appSpec, ns)
	initBeforeReadline(appSpec, ev, ns)
	initAfterReadline(appSpec, ev, ns)
}

//gydoc:var max-height
//
// Maximum height the editor is allowed to use, defaults to `+Inf`.
//
// By default, the height of the editor is only restricted by the terminal
// height. Some modes like location mode can use a lot of lines; as a result,
// it can often occupy the entire terminal, and push up your scrollback buffer.
// Change this variable to a finite number to restrict the height of the editor.

func initMaxHeight(appSpec *cli.AppSpec, ns eval.Ns) {
	maxHeight := newIntVar(-1)
	appSpec.MaxHeight = func() int { return maxHeight.GetRaw().(int) }
	ns.Add("max-height", maxHeight)
}

func initBeforeReadline(appSpec *cli.AppSpec, ev *eval.Evaler, ns eval.Ns) {
	hook := newListVar(vals.EmptyList)
	ns["before-readline"] = hook
	appSpec.BeforeReadline = append(appSpec.BeforeReadline, func() {
		callHooks(ev, "$<edit>:before-readline", hook.Get().(vals.List))
	})
}

func initAfterReadline(appSpec *cli.AppSpec, ev *eval.Evaler, ns eval.Ns) {
	hook := newListVar(vals.EmptyList)
	ns["after-readline"] = hook
	appSpec.AfterReadline = append(appSpec.AfterReadline, func(code string) {
		callHooks(ev, "$<edit>:after-readline", hook.Get().(vals.List), code)
	})
}

func callHooks(ev *eval.Evaler, name string, hook vals.List, args ...interface{}) {
	i := -1
	for it := hook.Iterator(); it.HasElem(); it.Next() {
		i++
		name := fmt.Sprintf("%s[%d]", name, i)
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			// TODO(gyrux): This is not testable as it depends on stderr.
			// Make it testable.
			diag.Complainf("%s not function", name)
			continue
		}
		// TODO(gyrux): This should use stdPorts, but stdPorts is currently
		// unexported from eval.
		ports := []*eval.Port{
			{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr}}
		fm := eval.NewTopFrame(ev, eval.NewInternalGoSource(name), ports)
		fm.Call(fn, args, eval.NoOpts)
	}
}

func newIntVar(i int) vars.PtrVar            { return vars.FromPtr(&i) }
func newFloatVar(f float64) vars.PtrVar      { return vars.FromPtr(&f) }
func newBoolVar(b bool) vars.PtrVar          { return vars.FromPtr(&b) }
func newListVar(l vals.List) vars.PtrVar     { return vars.FromPtr(&l) }
func newMapVar(m vals.Map) vars.PtrVar       { return vars.FromPtr(&m) }
func newFnVar(c eval.Callable) vars.PtrVar   { return vars.FromPtr(&c) }
func newBindingVar(b BindingMap) vars.PtrVar { return vars.FromPtr(&b) }
