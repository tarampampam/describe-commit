package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/cli/cmd"
	"gh.tarampamp.am/describe-commit/internal/config"
	"gh.tarampamp.am/describe-commit/internal/git"
	"gh.tarampamp.am/describe-commit/internal/version"
)

type (
	App struct {
		cmd cmd.Command

		opt options
	}

	options struct {
		ShortMessageOnly    bool
		CommitHistoryLength int64
		EnableEmoji         bool
		MaxOutputTokens     int64
		AIProviderName      string

		Providers struct {
			Gemini struct{ ApiKey, ModelName string }
			OpenAI struct{ ApiKey, ModelName string }
		}
	}
)

func NewApp(name string) *App {
	var app = App{
		cmd: cmd.Command{
			Name:    name,
			Version: version.Version(),
		},
	}

	// set default options

	var (
		configFile = cmd.Flag[string]{
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
		}
		shortMessageOnly = cmd.Flag[bool]{
			Names:   []string{"short-message-only", "s"},
			Usage:   "Generate a short commit message (subject line) only",
			EnvVars: []string{"SHORT_MESSAGE_ONLY"},
		}
		commitHistoryLength = cmd.Flag[int64]{
			Names:   []string{"commit-history-length", "cl", "hl"},
			Usage:   "Number of previous commits from the Git history (0 = disabled)",
			EnvVars: []string{"COMMIT_HISTORY_LENGTH"},
			Default: 20, //nolint:mnd
		}
		enableEmoji = cmd.Flag[bool]{
			Names:   []string{"enable-emoji", "e"},
			Usage:   "Enable emoji in the commit message",
			EnvVars: []string{"ENABLE_EMOJI"},
		}
		maxOutputTokens = cmd.Flag[int64]{
			Names:   []string{"max-output-tokens"},
			Usage:   "Maximum number of tokens in the output message",
			EnvVars: []string{"MAX_OUTPUT_TOKENS"},
			Default: 500, //nolint:mnd
		}
		aiProviderName = cmd.Flag[string]{
			Names:   []string{"ai-provider", "ai"},
			Usage:   fmt.Sprintf("AI provider name (%s)", strings.Join(ai.SupportedProviders(), "|")),
			EnvVars: []string{"AI_PROVIDER"},
			Default: ai.ProviderGemini, // due to its free
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
		}
		geminiModelName = cmd.Flag[string]{
			Names:   []string{"gemini-model-name", "gm"},
			Usage:   "Gemini model name (https://bit.ly/4i02ARR)",
			EnvVars: []string{"GEMINI_MODEL_NAME"},
			Default: "gemini-2.0-flash",
		}
		openAIApiKey = cmd.Flag[string]{
			Names:   []string{"openai-api-key", "oa"},
			Usage:   "OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account)",
			EnvVars: []string{"OPENAI_API_KEY"},
		}
		openAIModelName = cmd.Flag[string]{
			Names:   []string{"openai-model-name", "om"},
			Usage:   "OpenAI model name (https://bit.ly/4hXCXkL)",
			EnvVars: []string{"OPENAI_MODEL_NAME"},
			Default: "gpt-4o-mini",
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
		var cfg config.Config

		{ // set default options
			app.opt.ShortMessageOnly = shortMessageOnly.Default
			app.opt.CommitHistoryLength = commitHistoryLength.Default
			app.opt.EnableEmoji = enableEmoji.Default
			app.opt.MaxOutputTokens = maxOutputTokens.Default
			app.opt.AIProviderName = aiProviderName.Default
			app.opt.Providers.Gemini.ApiKey = geminiApiKey.Default
			app.opt.Providers.Gemini.ModelName = geminiModelName.Default
			app.opt.Providers.OpenAI.ApiKey = openAIApiKey.Default
			app.opt.Providers.OpenAI.ModelName = openAIModelName.Default
		}

		// override the default options with the configuration file values, if they are set
		if configFile.Value != nil && *configFile.Value != "" {
			if err := cfg.FromFile(*configFile.Value); err == nil {
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
				debugf("failed to load the configuration file: %s", err)
			}
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

		{ // validate the options
			if app.opt.MaxOutputTokens <= 1 {
				return errors.New("max output tokens must be greater than 1")
			}
			if v := app.opt.AIProviderName; !ai.IsProviderSupported(v) {
				return fmt.Errorf("unsupported AI provider: %s", v)
			}

			if app.opt.AIProviderName == ai.ProviderGemini {
				if app.opt.Providers.Gemini.ApiKey == "" {
					return errors.New("gemini API key is required")
				}
				if app.opt.Providers.Gemini.ModelName == "" {
					return errors.New("gemini model name is required")
				}
			}

			if app.opt.AIProviderName == ai.ProviderOpenAI {
				if app.opt.Providers.OpenAI.ApiKey == "" {
					return errors.New("OpenAI API key is required")
				}
				if app.opt.Providers.OpenAI.ModelName == "" {
					return errors.New("OpenAI model name is required")
				}
			}
		}

		var wd, wdErr = app.getWorkingDir(args)
		if wdErr != nil {
			return fmt.Errorf("wrong working directory: %w", wdErr)
		}

		return app.run(ctx, wd)
	}

	return &app
}

// getWorkingDir returns the working directory to use for the application.
func (a *App) getWorkingDir(args []string) (string, error) {
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

	return dir, nil
}

func (a *App) Run(ctx context.Context, args []string) error { return a.cmd.Run(ctx, args) }

func (a *App) run(ctx context.Context, workingDir string) error {
	debugf("AI provider: %s", a.opt.AIProviderName)

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

	debugf("working directory: %s", workingDir)

	var (
		eg, egCtx        = errgroup.WithContext(ctx)
		changes, commits string
	)

	eg.Go(func() (err error) { changes, err = git.Diff(egCtx, workingDir); return }) //nolint:nlreturn

	if histLen := int(a.opt.CommitHistoryLength); histLen > 0 {
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
		ai.WithShortMessageOnly(a.opt.ShortMessageOnly),
		ai.WithEmoji(a.opt.EnableEmoji),
		ai.WithMaxOutputTokens(a.opt.MaxOutputTokens),
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
