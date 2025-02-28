package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gh.tarampamp.am/describe-commit/internal/config"
)

func TestConfig_FromFile(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveContent   string
		wantStruct    config.Config
		wantErrSubstr string
	}{
		"empty file": {
			giveContent: "",
			wantStruct:  config.Config{},
		},
		"full config": {
			giveContent: `
shortMessageOnly: true
commitHistoryLength: 312312
enableEmoji: false
maxOutputTokens: 123123123
aiProvider: foobar
gemini:
  apiKey: <your-api-key>
  modelName: <gemini-model-name>
openai:
  apiKey: <openai-api-key>
  modelName: <openai-model-name>
openrouter:
  apiKey: <openrouter-api-key>
  modelName: <openrouter-model-name>
anthropic:
  apiKey: <anthropic-api-key>
  modelName: <anthropic-model-name>`,
			wantStruct: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(true)
				c.CommitHistoryLength = toPtr[int64](312312)
				c.EnableEmoji = toPtr(false)
				c.MaxOutputTokens = toPtr[int64](123123123)
				c.AIProviderName = toPtr("foobar")
				c.Gemini = &config.Gemini{
					ApiKey:    toPtr("<your-api-key>"),
					ModelName: toPtr("<gemini-model-name>"),
				}
				c.OpenAI = &config.OpenAI{
					ApiKey:    toPtr("<openai-api-key>"),
					ModelName: toPtr("<openai-model-name>"),
				}
				c.OpenRouter = &config.OpenRouter{
					ApiKey:    toPtr("<openrouter-api-key>"),
					ModelName: toPtr("<openrouter-model-name>"),
				}
				c.Anthropic = &config.Anthropic{
					ApiKey:    toPtr("<anthropic-api-key>"),
					ModelName: toPtr("<anthropic-model-name>"),
				}

				return
			}(),
		},
		"partial": {
			giveContent: `
shortMessageOnly: true
gemini:
  apiKey: <your-api-key>`,
			wantStruct: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(true)
				c.Gemini = &config.Gemini{
					ApiKey: toPtr("<your-api-key>"),
				}

				return
			}(),
		},

		"broken yaml": {
			giveContent:   "$rossia-budet-svobodnoy$",
			wantErrSubstr: "failed to decode the config file",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var filePath = filepath.Join(t.TempDir(), "config.yml")

			if err := os.WriteFile(filePath, []byte(tc.giveContent), 0o600); err != nil {
				t.Fatalf("failed to create a config file: %v", err)
			}

			var (
				c   config.Config
				err = c.FromFile(filePath)
			)

			if tc.wantErrSubstr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				// assert the structure
				if !reflect.DeepEqual(c, tc.wantStruct) {
					t.Fatalf("expected: %+v, got: %+v", tc.wantStruct, c)
				}

				return
			}

			if err == nil {
				t.Fatalf("expected an error, got nil")
			}

			if got := err.Error(); !strings.Contains(got, tc.wantErrSubstr) {
				t.Fatalf("expected error to contain %q, got %q", tc.wantErrSubstr, got)
			}
		})
	}

	t.Run("merge", func(t *testing.T) {
		var (
			tmpDir  = t.TempDir()
			config1 = filepath.Join(tmpDir, "config1.yml")
			config2 = filepath.Join(tmpDir, "config2.yml")
		)

		// create config files
		for _, err := range []error{
			os.WriteFile(config1, []byte(`
shortMessageOnly: true
gemini:
  apiKey: gemini-api-key-1
`), 0o600),
			os.WriteFile(config2, []byte(`
shortMessageOnly: false
gemini:
  apiKey: gemini-api-key-2
openai:
  apiKey: openai-api-key
  modelName: openai-model-name
`), 0o600),
		} {
			if err != nil {
				t.Fatalf("failed to create a config file: %v", err)
			}
		}

		var cfg config.Config

		// read the first file
		if err := cfg.FromFile(config1); err != nil {
			t.Fatalf("failed to read the first config file: %v", err)
		}

		// read the second file
		if err := cfg.FromFile(config2); err != nil {
			t.Fatalf("failed to read the second config file: %v", err)
		}

		// assert the structure
		if !reflect.DeepEqual(cfg, config.Config{
			ShortMessageOnly: toPtr(false),
			Gemini: &config.Gemini{
				ApiKey: toPtr("gemini-api-key-2"),
			},
			OpenAI: &config.OpenAI{
				ApiKey:    toPtr("openai-api-key"),
				ModelName: toPtr("openai-model-name"),
			},
		}) {
			t.Fatalf("unexpected config: %+v", cfg)
		}
	})
}

func toPtr[T any](v T) *T { return &v }
