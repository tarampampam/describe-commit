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
			Description: "This tool leverages AI to generate commit messages based on changes made in a Git repository.",
			Usage:       "[<options>] [<git-dir-path>]",
			Version:     version.Version(),
		},
		opt: newOptionsWithDefaults(),
	}

	var (
		configFile = cmd.Flag[string]{
			Names:   []string{"config-file", "c"},
			Usage:   "Path to the configuration file",
			EnvVars: []string{"CONFIG_FILE"},
			Default: filepath.Join(config.DefaultDirPath(), config.FileName),
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
		openRouterApiKey = cmd.Flag[string]{
			Names:   []string{"openrouter-api-key", "ora"},
			Usage:   "OpenRouter API key (https://bit.ly/4hU1yY1)",
			EnvVars: []string{"OPENROUTER_API_KEY"},
			Default: app.opt.Providers.OpenRouter.ApiKey,
		}
		openRouterModelName = cmd.Flag[string]{
			Names:   []string{"openrouter-model-name", "orm"},
			Usage:   "OpenRouter model name (https://bit.ly/4ktktuG)",
			EnvVars: []string{"OPENROUTER_MODEL_NAME"},
			Default: app.opt.Providers.OpenRouter.ModelName,
		}
		anthropicApiKey = cmd.Flag[string]{
			Names:   []string{"anthropic-api-key", "ana"},
			Usage:   "Anthropic API key (https://bit.ly/4hU1yY1)",
			EnvVars: []string{"ANTHROPIC_API_KEY"},
			Default: app.opt.Providers.Anthropic.ApiKey,
		}
		anthropicModelName = cmd.Flag[string]{
			Names:   []string{"anthropic-model-name", "anm"},
			Usage:   "Anthropic model name (https://bit.ly/4ktktuG)",
			EnvVars: []string{"ANTHROPIC_MODEL_NAME"},
			Default: app.opt.Providers.Anthropic.ModelName,
		}
		anthropicVersion = cmd.Flag[string]{
			Names:   []string{"anthropic-version", "anv"},
			Usage:   "Anthropic version",
			EnvVars: []string{"ANTHROPIC_VERSION"},
			Default: app.opt.Providers.Anthropic.Version,
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
		&openRouterApiKey,
		&openRouterModelName,
		&anthropicApiKey,
		&anthropicModelName,
		&anthropicVersion,
	}

	app.cmd.Action = func(ctx context.Context, c *cmd.Command, args []string) error {
		// determine the working directory
		var wd, wdErr = app.getWorkingDir(args)
		if wdErr != nil {
			return fmt.Errorf("wrong working directory: %w", wdErr)
		}

		// update the options from the configuration file(s)
		if err := app.opt.UpdateFromConfigFile(append([]string{*configFile.Value}, config.FindIn(wd)...)); err != nil {
			return err
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
			setIfFlagIsSet(&app.opt.Providers.OpenRouter.ApiKey, openRouterApiKey)
			setIfFlagIsSet(&app.opt.Providers.OpenRouter.ModelName, openRouterModelName)
			setIfFlagIsSet(&app.opt.Providers.Anthropic.ApiKey, anthropicApiKey)
			setIfFlagIsSet(&app.opt.Providers.Anthropic.ModelName, anthropicModelName)
		}

		if err := app.opt.Validate(); err != nil {
			return fmt.Errorf("invalid options: %w", err)
		}

		return app.run(ctx, wd)
	}

	return &app
}

// setIfFlagIsSet sets the value from the flag to the option if the flag is set and the value is not nil.
func setIfFlagIsSet[T cmd.FlagType](target *T, source cmd.Flag[T]) {
	if target == nil || source.Value == nil || !source.IsSet() {
		return
	}

	*target = *source.Value
}

// getWorkingDir returns the working directory to use for the application.
func (*App) getWorkingDir(args []string) (string, error) {
	var dir string

	if len(args) > 0 {
		dir = filepath.Clean(strings.TrimSpace(args[0]))
	}

	// if the argument was not set, use the `os.Getwd`
	if dir == "" {
		if d, err := os.Getwd(); err != nil {
			return "", err
		} else {
			dir = d
		}
	}

	// convert the relative path to the absolute one
	if !filepath.IsAbs(dir) {
		if abs, absErr := filepath.Abs(dir); absErr != nil {
			return "", absErr
		} else {
			dir = abs
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
	case ai.ProviderOpenRouter:
		provider = ai.NewOpenRouter(
			a.opt.Providers.OpenRouter.ApiKey,
			a.opt.Providers.OpenRouter.ModelName,
		)
	case ai.ProviderAnthropic:
		provider = ai.NewAnthropic(
			a.opt.Providers.Anthropic.ApiKey,
			a.opt.Providers.Anthropic.ModelName,
			a.opt.Providers.Anthropic.Version,
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
