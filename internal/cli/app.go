package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/cli/app"
	"gh.tarampamp.am/describe-commit/internal/config"
	"gh.tarampamp.am/describe-commit/internal/git"
	"gh.tarampamp.am/describe-commit/internal/version"
)

type App struct {
	c   app.Command
	opt options
}

func NewApp2(name string) *App {
	var a = App{c: app.Command{Name: name}}

	a.c.Flags = []app.Flagger{
		&app.Flag[string]{
			Names:   []string{"config-file", "c"},
			Usage:   "Path to the configuration file",
			EnvVars: []string{"CONFIG_FILE"},
			Default: func() string { // default file path depends on the user's OS
				const fileName = "describe-commit.yml"

				if dir, err := os.UserConfigDir(); err == nil {
					return filepath.Join(dir, fileName)
				}

				if cwd, err := os.Getwd(); err == nil {
					return filepath.Join(cwd, fileName)
				}

				return filepath.Join(".", fileName)
			}(),
			Value: &a.opt.ConfigFilePath.Flag,
		},
		&app.Flag[bool]{
			Names:   []string{"short-message-only", "s"},
			Usage:   "Generate a short commit message (subject line) only",
			EnvVars: []string{"SHORT_MESSAGE_ONLY"},
			Value:   &a.opt.ShortMessageOnly.Flag,
		},
		&app.Flag[int64]{
			Names:   []string{"commit-history-length", "cl", "hl"},
			Usage:   "Number of previous commits from the Git history (0 = disabled)",
			EnvVars: []string{"COMMIT_HISTORY_LENGTH"},
			Default: 20, //nolint:mnd
			Value:   &a.opt.CommitHistoryLength.Flag,
		},
		&app.Flag[bool]{
			Names:   []string{"enable-emoji", "e"},
			Usage:   "Enable emoji in the commit message",
			EnvVars: []string{"ENABLE_EMOJI"},
			Value:   &a.opt.EnableEmoji.Flag,
		},
		&app.Flag[int64]{
			Names:   []string{"max-output-tokens"},
			Usage:   "Maximum number of tokens in the output message",
			EnvVars: []string{"MAX_OUTPUT_TOKENS"},
			Default: 500, //nolint:mnd
			Value:   &a.opt.MaxOutputTokens.Flag,
		},
		&app.Flag[string]{
			Names:   []string{"ai-provider", "ai"},
			Usage:   fmt.Sprintf("AI provider name (%s)", strings.Join(ai.SupportedProviders(), "|")),
			EnvVars: []string{"AI_PROVIDER"},
			Default: ai.ProviderGemini, // due to its free
			Validator: func(_ *app.Command, s string) error {
				if !ai.IsProviderSupported(s) {
					return fmt.Errorf("unsupported AI provider: %s", s)
				}

				return nil
			},
			Value: &a.opt.AIProviderName.Flag,
		},
		&app.Flag[string]{
			Names:   []string{"gemini-api-key", "ga"},
			Usage:   "Gemini API key (https://bit.ly/4jZhiKI, as of February 2025 it's free)",
			EnvVars: []string{"GEMINI_API_KEY"},
			Value:   &a.opt.Providers.Gemini.ApiKey.Flag,
		},
		&app.Flag[string]{
			Names:   []string{"gemini-model-name", "gm"},
			Usage:   "Gemini model name (https://bit.ly/4i02ARR)",
			EnvVars: []string{"GEMINI_MODEL_NAME"},
			Default: "gemini-2.0-flash",
			Value:   &a.opt.Providers.Gemini.ApiKey.Flag,
		},
		&app.Flag[string]{
			Names:   []string{"openai-api-key", "oa"},
			Usage:   "OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account)",
			EnvVars: []string{"OPENAI_API_KEY"},
			Value:   &a.opt.Providers.OpenAI.ApiKey.Flag,
		},
		&app.Flag[string]{
			Names:   []string{"openai-model-name", "om"},
			Usage:   "OpenAI model name (https://bit.ly/4hXCXkL)",
			EnvVars: []string{"OPENAI_MODEL_NAME"},
			Default: "gpt-4o-mini",
			Value:   &a.opt.Providers.Gemini.ModelName.Flag,
		},
	}

	a.c.Action = func(ctx context.Context, c *app.Command) error {
		var cfg config.Config

		if err := cfg.FromFile(a.opt.ConfigFilePath.Flag); err == nil {
			a.opt.FromConfig(cfg)
		} else {
			debugf("failed to load the configuration file: %s", err)
		}

		if err := a.opt.Validate(); err != nil {
			return err
		}

		fmt.Printf("%+v\n", a.opt)

		return nil
	}

	return &a
}

func (a *App) Run(ctx context.Context, args []string) error { return a.c.Run(ctx, args) }

//go:generate go run app_readme.go

type cliApp struct {
	c *cli.Command

	options struct {
		ShortMessageOnly    option[bool]
		CommitHistoryLength option[int64]
		EnableEmoji         option[bool]
		MaxOutputTokens     option[int64]

		AIProviderName option[string]

		Providers struct {
			Gemini struct {
				ApiKey    option[string]
				ModelName option[string]
			}

			OpenAI struct {
				ApiKey    option[string]
				ModelName option[string]
			}
		}
	}
}

