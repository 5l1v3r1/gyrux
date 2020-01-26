// Package bundled keeps bundled modules.
package bundled

// Get returns a map of bundled modules.
func Get() map[string]string {
	return map[string]string{
		"binding":          bindingcy,
		"gypm":             gypmgy,
		"narrow":           narrowcy,
		"readline-binding": readlineBindingcy,
	}
}
