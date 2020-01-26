// Package program provides the entry point to Gyrux. Its subpackages
// correspond to subprograms of Gyrux.
package program

// This package sets up the basic environment and calls the appropriate
// "subprogram", one of the daemon, the terminal interface, or the web
// interface.

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strconv"

	"github.com/entynetproject/gyrux/pkg/program/daemon"
	"github.com/entynetproject/gyrux/pkg/program/shell"
	"github.com/entynetproject/gyrux/pkg/program/web"
	"github.com/entynetproject/gyrux/pkg/util"
)

// defaultPort is the default port on which the web interface runs. The number
// is chosen because it resembles "gyi".
const defaultWebPort = 3171

var logger = util.GetLogger("[main] ")

type flagSet struct {
	flag.FlagSet

	Log, LogPrefix, CPUProfile string

	Help, Version, BuildInfo, JSON bool

	CodeInArg, CompileOnly, NoRc bool

	Web  bool
	Port int

	Daemon bool
	Forked int

	Bin, DB, Sock string
}

func newFlagSet(stderr io.Writer) *flagSet {
	f := flagSet{}
	f.Init("gyrux", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.Usage = func() { usage(stderr, &f) }

	f.StringVar(&f.Log, "log", "", "A file to write debug log to.")
	f.StringVar(&f.LogPrefix, "logprefix", "", "The prefix for the daemon log file.")
	f.StringVar(&f.CPUProfile, "cpuprofile", "", "Write cpu profile to file.")

	f.BoolVar(&f.Help, "help", false, "Show usage help and quit.")
	f.BoolVar(&f.Version, "version", false, "Show version and quit.")
	f.BoolVar(&f.BuildInfo, "buildinfo", false, "Show build info and quit.")
	f.BoolVar(&f.JSON, "json", false, "Show output in JSON. Useful with -buildinfo.")

	f.BoolVar(&f.CodeInArg, "c", false, "Take first argument as code to execute.")
	f.BoolVar(&f.CompileOnly, "compile", false, "Parse/Compile but do not execute.")
	f.BoolVar(&f.NoRc, "norc", false, "Run Gyrux without invoking rc.gy.")

	f.BoolVar(&f.Web, "gui", false, "Run backend of GUI interface.")
	f.IntVar(&f.Port, "port", defaultWebPort, "The port of the GUI backend.")

	f.BoolVar(&f.Daemon, "daemon", false, "Run daemon instead of Gyrux.")

	f.StringVar(&f.Bin, "bin", "", "Path to the Gyrux binary.")
	f.StringVar(&f.DB, "db", "", "Path to the database.")
	f.StringVar(&f.Sock, "sock", "", "Path to the daemon socket.")

	return &f
}

func Main(fds [3]*os.File, args []string) int {
	flag := newFlagSet(fds[2])
	err := flag.Parse(args[1:])
	if err != nil {
		// Error and usage messages are already shown.
		return 2
	}

	// Handle flags common to all subprograms.

	if flag.CPUProfile != "" {
		f, err := os.Create(flag.CPUProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flag.Log != "" {
		err = util.SetOutputFile(flag.Log)
	} else if flag.LogPrefix != "" {
		err = util.SetOutputFile(flag.LogPrefix + strconv.Itoa(os.Getpid()))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return FindProgram(flag).Main(fds, flag.Args())
}

// Program represents a subprogram.
type Program interface {
	// Main runs the subprogram, with given standard files and arguments. The
	// return value will be used as the exit status of the entire program.
	Main(fds [3]*os.File, args []string) int
}

// FindProgram finds a suitable Program according to flags. It does not have any
// side effects.
func FindProgram(flag *flagSet) Program {
	switch {
	case flag.Help:
		return helpProgram{flag}
	case flag.Version:
		return versionProgram{}
	case flag.BuildInfo:
		return buildInfoProgram{flag.JSON}
	case flag.Daemon:
		if len(flag.Args()) > 0 {
			return badUsageProgram{"arguments are not allowed with -daemon", flag}
		}
		return daemonProgram{&daemon.Daemon{
			BinPath:       flag.Bin,
			DbPath:        flag.DB,
			SockPath:      flag.Sock,
			LogPathPrefix: flag.LogPrefix,
		}}
	case flag.Web:
		if len(flag.Args()) > 0 {
			return badUsageProgram{"arguments are not allowed with -gui", flag}
		}
		if flag.CodeInArg {
			return badUsageProgram{"-c cannot be used together with -gui", flag}
		}
		return &web.Web{
			BinPath: flag.Bin, SockPath: flag.Sock, DbPath: flag.DB,
			Port: flag.Port}
	default:
		return &shell.Shell{
			BinPath: flag.Bin, SockPath: flag.Sock, DbPath: flag.DB,
			Cmd: flag.CodeInArg, CompileOnly: flag.CompileOnly,
			NoRc: flag.NoRc, JSON: flag.JSON}
	}
}