// NewApp creates new console application.
func NewApp() *cli.Command { //nolint:funlen,gocognit
	var app cliApp

	app.c = &cli.Command{
		Usage:       "This tool uses AI to generate a commit message based on the changes made",
		Description: fmt.Sprintf("To debug the application, set the %s environment variable to `true`", debugEnvName),
		ArgsUsage:   "[dir-path]",
		Flags: []cli.Flag{
			&configFilePathFlag,
			&shortMessageOnlyFlag,
			&commitHistoryLengthFlag,
			&enableEmojiFlag,
			&maxOutputTokensFlag,
			&aiProviderNameFlag,
			&geminiApiKeyFlag,
			&geminiModelNameFlag,
			&openAIApiKeyFlag,
			&openAIModelNameFlag,
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			{ // initialize the options
				var opt = &app.options

				{ // first, try to load the configuration file to initialize the options as defaults
					var cfg = new(config.Config)

					if err := cfg.FromFile(c.String(configFilePathFlag.Name)); err == nil {
						opt.ShortMessageOnly.SetIfNotNil(cfg.ShortMessageOnly)
						opt.CommitHistoryLength.SetIfNotNil(cfg.CommitHistoryLength)
						opt.EnableEmoji.SetIfNotNil(cfg.EnableEmoji)
						opt.MaxOutputTokens.SetIfNotNil(cfg.MaxOutputTokens)
						opt.AIProviderName.SetIfNotNil(cfg.AIProviderName)

						if sub := cfg.Gemini; sub != nil {
							opt.Providers.Gemini.ApiKey.SetIfNotNil(sub.ApiKey)
							opt.Providers.Gemini.ModelName.SetIfNotNil(sub.ModelName)
						}

						if sub := cfg.OpenAI; sub != nil {
							opt.Providers.OpenAI.ApiKey.SetIfNotNil(sub.ApiKey)
							opt.Providers.OpenAI.ModelName.SetIfNotNil(sub.ModelName)
						}
					} else {
						debugf("failed to load the configuration file: %s", err)
					}
				}

				{ // next, override the options with the command-line flags or use their default values if they are not provided
					opt.ShortMessageOnly.SetFromFlagIfUnset(c, shortMessageOnlyFlag.Name, c.Bool)
					opt.CommitHistoryLength.SetFromFlagIfUnset(c, commitHistoryLengthFlag.Name, c.Int)
					opt.EnableEmoji.SetFromFlagIfUnset(c, enableEmojiFlag.Name, c.Bool)
					opt.MaxOutputTokens.SetFromFlagIfUnset(c, maxOutputTokensFlag.Name, c.Int)
					opt.AIProviderName.SetFromFlagIfUnset(c, aiProviderNameFlag.Name, c.String)
					opt.Providers.Gemini.ApiKey.SetFromFlagIfUnset(c, geminiApiKeyFlag.Name, c.String)
					opt.Providers.Gemini.ModelName.SetFromFlagIfUnset(c, geminiModelNameFlag.Name, c.String)
					opt.Providers.OpenAI.ApiKey.SetFromFlagIfUnset(c, openAIApiKeyFlag.Name, c.String)
					opt.Providers.OpenAI.ModelName.SetFromFlagIfUnset(c, openAIModelNameFlag.Name, c.String)
				}

				{ // validate the options
					if opt.MaxOutputTokens.Value <= 1 {
						return errors.New("max output tokens must be greater than 1")
					}

					if !ai.IsProviderSupported(opt.AIProviderName.Value) {
						return fmt.Errorf("unsupported AI provider: %s", opt.AIProviderName.Value)
					}

					if opt.AIProviderName.Value == ai.ProviderGemini {
						if opt.Providers.Gemini.ApiKey.Value == "" {
							return errors.New("gemini API key is required")
						}

						if opt.Providers.Gemini.ModelName.Value == "" {
							return errors.New("gemini model name is required")
						}
					}

					if opt.AIProviderName.Value == ai.ProviderOpenAI {
						if opt.Providers.OpenAI.ApiKey.Value == "" {
							return errors.New("OpenAI API key is required")
						}

						if opt.Providers.OpenAI.ModelName.Value == "" {
							return errors.New("OpenAI model name is required")
						}
					}
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

func (app *cliApp) Run(ctx context.Context, workingDir string) error { //nolint:funlen
	debugf("AI provider: %s", app.options.AIProviderName.Value)

	var provider ai.Provider

	switch app.options.AIProviderName.Value {
	case ai.ProviderGemini:
		provider = ai.NewGemini(
			app.options.Providers.Gemini.ApiKey.Value,
			app.options.Providers.Gemini.ModelName.Value,
		)
	case ai.ProviderOpenAI:
		provider = ai.NewOpenAI(
			app.options.Providers.OpenAI.ApiKey.Value,
			app.options.Providers.OpenAI.ModelName.Value,
		)
	default:
		return fmt.Errorf("unsupported AI provider: %s", app.options.AIProviderName.Value)
	}

	debugf("working directory: %s", workingDir)

	var (
		eg, egCtx        = errgroup.WithContext(ctx)
		changes, commits string
	)

	eg.Go(func() (err error) { changes, err = git.Diff(egCtx, workingDir); return }) //nolint:nlreturn

	if histLen := int(app.options.CommitHistoryLength.Value); histLen > 0 {
		eg.Go(func() (err error) { commits, err = git.Log(egCtx, workingDir, histLen); return }) //nolint:nlreturn
	} else {
		commits = "NO COMMITS"
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	debugf("changes:\n%s", changes)
	debugf("commits:\n%s", commits)

	if changes == "" {
		return fmt.Errorf("no changes found in %s (probably nothing staged; try `git add -A`)", workingDir)
	}

	response, respErr := provider.Query(
		ctx,
		changes,
		commits,
		ai.WithShortMessageOnly(app.options.ShortMessageOnly.Value),
		ai.WithEmoji(app.options.EnableEmoji.Value),
		ai.WithMaxOutputTokens(app.options.MaxOutputTokens.Value),
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
