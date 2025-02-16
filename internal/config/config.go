package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is used to unmarshal the configuration file content.
type ( // pointers are used to distinguish between unset and set values (nil = unset)
	Config struct {
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
