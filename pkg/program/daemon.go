package program

import (
	"os"

	daemonsvc "github.com/entynetproject/gyrux/pkg/daemon"
	"github.com/entynetproject/gyrux/pkg/program/daemon"
)

type daemonProgram struct{ inner *daemon.Daemon }

func (p daemonProgram) Main(fds [3]*os.File, _ []string) int {
	err := p.inner.Main(daemonsvc.Serve)
	if err != nil {
		logger.Println("daemon error:", err)
		return 2
	}
	return 0
}
