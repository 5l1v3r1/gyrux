// Package edit implements the line editor for Gyrux.
//
// The line editor is based on the cli package, which implements a general,
// Gyrux-agnostic line editor, and multiple "addon" packages. This package
// glues them together and provides Gyrux bindings for them.
package edit

import (
	"os"
	"strings"

	"github.com/entynetproject/gyrux/pkg/cli"
	"github.com/entynetproject/gyrux/pkg/cli/histutil"
	"github.com/entynetproject/gyrux/pkg/eval"
	"github.com/entynetproject/gyrux/pkg/parse"
	"github.com/entynetproject/gyrux/pkg/store"
)

// Editor is the interface line editor for Gyrux.
type Editor struct {
	app cli.App
	ns  eval.Ns
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(tty cli.TTY, ev *eval.Evaler, st store.Store) *Editor {
	ns := eval.NewNs()
	appSpec := cli.AppSpec{TTY: tty}

	fuser, err := histutil.NewFuser(st)
	if err != nil {
		// TODO(gyrux): Report the error.
	}

	if fuser != nil {
		appSpec.AfterReadline = []func(string){func(code string) {
			if code != "" && !strings.HasPrefix(code, " ") {
				fuser.AddCmd(code)
			}
			// TODO(gyrux): Handle the error.
		}}
	}

	// Make a variable for the app first. This is to work around the
	// bootstrapping of initPrompts, which expects a notifier.
	var app cli.App
	initHighlighter(&appSpec, ev)
	initConfigAPI(&appSpec, ev, ns)
	initInsertAPI(&appSpec, appNotifier{&app}, ev, ns)
	initPrompts(&appSpec, appNotifier{&app}, ev, ns)
	app = cli.NewApp(appSpec)

	initCommandAPI(app, ev, ns)
	initListings(app, ev, ns, st, fuser)
	initNavigation(app, ev, ns)
	initCompletion(app, ev, ns)
	initHistWalk(app, ev, ns, fuser)
	initInstant(app, ev, ns)
	initMinibuf(app, ev, ns)

	initBufferBuiltins(app, ns)
	initTTYBuiltins(app, tty, ns)
	initMiscBuiltins(app, ns)
	initStateAPI(app, ns)
	initStoreAPI(app, ns, fuser)
	evalDefaultBinding(ev, ns)

	return &Editor{app, ns}
}

func evalDefaultBinding(ev *eval.Evaler, ns eval.Ns) {
	// TODO(gyrux): The evaler API should accodomate the use case of evaluating a
	// piece of code in an alternative global namespace.

	n, err := parse.AsChunk("[default bindings]", defaultBindingscy)
	if err != nil {
		panic(err)
	}
	src := eval.NewInternalGyruxSource(
		true, "[default bindings]", defaultBindingscy)
	op, err := ev.CompileWithGlobal(n, src, ns)
	if err != nil {
		panic(err)
	}
	// TODO(gyrux): Use stdPorts when it is possible to do so.
	fm := eval.NewTopFrame(ev, src, []*eval.Port{
		{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr},
	})
	fm.SetLocal(ns)
	err = fm.Eval(op)
	if err != nil {
		panic(err)
	}
}

// ReadCode reads input from the user.
func (ed *Editor) ReadCode() (string, error) {
	return ed.app.ReadCode()
}

// Ns returns a namespace for manipulating the editor from Gyrux code.
func (ed *Editor) Ns() eval.Ns {
	return ed.ns
}
