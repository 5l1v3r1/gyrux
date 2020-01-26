package edit

import (
	"github.com/entynetproject/gyrux/pkg/cli"
	"github.com/entynetproject/gyrux/pkg/cli/addons/stub"
	"github.com/entynetproject/gyrux/pkg/eval"
)

//gy:fn command:start
//
// Starts the command mode.

func initCommandAPI(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	ns.AddNs("command",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:command>:", map[string]interface{}{
			"start": func() {
				stub.Start(app, stub.Config{
					Binding: binding,
					Name:    " COMMAND ",
					Focus:   false,
				})
			},
		}))
}
