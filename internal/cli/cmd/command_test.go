package cmd_test

import (
	"testing"

	"gh.tarampamp.am/describe-commit/internal/cli/cmd"
)

func TestCommand_Help(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveCommand cmd.Command
		wantHelp    string
	}{
		"empty": {
			giveCommand: cmd.Command{},
			wantHelp:    "",
		},
		"with description": {
			giveCommand: cmd.Command{
				Description: "Some description here",
			},
			wantHelp: "Description:\n   Some description here",
		},
		"with name": {
			giveCommand: cmd.Command{
				Name: "some-name",
			},
			wantHelp: "Usage:\n   some-name",
		},
		"with name and usage": {
			giveCommand: cmd.Command{
				Name:  "some-name",
				Usage: "some-usage",
			},
			wantHelp: "Usage:\n   some-name some-usage",
		},
		"full": {
			giveCommand: cmd.Command{
				Name:        "some-name",
				Description: "Some description here",
				Usage:       "some-usage",
				Version:     "some-version",
				Flags: []cmd.Flagger{
					&cmd.Flag[string]{
						Names:   []string{"config-file", "c"},
						Usage:   "Path to the configuration file",
						EnvVars: []string{"CONFIG_FILE"},
					},
					&cmd.Flag[bool]{
						Names: []string{"help", "h"},
						Usage: "Show help",
					},
				},
			},
			wantHelp: `Description:
   Some description here

Usage:
   some-name some-usage

Version:
   some-version

Options:
   --config-file="…", -c="…"  Path to the configuration file [$CONFIG_FILE]
   --help, -h                 Show help`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotHelp := tc.giveCommand.Help()

			assertEqual(t, tc.wantHelp, gotHelp)
		})
	}
}

func TestCommand_Run(t *testing.T) {
	t.Parallel()
}
