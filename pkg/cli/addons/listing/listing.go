// Package listing provides the custom listing addon.
package listing

import (
	"github.com/entynetproject/gyrux/pkg/cli"
	"github.com/entynetproject/gyrux/pkg/ui"
)

// Config is the configuration to start the custom listing addon.
type Config struct {
	// Keybinding.
	Binding cli.Handler
	// Caption of the listing. If empty, defaults to " LISTING ".
	Caption string
	// A function that takes the query string and returns a list of Item's and
	// the index of the Item to select. Required.
	GetItems func(query string) (items []Item, selected int)
	// A function to call when the user has accepted the selected item. If the
	// return value is true, the listing will not be closed after accpeting.
	// If unspecified, the Accept function default to a function that does
	// nothing other than returning false.
	Accept func(string) bool
	// Whether to automatically accept when there is only one item.
	AutoAccept bool
}

// Item is an item to show in the listing.
type Item struct {
	// Passed to the Accept callback in Config.
	ToAccept string
	// How the item is shown.
	ToShow ui.Text
}

// Start starts the custom listing addon.
func Start(app cli.App, cfg Config) {
	if cfg.GetItems == nil {
		app.Notify("internal error: GetItems must be specified")
		return
	}
	if cfg.Accept == nil {
		cfg.Accept = func(string) bool { return false }
	}
	if cfg.Caption == "" {
		cfg.Caption = " LISTING "
	}
	accept := func(s string) {
		retain := cfg.Accept(s)
		if !retain {
			cli.SetAddon(app, nil)
		}
	}
	w := cli.NewComboBox(cli.ComboBoxSpec{
		CodeArea: cli.CodeAreaSpec{
			Prompt: cli.ModePrompt(cfg.Caption, true),
		},
		ListBox: cli.ListBoxSpec{
			OverlayHandler: cfg.Binding,
			OnAccept: func(it cli.Items, i int) {
				accept(it.(items)[i].ToAccept)
			},
			ExtendStyle: true,
		},
		OnFilter: func(w cli.ComboBox, q string) {
			it, selected := cfg.GetItems(q)
			w.ListBox().Reset(items(it), selected)
			if cfg.AutoAccept && len(it) == 1 {
				accept(it[0].ToAccept)
			}
		},
	})
	cli.SetAddon(app, w)
	app.Redraw()
}

type items []Item

func (it items) Len() int           { return len(it) }
func (it items) Show(i int) ui.Text { return it[i].ToShow }
