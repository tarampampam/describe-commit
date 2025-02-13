package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/config"
	"gh.tarampamp.am/describe-commit/internal/diff"
	"gh.tarampamp.am/describe-commit/internal/version"
)

//go:generate go run app_readme.go

type cliApp struct {
	c *cli.Command

	options struct {
		ShortMessageOnly option[bool]
		EnableEmoji      option[bool]

		Providers struct {
			Gemini struct {
				ApiKey    option[string]
				ModelName option[string]
			}
		}
	}
}

// NewApp creates new console application.
func NewApp() *cli.Command {
	var app cliApp

	app.c = &cli.Command{
		Usage:       "This tool uses AI to generate a commit message based on the changes made",
		Description: fmt.Sprintf("To debug the application, set the %s environment variable to `true`", debugEnvName),
		ArgsUsage:   "[dir-path]",
		Flags: []cli.Flag{
			&configFilePathFlag,
			&shortMessageOnlyFlag,
			&enableEmojiFlag,
			&geminiApiKeyFlag,
			&geminiModelNameFlag,
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			{ // initialize the options
				var opt = &app.options

				{ // first, try to load the configuration file to initialize the options as defaults
					var cfg = new(config.Config)

					if err := cfg.FromFile(c.String(configFilePathFlag.Name)); err == nil {
						opt.ShortMessageOnly.SetIfNotNil(cfg.ShortMessageOnly)
						opt.EnableEmoji.SetIfNotNil(cfg.EnableEmoji)

						if sub := cfg.Gemini; sub != nil {
							opt.Providers.Gemini.ApiKey.SetIfNotNil(sub.ApiKey)
							opt.Providers.Gemini.ModelName.SetIfNotNil(sub.ModelName)
						}
					} else {
						debugf("failed to load the configuration file: %s", err)
					}
				}

				{ // next, override the options with the command-line flags or use their default values if they are not provided
					opt.ShortMessageOnly.SetFromFlagIfUnset(c, shortMessageOnlyFlag.Name, c.Bool)
					opt.EnableEmoji.SetFromFlagIfUnset(c, enableEmojiFlag.Name, c.Bool)
					opt.Providers.Gemini.ApiKey.SetFromFlagIfUnset(c, geminiApiKeyFlag.Name, c.String)
					opt.Providers.Gemini.ModelName.SetFromFlagIfUnset(c, geminiModelNameFlag.Name, c.String)
				}
			}

			var wd, wdErr = app.getWorkingDir(c)
			if wdErr != nil {
				return fmt.Errorf("wrong working directory: %w", wdErr)
			}

			return app.Run(ctx, wd)
		},
		Version: fmt.Sprintf("%s (%s)", version.Version(), runtime.Version()),
	}

	return app.c
}

func (app *cliApp) Run(ctx context.Context, workingDir string) error {
	var changes string

	debugf("working directory: %s", workingDir)

	if delta, err := diff.Git(workingDir); err != nil {
		return err
	} else {
		changes = delta
	}

	debugf("changes:\n%s", changes)

	if changes == "" {
		return fmt.Errorf("no changes found in %s (probably nothing staged; try `git add -A`)", workingDir)
	}

	var provider = ai.NewGemini(
		ctx,
		app.options.Providers.Gemini.ApiKey.Value,
		app.options.Providers.Gemini.ModelName.Value,
	)

	response, respErr := provider.Query(
		ctx,
		changes,
		ai.WithShortMessageOnly(app.options.ShortMessageOnly.Value),
		ai.WithEmoji(app.options.EnableEmoji.Value),
	)
	if respErr != nil {
		return respErr
	}

	debugf("prompt:\n%s", response.Prompt)
	debugf("answer:\n%s\n", response.Answer)

	if _, err := fmt.Fprintln(os.Stdout, response.Answer); err != nil {
		return err
	}

	return nil
}

// getWorkingDir returns the working directory to use for the application.
func (app *cliApp) getWorkingDir(c *cli.Command) (string, error) {
	// get the working directory from the first command-line argument
	var dir = strings.TrimSpace(c.Args().First())

	// if the argument was not set, use the `os.Getwd`
	if dir == "" {
		if d, err := os.Getwd(); err != nil {
			return "", err
		} else {
			dir = d
		}
	}

	// check the working directory existence
	if stat, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("working directory does not exist: %s", dir)
		}

		return "", err
	} else if !stat.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dir)
	}

	return dir, nil
}
