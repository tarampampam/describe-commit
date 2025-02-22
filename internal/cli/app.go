package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/cli/cmd"
	"gh.tarampamp.am/describe-commit/internal/config"
	"gh.tarampamp.am/describe-commit/internal/debug"
	"gh.tarampamp.am/describe-commit/internal/errgroup"
	"gh.tarampamp.am/describe-commit/internal/git"
	"gh.tarampamp.am/describe-commit/internal/version"
)

//go:generate go run ./generate/readme.go

type App struct {
	cmd cmd.Command
	opt options
}

func NewApp(name string) *App { //nolint:funlen
	var app = App{
		cmd: cmd.Command{
			Name:        name,
			Description: "This tool uses AI to generate a commit message based on the changes made.",
			Usage:       "[<options>] [<dir-path>]",
			Version:     version.Version(),
		},
		opt: newOptionsWithDefaults(),
	}

	var (
		configFile = cmd.Flag[string]{
			Names:   []string{"config-file", "c"},
			Usage:   "Path to the configuration file",
			EnvVars: []string{"CONFIG_FILE"},
			Default: DefaultConfigFilePath,
		}
		shortMessageOnly = cmd.Flag[bool]{
			Names:   []string{"short-message-only", "s"},
			Usage:   "Generate a short commit message (subject line) only",
			EnvVars: []string{"SHORT_MESSAGE_ONLY"},
			Default: app.opt.ShortMessageOnly,
		}
		commitHistoryLength = cmd.Flag[int64]{
			Names:   []string{"commit-history-length", "cl", "hl"},
			Usage:   "Number of previous commits from the Git history (0 = disabled)",
			EnvVars: []string{"COMMIT_HISTORY_LENGTH"},
			Default: app.opt.CommitHistoryLength,
		}
		enableEmoji = cmd.Flag[bool]{
			Names:   []string{"enable-emoji", "e"},
			Usage:   "Enable emoji in the commit message",
			EnvVars: []string{"ENABLE_EMOJI"},
			Default: app.opt.EnableEmoji,
		}
		maxOutputTokens = cmd.Flag[int64]{
			Names:   []string{"max-output-tokens"},
			Usage:   "Maximum number of tokens in the output message",
			EnvVars: []string{"MAX_OUTPUT_TOKENS"},
			Validator: func(_ *cmd.Command, i int64) error {
				if i <= 1 {
					return errors.New("max output tokens must be greater than 1")
				}

				return nil
			},
			Default: app.opt.MaxOutputTokens,
		}
		aiProviderName = cmd.Flag[string]{
			Names:   []string{"ai-provider", "ai"},
			Usage:   fmt.Sprintf("AI provider name (%s)", strings.Join(ai.SupportedProviders(), "|")),
			EnvVars: []string{"AI_PROVIDER"},
			Default: app.opt.AIProviderName,
			Validator: func(_ *cmd.Command, s string) error {
				if !ai.IsProviderSupported(s) {
					return fmt.Errorf("unsupported AI provider: %s", s)
				}

				return nil
			},
		}
		geminiApiKey = cmd.Flag[string]{
			Names:   []string{"gemini-api-key", "ga"},
			Usage:   "Gemini API key (https://bit.ly/4jZhiKI, as of February 2025 it's free)",
			EnvVars: []string{"GEMINI_API_KEY"},
			Default: app.opt.Providers.Gemini.ApiKey,
		}
		geminiModelName = cmd.Flag[string]{
			Names:   []string{"gemini-model-name", "gm"},
			Usage:   "Gemini model name (https://bit.ly/4i02ARR)",
			EnvVars: []string{"GEMINI_MODEL_NAME"},
			Default: app.opt.Providers.Gemini.ModelName,
		}
		openAIApiKey = cmd.Flag[string]{
			Names:   []string{"openai-api-key", "oa"},
			Usage:   "OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account)",
			EnvVars: []string{"OPENAI_API_KEY"},
			Default: app.opt.Providers.OpenAI.ApiKey,
		}
		openAIModelName = cmd.Flag[string]{
			Names:   []string{"openai-model-name", "om"},
			Usage:   "OpenAI model name (https://bit.ly/4hXCXkL)",
			EnvVars: []string{"OPENAI_MODEL_NAME"},
			Default: app.opt.Providers.OpenAI.ModelName,
		}
	)

	app.cmd.Flags = []cmd.Flagger{
		&configFile,
		&shortMessageOnly,
		&commitHistoryLength,
		&enableEmoji,
		&maxOutputTokens,
		&aiProviderName,
		&geminiApiKey,
		&geminiModelName,
		&openAIApiKey,
		&openAIModelName,
	}

	app.cmd.Action = func(ctx context.Context, c *cmd.Command, args []string) error {
		var wd, wdErr = app.getWorkingDir(args)
		if wdErr != nil {
			return fmt.Errorf("wrong working directory: %w", wdErr)
		}

		var cfgFilePath string

		if configFile.Value != nil {
			cfgFilePath = *configFile.Value
		} else if path, found := app.getConfigFilePath(wd); found {
			cfgFilePath = path
		}

		if cfgFilePath != "" {
			var cfg config.Config

			// override the default options with the configuration file values, if they are set
			if err := cfg.FromFile(cfgFilePath); err == nil {
				setIfSourceNotNil(&app.opt.ShortMessageOnly, cfg.ShortMessageOnly)
				setIfSourceNotNil(&app.opt.CommitHistoryLength, cfg.CommitHistoryLength)
				setIfSourceNotNil(&app.opt.EnableEmoji, cfg.EnableEmoji)
				setIfSourceNotNil(&app.opt.MaxOutputTokens, cfg.MaxOutputTokens)
				setIfSourceNotNil(&app.opt.AIProviderName, cfg.AIProviderName)

				if sub := cfg.Gemini; sub != nil {
					setIfSourceNotNil(&app.opt.Providers.Gemini.ApiKey, sub.ApiKey)
					setIfSourceNotNil(&app.opt.Providers.Gemini.ModelName, sub.ModelName)
				}

				if sub := cfg.OpenAI; sub != nil {
					setIfSourceNotNil(&app.opt.Providers.OpenAI.ApiKey, sub.ApiKey)
					setIfSourceNotNil(&app.opt.Providers.OpenAI.ModelName, sub.ModelName)
				}
			} else {
				debug.Printf("failed to load the configuration file: %s", err)
			}
		} else {
			debug.Printf("configuration file not found")
		}

		{ // override the options with the command-line flags
			setIfFlagIsSet(&app.opt.ShortMessageOnly, shortMessageOnly)
			setIfFlagIsSet(&app.opt.CommitHistoryLength, commitHistoryLength)
			setIfFlagIsSet(&app.opt.EnableEmoji, enableEmoji)
			setIfFlagIsSet(&app.opt.MaxOutputTokens, maxOutputTokens)
			setIfFlagIsSet(&app.opt.AIProviderName, aiProviderName)
			setIfFlagIsSet(&app.opt.Providers.Gemini.ApiKey, geminiApiKey)
			setIfFlagIsSet(&app.opt.Providers.Gemini.ModelName, geminiModelName)
			setIfFlagIsSet(&app.opt.Providers.OpenAI.ApiKey, openAIApiKey)
			setIfFlagIsSet(&app.opt.Providers.OpenAI.ModelName, openAIModelName)
		}

		if err := app.opt.Validate(); err != nil {
			return fmt.Errorf("invalid options: %w", err)
		}

		return app.run(ctx, wd)
	}

	return &app
}

