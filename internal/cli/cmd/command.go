package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"unicode/utf8"
)

type Command struct {
	Name    string
	Version string
	Output  io.Writer
	Flags   []Flagger
	Action  func(_ context.Context, _ *Command, args []string) error

	runOnce bool
}

func (c *Command) Help() string {
	const offset = "   "

	var b strings.Builder

	b.Grow(1024) //nolint:mnd // preallocate some memory

	b.WriteString("Description:\n")
	b.WriteString(offset)
	b.WriteString("This tool uses AI to generate a commit message based on the changes made.\n\n")

	if c.Name != "" {
		b.WriteString("Usage:\n")
		b.WriteString(offset)
		b.WriteString(c.Name)
		b.WriteString(" [<options>] [<dir-path>]")
	}

	if c.Version != "" {
		b.WriteString("\n\n")
		b.WriteString("Version:\n")
		b.WriteString(offset)
		b.WriteString(c.Version)
	}

	if len(c.Flags) > 0 {
		b.WriteString("\n\n")
		b.WriteString("Options:\n")

		var (
			longest               int // longest flag names length
			flagNames, flagUsages = make([]string, len(c.Flags)), make([]string, len(c.Flags))
		)

		for i, f := range c.Flags {
			flagNames[i], flagUsages[i] = f.Help()

			if l := utf8.RuneCountInString(flagNames[i]); l > longest {
				longest = l
			}
		}

		for i, flagName := range flagNames {
			if i > 0 {
				b.WriteRune('\n')
			}

			b.WriteString(offset)
			b.WriteString(flagName)

			for j := utf8.RuneCountInString(flagName); j < longest; j++ {
				b.WriteRune(' ')
			}

			b.WriteString("  ")
			b.WriteString(flagUsages[i])
		}
	}

	return b.String()
}

func (c *Command) Run(ctx context.Context, args []string) error { //nolint:funlen
	defer func() { c.runOnce = true }()

	if c.runOnce {
		return fmt.Errorf("command %q has already been run", c.Name)
	}

	var set = flag.NewFlagSet(c.Name, flag.ContinueOnError)

	// mute everything from the standard library
	set.SetOutput(io.Discard)

	// set the default output if not set
	if c.Output == nil {
		c.Output = os.Stdout
	}

	var showHelp, showVersion bool

	// append "built-in" flags
	c.Flags = append(c.Flags, []Flagger{
		&Flag[bool]{
			Names: []string{"help", "h"},
			Usage: "Show help",
			Value: &showHelp,
		},
		&Flag[bool]{
			Names: []string{"version", "v"},
			Usage: "Print the version",
			Value: &showVersion,
		},
	}...)

	// add flags to the set
	for _, f := range c.Flags {
		if err := f.Apply(set); err != nil {
			return err
		}
	}

	// parse the arguments
	if err := set.Parse(args); err != nil {
		if _, outErr := fmt.Fprintf(c.Output, "%s\n\n", c.Help()); outErr != nil {
			err = fmt.Errorf("%w: %w", outErr, err)
		}

		return err
	}

	// if help flag is set then show the help and exit (before flags validation and other actions)
	if showHelp {
		_, err := fmt.Fprintf(c.Output, "%s\n", c.Help())

		return err
	} else if showVersion { // and the same for the version flag
		var (
			runtimeVersion = runtime.Version()
			out            string
		)

		if c.Version != "" {
			out = fmt.Sprintf("%s (%s)\n", c.Version, runtimeVersion)
		} else {
			out = fmt.Sprintf("unknown (%s)\n", runtimeVersion)
		}

		_, err := fmt.Fprint(c.Output, out)

		return err
	}

	// check if the flag has the action and if it's set then run it
	for _, f := range c.Flags {
		if !f.IsSet() {
			continue
		}

		if err := f.Validate(c); err != nil {
			return err
		}

		if err := f.RunAction(c); err != nil {
			return err
		}
	}

	// run the "main" action, if set
	if c.Action != nil {
		return c.Action(ctx, c, set.Args())
	}

	return nil
}
