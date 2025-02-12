package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/describe-commit/internal/diff"
	"gh.tarampamp.am/describe-commit/internal/providers"
	"gh.tarampamp.am/describe-commit/internal/version"
)

type App struct {
	c *cli.Command

	options struct {
		ShortMessageOnly bool

		Providers struct {
			Gemini struct {
				ApiKey    string
				ModelName string
			}
		}
	}
}

func NewApp() func(context.Context, []string /* args */) error {
	var app App

	var (
		shortMessageOnlyFlag = cli.BoolFlag{
			Name:     "short-message-only",
			Aliases:  []string{"s"},
			Usage:    "generate a short commit message (subject line) only",
			Sources:  cli.EnvVars("SHORT_MESSAGE_ONLY"),
			OnlyOnce: true,
		}
		geminiApiKeyFlag = cli.StringFlag{
			Name:     "gemini-api-key",
			Aliases:  []string{"ga"},
			Usage:    "Gemini API key",
			Sources:  cli.EnvVars("GEMINI_API_KEY"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
		}
		geminiModelNameFlag = cli.StringFlag{
			Name:     "gemini-model-name",
			Aliases:  []string{"gm"},
			Usage:    "Gemini model name",
			Sources:  cli.EnvVars("GEMINI_MODEL_NAME"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Value:    "gemini-2.0-flash",
		}
	)

	app.c = &cli.Command{
		Usage:     "describe commit",
		ArgsUsage: "[dir-path]",
		Flags: []cli.Flag{ // global flags
			&geminiApiKeyFlag,
			&geminiModelNameFlag,
			&shortMessageOnlyFlag,
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var opt = &app.options

			opt.ShortMessageOnly = c.Bool(shortMessageOnlyFlag.Name)
			opt.Providers.Gemini.ApiKey = c.String(geminiApiKeyFlag.Name)
			opt.Providers.Gemini.ModelName = c.String(geminiModelNameFlag.Name)

			var workingDir = c.Args().First()

			if workingDir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getwd failed: %w", err)
				}

				workingDir = cwd
			}

			if stat, err := os.Stat(workingDir); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("not found: %s", workingDir)
				}

				return fmt.Errorf("stat failed: %w", err)
			} else if !stat.IsDir() {
				return fmt.Errorf("not a directory: %s", workingDir)
			}

			return app.Run(ctx, workingDir)
		},
		Version: fmt.Sprintf("%s (%s)", version.Version(), runtime.Version()),
	}

	return app.c.Run
}

func (app *App) Run(ctx context.Context, workingDir string) error {
	var changes string

	if delta, err := diff.Git(workingDir); err != nil {
		return err
	} else {
		changes = delta
	}

	if changes == "" {
		return fmt.Errorf("no changes found in %s (probably nothing staged; try `git add .`)", workingDir)
	}

	var provider = providers.NewGemini(
		ctx,
		app.options.Providers.Gemini.ApiKey,
		app.options.Providers.Gemini.ModelName,
	)

	response, respErr := provider.Query(ctx, changes, providers.WithShortMessageOnly(app.options.ShortMessageOnly))
	if respErr != nil {
		return respErr
	}

	if _, err := fmt.Fprintln(os.Stdout, response); err != nil {
		return err
	}

	return nil
}
