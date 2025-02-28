package cli

import (
	"errors"
	"fmt"
	"os"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/config"
)

// options represents the command-line options. this struct should be used ONLY in this package (do not try to pass
// it somewhere else).
type options struct {
	ShortMessageOnly    bool
	CommitHistoryLength int64
	EnableEmoji         bool
	MaxOutputTokens     int64
	AIProviderName      string

	Providers struct {
		Gemini     struct{ ApiKey, ModelName string }
		OpenAI     struct{ ApiKey, ModelName string }
		OpenRouter struct{ ApiKey, ModelName string }
		Anthropic  struct{ ApiKey, ModelName string }
	}
}

func newOptionsWithDefaults() options {
	var opt = options{
		CommitHistoryLength: 20,                //nolint:mnd
		MaxOutputTokens:     500,               //nolint:mnd
		AIProviderName:      ai.ProviderGemini, // due to its free
	}

	opt.Providers.Gemini.ModelName = "gemini-2.0-flash"
	opt.Providers.OpenAI.ModelName = "gpt-4o-mini"
	opt.Providers.OpenRouter.ModelName = "nvidia/llama-3.1-nemotron-70b-instruct:free"
	opt.Providers.Anthropic.ModelName = "claude-3-7-sonnet-20250219"

	return opt
}

// UpdateFromConfigFile loads the configuration from the file(s) and applies it to the options.
// The values loaded from the earlier files will be overridden by those from the later files, with the last
// file taking the highest priority.
// Missing files and directories are ignored.
func (o *options) UpdateFromConfigFile(filePath []string) error {
	if len(filePath) == 0 {
		return nil
	}

	var cfg config.Config

	for _, path := range filePath {
		if path == "" {
			continue // skip empty paths
		}

		if stat, err := os.Stat(path); err != nil || stat.IsDir() {
			continue // skip missing files and directories
		}

		if err := cfg.FromFile(path); err != nil {
			return fmt.Errorf("failed to load the configuration file: %w", err)
		}
	}

	setIfSourceNotNil(&o.ShortMessageOnly, cfg.ShortMessageOnly)
	setIfSourceNotNil(&o.CommitHistoryLength, cfg.CommitHistoryLength)
	setIfSourceNotNil(&o.EnableEmoji, cfg.EnableEmoji)
	setIfSourceNotNil(&o.MaxOutputTokens, cfg.MaxOutputTokens)
	setIfSourceNotNil(&o.AIProviderName, cfg.AIProviderName)

	if sub := cfg.Gemini; sub != nil {
		setIfSourceNotNil(&o.Providers.Gemini.ApiKey, sub.ApiKey)
		setIfSourceNotNil(&o.Providers.Gemini.ModelName, sub.ModelName)
	}

	if sub := cfg.OpenAI; sub != nil {
		setIfSourceNotNil(&o.Providers.OpenAI.ApiKey, sub.ApiKey)
		setIfSourceNotNil(&o.Providers.OpenAI.ModelName, sub.ModelName)
	}

	if sub := cfg.OpenRouter; sub != nil {
		setIfSourceNotNil(&o.Providers.OpenRouter.ApiKey, sub.ApiKey)
		setIfSourceNotNil(&o.Providers.OpenRouter.ModelName, sub.ModelName)
	}

	if sub := cfg.Anthropic; sub != nil {
		setIfSourceNotNil(&o.Providers.Anthropic.ApiKey, sub.ApiKey)
		setIfSourceNotNil(&o.Providers.Anthropic.ModelName, sub.ModelName)
	}

	return nil
}

// setIfSourceNotNil sets the target value to the source value if both are not nil.
func setIfSourceNotNil[T any](target, source *T) {
	if target == nil || source == nil {
		return
	}

	*target = *source
}

func (o *options) Validate() error {
	if o.MaxOutputTokens <= 1 {
		return errors.New("max output tokens must be greater than 1")
	}

	if v := o.AIProviderName; !ai.IsProviderSupported(v) {
		return fmt.Errorf("unsupported AI provider: %s", v)
	}

	if o.AIProviderName == ai.ProviderGemini {
		if o.Providers.Gemini.ApiKey == "" {
			return errors.New("gemini API key is required")
		}

		if o.Providers.Gemini.ModelName == "" {
			return errors.New("gemini model name is required")
		}
	}

	if o.AIProviderName == ai.ProviderOpenAI {
		if o.Providers.OpenAI.ApiKey == "" {
			return errors.New("OpenAI API key is required")
		}

		if o.Providers.OpenAI.ModelName == "" {
			return errors.New("OpenAI model name is required")
		}
	}

	if o.AIProviderName == ai.ProviderOpenRouter {
		if o.Providers.OpenRouter.ApiKey == "" {
			return errors.New("OpenRouter API key is required")
		}

		if o.Providers.OpenRouter.ModelName == "" {
			return errors.New("OpenRouter model name is required")
		}
	}

	if o.AIProviderName == ai.ProviderAnthropic {
		if o.Providers.Anthropic.ApiKey == "" {
			return errors.New("Anthropic API key is required")
		}

		if o.Providers.Anthropic.ModelName == "" {
			return errors.New("Anthropic model name is required")
		}
	}

	return nil
}
