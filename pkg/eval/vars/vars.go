// Package vars contains basic types for manipulating Gyrux variables.
package vars

// Var represents an Gyrux variable.
type Var interface {
	Set(v interface{}) error
	Get() interface{}
}
