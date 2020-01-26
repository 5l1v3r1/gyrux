package main

import (
	"os"

	"github.com/entynetproject/gyrux/pkg/program"
)

func main() {
	os.Exit(program.Main([3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args))
}
