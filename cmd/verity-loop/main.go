package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/verity-bdd/verity-loop/internal/harness"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Fprintf(os.Stderr, "Usage: verity-loop run [--config <path>]\n\nRuns the harness loop using verity.yaml (default: ./verity.yaml).\n")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", "verity.yaml", "path to verity.yaml")
	fs.Parse(os.Args[2:])

	absConfig, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving config path: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	os.Exit(harness.Run(ctx, absConfig))
}
