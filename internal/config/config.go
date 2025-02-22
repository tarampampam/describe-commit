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
		// important note: do NOT forget to update the `Merge` method when adding new fields!
		ShortMessageOnly    *bool   `yaml:"shortMessageOnly"`
		CommitHistoryLength *int64  `yaml:"commitHistoryLength"`
		EnableEmoji         *bool   `yaml:"enableEmoji"`
		AIProviderName      *string `yaml:"aiProvider"`
		MaxOutputTokens     *int64  `yaml:"maxOutputTokens"`
		Gemini              *Gemini `yaml:"gemini"`
		OpenAI              *OpenAI `yaml:"openai"`
	}

	Gemini struct {
		ApiKey    *string `yaml:"apiKey"`
		ModelName *string `yaml:"modelName"`
	}

	OpenAI struct {
		ApiKey    *string `yaml:"apiKey"`
		ModelName *string `yaml:"modelName"`
	}
)

// FromFile initializes self state by reading the configuration file from the provided path.
func (c *Config) FromFile(path string) error {
	if c == nil {
		return errors.New("config is nil")
	}

	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", path)
		}

		return err
	} else if stat.IsDir() {
		return fmt.Errorf("config file path is a directory: %s", path)
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

// Merge updates the current configuration with the provided one.
func (c *Config) Merge(with *Config) {
	if with == nil {
		return
	}

	setIfNotNilPtr(&c.ShortMessageOnly, with.ShortMessageOnly)
	setIfNotNilPtr(&c.CommitHistoryLength, with.CommitHistoryLength)
	setIfNotNilPtr(&c.EnableEmoji, with.EnableEmoji)
	setIfNotNilPtr(&c.AIProviderName, with.AIProviderName)
	setIfNotNilPtr(&c.MaxOutputTokens, with.MaxOutputTokens)

	if with.Gemini != nil {
		if c.Gemini == nil {
			c.Gemini = &Gemini{}
		}

		setIfNotNilPtr(&c.Gemini.ApiKey, with.Gemini.ApiKey)
		setIfNotNilPtr(&c.Gemini.ModelName, with.Gemini.ModelName)
	}

	if with.OpenAI != nil {
		if c.OpenAI == nil {
			c.OpenAI = &OpenAI{}
		}

		setIfNotNilPtr(&c.OpenAI.ApiKey, with.OpenAI.ApiKey)
		setIfNotNilPtr(&c.OpenAI.ModelName, with.OpenAI.ModelName)
	}
}

// setIfNotNilPtr ensures the target pointer is initialized and updates it if the source is not nil.
func setIfNotNilPtr[T any](target **T, source *T) {
	if source != nil {
		if *target == nil {
			*target = new(T) // allocate memory if nil
		}

		**target = *source // copy the value
	}
}
