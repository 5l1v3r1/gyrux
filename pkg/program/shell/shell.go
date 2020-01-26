// Package shell is the entry point for the terminal interface of Gyrux.
package shell

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/entynetproject/gyrux/pkg/cli/term"
	"github.com/entynetproject/gyrux/pkg/diag"
	"github.com/entynetproject/gyrux/pkg/sys"
	"github.com/entynetproject/gyrux/pkg/util"
)

var logger = util.GetLogger("[shell] ")

// Shell keeps flags to the shell.
type Shell struct {
	BinPath     string
	SockPath    string
	DbPath      string
	Cmd         bool
	CompileOnly bool
	NoRc        bool
	JSON        bool
}

// Main runs Gyrux using the default terminal interface. It blocks until Gyrux
// quits, and returns the exit code.
func (sh *Shell) Main(fds [3]*os.File, args []string) int {
	defer rescue()

	restoreTTY := term.SetupGlobal()
	defer restoreTTY()

	ev, dataDir := InitRuntime(sh.BinPath, sh.SockPath, sh.DbPath)
	defer CleanupRuntime(ev)

	handleSignals(fds[2])

	if len(args) > 0 {
		err := script(ev, args, sh.Cmd, sh.CompileOnly)
		if err != nil {
			if sh.CompileOnly && sh.JSON {
				fmt.Fprintf(fds[1], "%s\n", errorToJSON(err))
			} else {
				diag.PPrintError(err)
			}
			return 2
		}
	} else {
		interact(fds, ev, dataDir, sh.NoRc)
	}

	return 0
}

// Global panic handler.
func rescue() {
	r := recover()
	if r != nil {
		println()
		println(r)
		print(sys.DumpStack())
		println("\nexecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}

func handleSignals(stderr *os.File) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			logger.Println("signal", sig)
			handleSignal(sig, stderr)
		}
	}()
}
