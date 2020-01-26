// Package lastcmd implements an addon that supports inserting the last command
// or words from it.
package lastcmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/entynetproject/gyrux/pkg/cli"
	"github.com/entynetproject/gyrux/pkg/cli/histutil"
	"github.com/entynetproject/gyrux/pkg/store"
	"github.com/entynetproject/gyrux/pkg/ui"
)

// Config is the configuration for starting lastcmd.
type Config struct {
	// Binding provides key binding.
	Binding cli.Handler
	// Store provides the source for the last command.
	Store Store
	// Wordifier breaks a command into words.
	Wordifier func(string) []string
}

// Store wraps the LastCmd method. It is a subset of histutil.Store.
type Store interface {
	LastCmd() (store.Cmd, error)
}

var _ = Store(histutil.Store(nil))

// Start starts lastcmd function.
func Start(app cli.App, cfg Config) {
	if cfg.Store == nil {
		app.Notify("no history store")
		return
	}
	cmd, err := cfg.Store.LastCmd()
	if err != nil {
		app.Notify("db error: " + err.Error())
		return
	}
	wordifier := cfg.Wordifier
	if wordifier == nil {
		wordifier = strings.Fields
	}
	cmdText := cmd.Text
	words := wordifier(cmdText)
	entries := make([]entry, len(words)+1)
	entries[0] = entry{content: cmdText}
	for i, word := range words {
		entries[i+1] = entry{strconv.Itoa(i), strconv.Itoa(i - len(words)), word}
	}

	accept := func(text string) {
		app.CodeArea().MutateState(func(s *cli.CodeAreaState) {
			s.Buffer.InsertAtDot(text)
		})
		app.MutateState(func(s *cli.State) { s.Addon = nil })
	}
	w := cli.NewComboBox(cli.ComboBoxSpec{
		CodeArea: cli.CodeAreaSpec{Prompt: cli.ModePrompt(" LASTCMD ", true)},
		ListBox: cli.ListBoxSpec{
			OverlayHandler: cfg.Binding,
			OnAccept: func(it cli.Items, i int) {
				accept(it.(items).entries[i].content)
			},
		},
		OnFilter: func(w cli.ComboBox, p string) {
			items := filter(entries, p)
			if len(items.entries) == 1 {
				accept(items.entries[0].content)
			} else {
				w.ListBox().Reset(items, 0)
			}
		},
	})
	app.MutateState(func(s *cli.State) { s.Addon = w })
	app.Redraw()
}

type items struct {
	negFilter bool
	entries   []entry
}

type entry struct {
	posIndex string
	negIndex string
	content  string
}

func filter(allEntries []entry, p string) items {
	if p == "" {
		return items{false, allEntries}
	}
	var entries []entry
	negFilter := strings.HasPrefix(p, "-")
	for _, entry := range allEntries {
		if (negFilter && strings.HasPrefix(entry.negIndex, p)) ||
			(!negFilter && strings.HasPrefix(entry.posIndex, p)) {
			entries = append(entries, entry)
		}
	}
	return items{negFilter, entries}
}

func (it items) Show(i int) ui.Text {
	index := ""
	entry := it.entries[i]
	if it.negFilter {
		index = entry.negIndex
	} else {
		index = entry.posIndex
	}
	// NOTE: We now use a hardcoded width of 3 for the index, which will work as
	// long as the command has less than 1000 words (when filter is positive) or
	// 100 words (when filter is negative).
	return ui.T(fmt.Sprintf("%3s %s", index, entry.content))
}

func (it items) Len() int { return len(it.entries) }
