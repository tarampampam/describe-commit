package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/diff"
	"gh.tarampamp.am/describe-commit/internal/version"
)

type App struct {
	c *cli.Command

	options struct {
		ShortMessageOnly bool
		EnableEmoji      bool

		Providers struct {
			Gemini struct {
				ApiKey    string
				ModelName string
			}
		}

		IsDebug bool
	}
}

func NewApp() func(context.Context, []string /* args */) error { //nolint:funlen
	var app App

	var (
		shortMessageOnlyFlag = cli.BoolFlag{
			Name:     "short-message-only",
			Aliases:  []string{"s"},
			Usage:    "generate a short commit message (subject line) only",
			Sources:  cli.EnvVars("SHORT_MESSAGE_ONLY"),
			OnlyOnce: true,
		}
		enableEmojiFlag = cli.BoolFlag{
			Name:     "enable-emoji",
			Aliases:  []string{"e"},
			Usage:    "enable emoji in the commit message",
			Sources:  cli.EnvVars("ENABLE_EMOJI"),
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
		debugFlag = cli.BoolFlag{
			Name:     "debug",
			Usage:    "enable debug mode",
			OnlyOnce: true,
		}
	)

	app.c = &cli.Command{
		Usage:     "This tool uses AI to generate a commit message based on the changes made",
		ArgsUsage: "[dir-path]",
		Flags: []cli.Flag{
			&shortMessageOnlyFlag,
			&enableEmojiFlag,
			&geminiApiKeyFlag,
			&geminiModelNameFlag,
			&debugFlag,
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var opt = &app.options

			opt.ShortMessageOnly = c.Bool(shortMessageOnlyFlag.Name)
			opt.EnableEmoji = c.Bool(enableEmojiFlag.Name)
			opt.Providers.Gemini.ApiKey = c.String(geminiApiKeyFlag.Name)
			opt.Providers.Gemini.ModelName = c.String(geminiModelNameFlag.Name)
			opt.IsDebug = c.Bool(debugFlag.Name)

			var workingDir = strings.TrimSpace(c.Args().First())

			if workingDir == "" {
				if cwd, err := os.Getwd(); err != nil {
					return fmt.Errorf("failed to get the current working directory: %w", err)
				} else {
					workingDir = cwd
				}
			}

			if stat, err := os.Stat(workingDir); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("working directory does not exist: %s", workingDir)
				}

				return err
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

	app.Debugf("working directory: %s\n", workingDir)

	if delta, err := diff.Git(workingDir); err != nil {
		return err
	} else {
		changes = delta
	}

	app.Debugf("changes:\n%s\n", changes)

	if changes == "" {
		return fmt.Errorf("no changes found in %s (probably nothing staged; try `git add -A`)", workingDir)
	}

	var provider = ai.NewGemini(
		ctx,
		app.options.Providers.Gemini.ApiKey,
		app.options.Providers.Gemini.ModelName,
	)

	response, respErr := provider.Query(
		ctx,
		changes,
		ai.WithShortMessageOnly(app.options.ShortMessageOnly),
		ai.WithEmoji(app.options.EnableEmoji),
	)
	if respErr != nil {
		return respErr
	}

	app.Debugf("prompt:\n%s\n", response.Prompt)
	app.Debugf("answer:\n%s\n\n", response.Answer)

	if _, err := fmt.Fprintln(os.Stdout, response.Answer); err != nil {
		return err
	}

	return nil
}

func (app *App) Debugf(format string, args ...any) {
	if app.options.IsDebug {
		_, _ = fmt.Fprintf(os.Stderr, "[debug] "+format, args...)
	}
}
