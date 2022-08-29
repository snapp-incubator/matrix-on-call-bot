package main

import (
	"os"

	_ "go.uber.org/automaxprocs"

	"github.com/snapp-incubator/matrix-on-call-bot/internal/cmd"
)

const (
	exitFailure = 1
)

func main() {
	root := cmd.NewRootCommand()

	if root != nil {
		if err := root.Execute(); err != nil {
			os.Exit(exitFailure)
		}
	}
}
