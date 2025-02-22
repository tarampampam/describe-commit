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
  modelName: <openai-model-name>`,
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
}

func TestConfig_Merge(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveConfig    config.Config
		giveMergeWith config.Config
		wantConfig    config.Config
	}{
		"empty with empty": {
			giveConfig:    config.Config{},
			giveMergeWith: config.Config{},
			wantConfig:    config.Config{},
		},
		"full with full": {
			giveConfig: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(true)
				c.CommitHistoryLength = toPtr[int64](1)
				c.EnableEmoji = toPtr(true)
				c.MaxOutputTokens = toPtr[int64](1)
				c.AIProviderName = toPtr("foobar1")
				c.Gemini = &config.Gemini{
					ApiKey:    toPtr("GeminiApiKey1"),
					ModelName: toPtr("GeminiModelName1"),
				}
				c.OpenAI = &config.OpenAI{
					ApiKey:    toPtr("OpenAIApiKey1"),
					ModelName: toPtr("OpenAIModelName1"),
				}

				return
			}(),
			giveMergeWith: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(false)
				c.CommitHistoryLength = toPtr[int64](2)
				c.EnableEmoji = toPtr(false)
				c.MaxOutputTokens = toPtr[int64](2)
				c.AIProviderName = toPtr("foobar2")
				c.Gemini = &config.Gemini{
					ApiKey:    toPtr("GeminiApiKey2"),
					ModelName: toPtr("GeminiModelName2"),
				}
				c.OpenAI = &config.OpenAI{
					ApiKey:    toPtr("OpenAIApiKey2"),
					ModelName: toPtr("OpenAIModelName2"),
				}

				return
			}(),
			wantConfig: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(false)
				c.CommitHistoryLength = toPtr[int64](2)
				c.EnableEmoji = toPtr(false)
				c.MaxOutputTokens = toPtr[int64](2)
				c.AIProviderName = toPtr("foobar2")
				c.Gemini = &config.Gemini{
					ApiKey:    toPtr("GeminiApiKey2"),
					ModelName: toPtr("GeminiModelName2"),
				}
				c.OpenAI = &config.OpenAI{
					ApiKey:    toPtr("OpenAIApiKey2"),
					ModelName: toPtr("OpenAIModelName2"),
				}

				return
			}(),
		},
		"empty with full": {
			giveConfig: config.Config{},
			giveMergeWith: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(false)
				c.CommitHistoryLength = toPtr[int64](2)
				c.EnableEmoji = toPtr(false)
				c.MaxOutputTokens = toPtr[int64](2)
				c.AIProviderName = toPtr("foobar2")
				c.Gemini = &config.Gemini{
					ApiKey:    toPtr("GeminiApiKey2"),
					ModelName: toPtr("GeminiModelName2"),
				}
				c.OpenAI = &config.OpenAI{
					ApiKey:    toPtr("OpenAIApiKey2"),
					ModelName: toPtr("OpenAIModelName2"),
				}

				return
			}(),
			wantConfig: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(false)
				c.CommitHistoryLength = toPtr[int64](2)
				c.EnableEmoji = toPtr(false)
				c.MaxOutputTokens = toPtr[int64](2)
				c.AIProviderName = toPtr("foobar2")
				c.Gemini = &config.Gemini{
					ApiKey:    toPtr("GeminiApiKey2"),
					ModelName: toPtr("GeminiModelName2"),
				}
				c.OpenAI = &config.OpenAI{
					ApiKey:    toPtr("OpenAIApiKey2"),
					ModelName: toPtr("OpenAIModelName2"),
				}

				return
			}(),
		},
		"partial with partial": {
			giveConfig: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(true)
				c.Gemini = &config.Gemini{
					ApiKey: toPtr("GeminiApiKey1"),
				}

				return
			}(),
			giveMergeWith: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(false)
				c.Gemini = &config.Gemini{
					ApiKey: toPtr("GeminiApiKey2"),
				}

				return
			}(),
			wantConfig: func() (c config.Config) {
				c.ShortMessageOnly = toPtr(false)
				c.Gemini = &config.Gemini{
					ApiKey: toPtr("GeminiApiKey2"),
				}

				return
			}(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.giveConfig.Merge(&tc.giveMergeWith)

			if !reflect.DeepEqual(tc.giveConfig, tc.wantConfig) {
				t.Fatalf("expected:\n%+v\ngot:\n%+v", tc.wantConfig, tc.giveConfig)
			}
		})
	}
}

func toPtr[T any](v T) *T { return &v }
