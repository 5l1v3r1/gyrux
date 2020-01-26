package shell

import (
	"os"

	"github.com/entynetproject/gyrux/pkg/util"
)

// Ensures Gyrux's data directory exists, creating it if necessary. It returns
// the path to the data directory (never with a trailing slash) and possible
// error.
func ensureDataDir() (string, error) {
	home, err := util.GetHome("")
	if err != nil {
		return "", err
	}
	ddir := home + "/.gyrux"
	return ddir, os.MkdirAll(ddir, 0700)
}