// getWorkingDir returns the working directory to use for the application.
func (*App) getWorkingDir(args []string) (string, error) {
	var dir string

	if len(args) > 0 {
		dir = strings.TrimSpace(args[0])
	}

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

	return filepath.Clean(dir), nil
}

// Run runs the application.
func (a *App) Run(ctx context.Context, args []string) error { return a.cmd.Run(ctx, args) }

// Help returns the help message.
func (a *App) Help() string { return a.cmd.Help() }

// run in the main logic of the application.
func (a *App) run(ctx context.Context, workingDir string) error { //nolint:funlen
	debug.Printf("AI provider: %s", a.opt.AIProviderName)

	var provider ai.Provider

	switch a.opt.AIProviderName {
	case ai.ProviderGemini:
		provider = ai.NewGemini(
			a.opt.Providers.Gemini.ApiKey,
			a.opt.Providers.Gemini.ModelName,
		)
	case ai.ProviderOpenAI:
		provider = ai.NewOpenAI(
			a.opt.Providers.OpenAI.ApiKey,
			a.opt.Providers.OpenAI.ModelName,
		)
	default:
		return fmt.Errorf("unsupported AI provider: %s", a.opt.AIProviderName)
	}

	debug.Printf("working directory: %s", workingDir)

	var (
		eg, _            = errgroup.New(ctx)
		changes, commits string
	)

	eg.Go(func(ctx context.Context) (err error) {
		changes, err = git.Diff(ctx, workingDir)

		return
	})

	if histLen := int(a.opt.CommitHistoryLength); histLen > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			commits, err = git.Log(ctx, workingDir, histLen)

			return
		})
	} else {
		commits = "NO COMMITS"
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	debug.Printf("changes:\n%s", changes)
	debug.Printf("commits:\n%s", commits)

	if changes == "" {
		return fmt.Errorf("no changes found in %s (probably nothing staged; try `git add -A`)", workingDir)
	}

	response, respErr := provider.Query(
		ctx,
		changes,
		commits,
		ai.WithShortMessageOnly(a.opt.ShortMessageOnly),
		ai.WithEmoji(a.opt.EnableEmoji),
		ai.WithMaxOutputTokens(a.opt.MaxOutputTokens),
	)
	if respErr != nil {
		return respErr
	}

	debug.Printf("prompt:\n%s", response.Prompt)
	debug.Printf("answer:\n%s\n", response.Answer)

	if _, err := fmt.Fprintln(os.Stdout, response.Answer); err != nil {
		return err
	}

	return nil
}

func setIfSourceNotNil[T any](target, source *T) {
	if target == nil || source == nil {
		return
	}

	*target = *source
}

func setIfFlagIsSet[T cmd.FlagType](target *T, source cmd.Flag[T]) {
	if target == nil || source.Value == nil || !source.IsSet() {
		return
	}

	*target = *source.Value
}
