package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata/internal/delegate"
	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata/internal/helpers"
)

func main() {
	env := helpers.EnvironToMap(os.Environ())

	if err := delegate.Main(os.Args, os.Stdin, env); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
