package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gh.tarampamp.am/describe-commit/internal/ai"
)

type options struct {
	ConfigFilePath      string
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

func newOptionsWithDefaults() options {
	var opt = options{
		ConfigFilePath: func() string { // default file path depends on the user's OS
			const fileName = "describe-commit.yml"

			// used to override the default config file path in readme generation
			if v, ok := os.LookupEnv("DEFAULT_CONFIG_FILE_DIR"); ok {
				return filepath.Join(v, fileName)
			}

			if dir, err := os.UserConfigDir(); err == nil {
				return filepath.Join(dir, fileName)
			}

			if cwd, err := os.Getwd(); err == nil {
				return filepath.Join(cwd, fileName)
			}

			return filepath.Join(".", fileName)
		}(),
		CommitHistoryLength: 20,                //nolint:mnd
		MaxOutputTokens:     500,               //nolint:mnd
		AIProviderName:      ai.ProviderGemini, // due to its free
	}

	opt.Providers.Gemini.ModelName = "gemini-2.0-flash"
	opt.Providers.OpenAI.ModelName = "gpt-4o-mini"

	return opt
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

	return nil
}
