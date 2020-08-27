package main

import (
	"fmt"
	"os"

	"github.com/pivotal/cf-for-k8s-disaster-recovery/backup-metadata/src/internal/delegate"
	"github.com/pivotal/cf-for-k8s-disaster-recovery/backup-metadata/src/internal/helpers"
)

func main() {
	env := helpers.EnvironToMap(os.Environ())

	if err := delegate.Main(os.Args, os.Stdin, env); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
