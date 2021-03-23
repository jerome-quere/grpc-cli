package main

import (
	"context"
	"os"

	"github.com/jerome-quere/grpc-cli/internal/core"
)

func main() {

	exitCode := core.Bootstrap(
		context.Background(),
		&core.BootstrapConfig{
			Stderr: os.Stderr,
			Stdout: os.Stdout,
			Stdin:  os.Stdin,
			Args:   os.Args,
		})
	os.Exit(exitCode)
}
