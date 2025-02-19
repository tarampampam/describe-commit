package cli

import (
	"errors"
	"fmt"
	"time"

	"gh.tarampamp.am/describe-commit/internal/ai"
	"gh.tarampamp.am/describe-commit/internal/config"
)

type options struct {
	ConfigFilePath option3[string]

	ShortMessageOnly    option3[bool]
	CommitHistoryLength option3[int64]
	EnableEmoji         option3[bool]
	MaxOutputTokens     option3[int64]
	AIProviderName      option3[string]

	Providers struct {
		Gemini struct {
			ApiKey    option3[string]
			ModelName option3[string]
		}

		OpenAI struct {
			ApiKey    option3[string]
			ModelName option3[string]
		}
	}
}

type option3[T bool | int | int64 | string | uint | uint64 | float64 | time.Duration] struct {
	Flag   T // this value is set by the user (using flags or environment variables)
	Config T // this one - from the config file (also used as a default value)
}

// SetFromConfig checks if the value is not nil and sets the option value as received from the config file.
func (o *option3[T]) SetFromConfig(v *T) {
	if v != nil {
		o.Config = *v
	}
}

// Get the value of the option regarding priority (flag > config).
func (o *option3[T]) Get() T {
	var empty T

	if o.Flag != empty {
		return o.Flag
	}

	return o.Config
}

// FromConfig sets the options from the config file (provided as a struct).
func (o options) FromConfig(cfg config.Config) {
	o.ShortMessageOnly.SetFromConfig(cfg.ShortMessageOnly)
	o.CommitHistoryLength.SetFromConfig(cfg.CommitHistoryLength)
	o.EnableEmoji.SetFromConfig(cfg.EnableEmoji)
	o.MaxOutputTokens.SetFromConfig(cfg.MaxOutputTokens)
	o.AIProviderName.SetFromConfig(cfg.AIProviderName)

	if sub := cfg.Gemini; sub != nil {
		o.Providers.Gemini.ApiKey.SetFromConfig(sub.ApiKey)
		o.Providers.Gemini.ModelName.SetFromConfig(sub.ModelName)
	}

	if sub := cfg.OpenAI; sub != nil {
		o.Providers.OpenAI.ApiKey.SetFromConfig(sub.ApiKey)
		o.Providers.OpenAI.ModelName.SetFromConfig(sub.ModelName)
	}
}

// Validate checks if the options are valid.
func (o options) Validate() error {
	if o.MaxOutputTokens.Get() <= 1 {
		return errors.New("max output tokens must be greater than 1")
	}

	if v := o.AIProviderName.Get(); !ai.IsProviderSupported(v) {
		return fmt.Errorf("unsupported AI provider: %s", v)
	}

	if o.AIProviderName.Get() == ai.ProviderGemini {
		if o.Providers.Gemini.ApiKey.Get() == "" {
			return errors.New("gemini API key is required")
		}

		if o.Providers.Gemini.ModelName.Get() == "" {
			return errors.New("gemini model name is required")
		}
	}

	if o.AIProviderName.Get() == ai.ProviderOpenAI {
		if o.Providers.OpenAI.ApiKey.Get() == "" {
			return errors.New("OpenAI API key is required")
		}

		if o.Providers.OpenAI.ModelName.Get() == "" {
			return errors.New("OpenAI model name is required")
		}
	}

	return nil
}
