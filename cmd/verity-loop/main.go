package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/verity-bdd/verity-loop/internal/harness"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Fprintf(os.Stderr, "Usage: verity-loop run\n\nRuns the harness loop using verity.yaml in the current directory.\n")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
	}

	os.Exit(harness.Run(ctx, wd))
}
