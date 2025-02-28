package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gh.tarampamp.am/describe-commit/internal/yaml"
)

type (
	// Config is used to unmarshal the configuration file content.
	Config struct {
		// pointers are used to distinguish between unset and set values (nil = unset)
		ShortMessageOnly    *bool       `yaml:"shortMessageOnly"`
		CommitHistoryLength *int64      `yaml:"commitHistoryLength"`
		EnableEmoji         *bool       `yaml:"enableEmoji"`
		AIProviderName      *string     `yaml:"aiProvider"`
		MaxOutputTokens     *int64      `yaml:"maxOutputTokens"`
		Gemini              *Gemini     `yaml:"gemini"`
		OpenAI              *OpenAI     `yaml:"openai"`
		OpenRouter          *OpenRouter `yaml:"openrouter"`
		Anthropic           *Anthropic  `yaml:"anthropic"`
	}

	Gemini struct {
		ApiKey    *string `yaml:"apiKey"`
		ModelName *string `yaml:"modelName"`
	}

	OpenAI struct {
		ApiKey    *string `yaml:"apiKey"`
		ModelName *string `yaml:"modelName"`
	}

	OpenRouter struct {
		ApiKey    *string `yaml:"apiKey"`
		ModelName *string `yaml:"modelName"`
	}

	Anthropic struct {
		ApiKey    *string `yaml:"apiKey"`
		ModelName *string `yaml:"modelName"`
		Version   *string `yaml:"version"`
	}
)

// FromFile initializes self state by reading the configuration file from the provided path.
// To merge values from one file with another, call this method multiple times with different paths (values
// from the last file will overwrite the previous ones).
func (c *Config) FromFile(path string) error {
	if c == nil {
		return errors.New("config is nil")
	}

	var f, err = os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open the config file: %w", err)
	}

	defer func() { _ = f.Close() }()

	if err = yaml.NewDecoder(f).Decode(c); err != nil {
		if errors.Is(err, io.EOF) { // empty file
			return nil
		}

		return fmt.Errorf("failed to decode the config file: %w", err)
	}

	return nil
}
