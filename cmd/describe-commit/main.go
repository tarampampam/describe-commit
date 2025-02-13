package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gh.tarampamp.am/describe-commit/internal/cli"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error: "+err.Error())

		os.Exit(1)
	}
}

func run() error {
	// create a context that is canceled when the user interrupts the program
	var ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// run the CLI application
	return cli.NewApp().Run(ctx, os.Args)
}
