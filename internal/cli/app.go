package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/cli/cmd"
	"gh.tarampamp.am/describe-commit/internal/config"
	"gh.tarampamp.am/describe-commit/internal/debug"
	"gh.tarampamp.am/describe-commit/internal/errgroup"
	"gh.tarampamp.am/describe-commit/internal/git"
	"gh.tarampamp.am/describe-commit/internal/retry"
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
		retryAttempts = cmd.Flag[uint]{
			Names:   []string{"retry-attempts"},
			Usage:   "Maximum number of retry attempts on retryable API errors (0 = unlimited retries)",
			EnvVars: []string{"RETRY_ATTEMPTS"},
			Default: app.opt.MaxRetries,
		}
		retryDelay = cmd.Flag[time.Duration]{
			Names:   []string{"retry-delay"},
			Usage:   "Delay between retry attempts (e.g. 1s, 500ms)",
			EnvVars: []string{"RETRY_DELAY"},
			Default: app.opt.RetryDelay,
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
			Usage:   "Gemini API key (https://aistudio.google.com/app/api-keys, as of February 2025 it's free)",
			EnvVars: []string{"GEMINI_API_KEY"},
			Default: app.opt.Providers.Gemini.ApiKey,
		}
		geminiModelName = cmd.Flag[string]{
			Names:   []string{"gemini-model-name", "gm"},
			Usage:   "Gemini model name (https://ai.google.dev/gemini-api/docs/models)",
			EnvVars: []string{"GEMINI_MODEL_NAME"},
			Default: app.opt.Providers.Gemini.ModelName,
		}
		geminiBaseURL = cmd.Flag[string]{
			Names:   []string{"gemini-base-url"},
			Usage:   "Gemini API base URL (overrides the default endpoint)",
			EnvVars: []string{"GEMINI_BASE_URL"},
			Default: app.opt.Providers.Gemini.BaseURL,
		}
		openAIApiKey = cmd.Flag[string]{
			Names:   []string{"openai-api-key", "oa"},
			Usage:   "OpenAI API key (https://platform.openai.com/api-keys, you need to add funds to your account)",
			EnvVars: []string{"OPENAI_API_KEY"},
			Default: app.opt.Providers.OpenAI.ApiKey,
		}
		openAIModelName = cmd.Flag[string]{
			Names:   []string{"openai-model-name", "om"},
			Usage:   "OpenAI model name (https://developers.openai.com/api/docs/models)",
			EnvVars: []string{"OPENAI_MODEL_NAME"},
			Default: app.opt.Providers.OpenAI.ModelName,
		}
		openAIBaseURL = cmd.Flag[string]{
			Names:   []string{"openai-base-url"},
			Usage:   "OpenAI API base URL (use to connect to OpenAI-compatible endpoints, e.g. Ollama)",
			EnvVars: []string{"OPENAI_BASE_URL"},
			Default: app.opt.Providers.OpenAI.BaseURL,
		}
		openRouterApiKey = cmd.Flag[string]{
			Names:   []string{"openrouter-api-key", "ora"},
			Usage:   "OpenRouter API key (https://openrouter.ai/workspaces/default/keys)",
			EnvVars: []string{"OPENROUTER_API_KEY"},
			Default: app.opt.Providers.OpenRouter.ApiKey,
		}
		openRouterModelName = cmd.Flag[string]{
			Names:   []string{"openrouter-model-name", "orm"},
			Usage:   "OpenRouter model name (https://openrouter.ai/models)",
			EnvVars: []string{"OPENROUTER_MODEL_NAME"},
			Default: app.opt.Providers.OpenRouter.ModelName,
		}
		openRouterBaseURL = cmd.Flag[string]{
			Names:   []string{"openrouter-base-url"},
			Usage:   "OpenRouter API base URL (overrides the default endpoint)",
			EnvVars: []string{"OPENROUTER_BASE_URL"},
			Default: app.opt.Providers.OpenRouter.BaseURL,
		}
		anthropicApiKey = cmd.Flag[string]{
			Names:   []string{"anthropic-api-key", "ana"},
			Usage:   "Anthropic API key (https://platform.claude.com/settings/keys)",
			EnvVars: []string{"ANTHROPIC_API_KEY"},
			Default: app.opt.Providers.Anthropic.ApiKey,
		}
		anthropicModelName = cmd.Flag[string]{
			Names:   []string{"anthropic-model-name", "anm"},
			Usage:   "Anthropic model name (https://platform.claude.com/docs/en/about-claude/models/overview)",
			EnvVars: []string{"ANTHROPIC_MODEL_NAME"},
			Default: app.opt.Providers.Anthropic.ModelName,
		}
		anthropicBaseURL = cmd.Flag[string]{
			Names:   []string{"anthropic-base-url"},
			Usage:   "Anthropic API base URL (overrides the default endpoint)",
			EnvVars: []string{"ANTHROPIC_BASE_URL"},
			Default: app.opt.Providers.Anthropic.BaseURL,
		}
	)

	app.cmd.Flags = []cmd.Flagger{
		&configFile,
		&shortMessageOnly,
		&commitHistoryLength,
		&enableEmoji,
		&maxOutputTokens,
		&retryAttempts,
		&retryDelay,
		&aiProviderName,
		&geminiApiKey,
		&geminiModelName,
		&geminiBaseURL,
		&openAIApiKey,
		&openAIModelName,
		&openAIBaseURL,
		&openRouterApiKey,
		&openRouterModelName,
		&openRouterBaseURL,
		&anthropicApiKey,
		&anthropicModelName,
		&anthropicBaseURL,
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
			setIfFlagIsSet(&app.opt.MaxRetries, retryAttempts)
			setIfFlagIsSet(&app.opt.RetryDelay, retryDelay)
			setIfFlagIsSet(&app.opt.AIProviderName, aiProviderName)
			setIfFlagIsSet(&app.opt.Providers.Gemini.ApiKey, geminiApiKey)
			setIfFlagIsSet(&app.opt.Providers.Gemini.ModelName, geminiModelName)
			setIfFlagIsSet(&app.opt.Providers.Gemini.BaseURL, geminiBaseURL)
			setIfFlagIsSet(&app.opt.Providers.OpenAI.ApiKey, openAIApiKey)
			setIfFlagIsSet(&app.opt.Providers.OpenAI.ModelName, openAIModelName)
			setIfFlagIsSet(&app.opt.Providers.OpenAI.BaseURL, openAIBaseURL)
			setIfFlagIsSet(&app.opt.Providers.OpenRouter.ApiKey, openRouterApiKey)
			setIfFlagIsSet(&app.opt.Providers.OpenRouter.ModelName, openRouterModelName)
			setIfFlagIsSet(&app.opt.Providers.OpenRouter.BaseURL, openRouterBaseURL)
			setIfFlagIsSet(&app.opt.Providers.Anthropic.ApiKey, anthropicApiKey)
			setIfFlagIsSet(&app.opt.Providers.Anthropic.ModelName, anthropicModelName)
			setIfFlagIsSet(&app.opt.Providers.Anthropic.BaseURL, anthropicBaseURL)
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
			ai.WithGeminiBaseURL(a.opt.Providers.Gemini.BaseURL),
		)
	case ai.ProviderOpenAI:
		provider = ai.NewOpenAI(
			a.opt.Providers.OpenAI.ApiKey,
			a.opt.Providers.OpenAI.ModelName,
			ai.WithOpenAIBaseURL(a.opt.Providers.OpenAI.BaseURL),
		)
	case ai.ProviderOpenRouter:
		provider = ai.NewOpenRouter(
			a.opt.Providers.OpenRouter.ApiKey,
			a.opt.Providers.OpenRouter.ModelName,
			ai.WithOpenRouterBaseURL(a.opt.Providers.OpenRouter.BaseURL),
		)
	case ai.ProviderAnthropic:
		provider = ai.NewAnthropic(
			a.opt.Providers.Anthropic.ApiKey,
			a.opt.Providers.Anthropic.ModelName,
			ai.WithAnthropicBaseURL(a.opt.Providers.Anthropic.BaseURL),
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

	var response *ai.Response

	if retryErr := retry.Do(ctx, func(ctx context.Context, attempt uint) (bool, error) {
		if attempt > 0 {
			debug.Printf("retrying after error (attempt %d of %d)", attempt, a.opt.MaxRetries)
		}

		var queryErr error

		response, queryErr = provider.Query(ctx, changes, commits,
			ai.WithShortMessageOnly(a.opt.ShortMessageOnly),
			ai.WithEmoji(a.opt.EnableEmoji),
			ai.WithMaxOutputTokens(a.opt.MaxOutputTokens),
		)
		if queryErr == nil {
			return false, nil
		}

		return !ai.IsRetryableError(queryErr), queryErr
	},
		retry.WithMaxAttempts(func() uint {
			if a.opt.MaxRetries == 0 {
				return 0 // unlimited
			}

			return a.opt.MaxRetries + 1
		}()),
		retry.WithDelay(a.opt.RetryDelay),
	); retryErr != nil {
		return retryErr
	}

	debug.Printf("prompt:\n%s", response.Prompt)
	debug.Printf("answer:\n%s\n", response.Answer)

	if _, err := fmt.Fprintln(os.Stdout, response.Answer); err != nil {
		return err
	}

	return nil
}
