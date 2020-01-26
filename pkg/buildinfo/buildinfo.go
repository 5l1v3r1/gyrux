// Package buildinfo contains build information.
//
// Build information should be set during compilation by passing
// -ldflags "-X github.com/entynetproject/gyrux/pkg/buildinfo.Var=value" to "go build" or
// "go get".
package buildinfo

// Build information.
var (
	Version      = "unknown"
	Reproducible = "false"
)
