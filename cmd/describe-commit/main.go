package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"gh.tarampamp.am/describe-commit/internal/cli"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error: "+err.Error())

		os.Exit(1)
	}
}

func run() error {
	const dotenvFileName = ".env" // dotenv (.env) file name

	// load .env file (if file exists; useful for the local app development)
	if stat, dotenvErr := os.Stat(dotenvFileName); dotenvErr == nil && stat.Mode().IsRegular() {
		if err := godotenv.Load(dotenvFileName); err != nil {
			return err
		}
	}

	// create a context that is canceled when the user interrupts the program
	var ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// run the CLI application
	return cli.NewApp()(ctx, os.Args)
}
