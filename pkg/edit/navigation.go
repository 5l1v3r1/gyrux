package edit

import (
	"github.com/entynetproject/gyrux/pkg/cli"
	"github.com/entynetproject/gyrux/pkg/cli/addons/navigation"
	"github.com/entynetproject/gyrux/pkg/eval"
	"github.com/entynetproject/gyrux/pkg/eval/vals"
	"github.com/entynetproject/gyrux/pkg/eval/vars"
	"github.com/entynetproject/gyrux/pkg/parse"
)

//gydoc:var selected-file
//
// Name of the currently selected file in navigation mode. $nil if not in
// navigation mode.

//gydoc:var navigation:binding
//
// Keybinding for the navigation mode.

//gydoc:fn navigation:start
//
// Start the navigation mode.

//gydoc:fn navigation:insert-selected
//
// Inserts the selected filename.

func navInsertSelected(app cli.App) {
	insertAtDot(app, " "+parse.Quote(navigation.SelectedName(app)))
}

//gydoc:fn navigation:insert-selected-and-quit
//
// Inserts the selected filename and closes the navigation addon.

func navInsertSelectedAndQuit(app cli.App) {
	navInsertSelected(app)
	closeListing(app)
}

//gydoc:fn navigation:trigger-filter
//
// Toggles the filtering status of the navigation addon.

func navToggleFilter(app cli.App) {
	navigation.MutateFiltering(app, func(b bool) bool { return !b })
}

//gydoc:fn navigation:trigger-shown-hidden
//
// Toggles whether the navigation addon should be showing hidden files.

func navToggleShowHidden(app cli.App) {
	navigation.MutateShowHidden(app, func(b bool) bool { return !b })
}

//gydoc:var navigation:width-ratio
//
// A list of 3 integers, used for specifying the width ratio of the 3 columns in
// navigation mode.

func convertNavWidthRatio(v interface{}) [3]int {
	var (
		numbers []int
		hasErr  bool
	)
	vals.Iterate(v, func(elem interface{}) bool {
		var i int
		err := vals.ScanToGo(elem, &i)
		if err != nil {
			hasErr = true
			return false
		}
		numbers = append(numbers, i)
		return true
	})
	if hasErr || len(numbers) != 3 {
		// TODO: Handle the error.
		return [3]int{1, 3, 4}
	}
	var ret [3]int
	copy(ret[:], numbers)
	return ret
}

func initNavigation(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(EmptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)
	widthRatioVar := newListVar(vals.MakeList(1.0, 3.0, 4.0))

	selectedFileVar := vars.FromGet(func() interface{} {
		name := navigation.SelectedName(app)
		if name == "" {
			return nil
		}
		return name
	})

	ns.Add("selected-file", selectedFileVar)
	ns.AddNs("navigation",
		eval.Ns{
			"binding":     bindingVar,
			"width-ratio": widthRatioVar,
		}.AddGoFns("<edit:navigation>", map[string]interface{}{
			"start": func() {
				navigation.Start(app, navigation.Config{
					Binding: binding,
					WidthRatio: func() [3]int {
						return convertNavWidthRatio(widthRatioVar.Get())
					},
				})
			},
			"left":      func() { navigation.Ascend(app) },
			"right":     func() { navigation.Descend(app) },
			"up":        func() { navigation.Select(app, cli.Prev) },
			"down":      func() { navigation.Select(app, cli.Next) },
			"page-up":   func() { navigation.Select(app, cli.PrevPage) },
			"page-down": func() { navigation.Select(app, cli.NextPage) },

			"file-preview-up":   func() { navigation.ScrollPreview(app, -1) },
			"file-preview-down": func() { navigation.ScrollPreview(app, 1) },

			"insert-selected":          func() { navInsertSelected(app) },
			"insert-selected-and-quit": func() { navInsertSelectedAndQuit(app) },

			"trigger-filter":       func() { navToggleFilter(app) },
			"trigger-shown-hidden": func() { navToggleShowHidden(app) },
		}))
}
