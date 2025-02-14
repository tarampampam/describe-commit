package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/describe-commit/internal/ai"
)

var configFilePathFlag = cli.StringFlag{
	Name:    "config-file",
	Aliases: []string{"c"},
	Usage:   "path to the configuration file",
	Value: func() string { // default file path depends on the user's OS
		const fileName = "describe-commit.yml"

		if dir, err := os.UserConfigDir(); err == nil {
			return filepath.Join(dir, fileName)
		}

		if cwd, err := os.Getwd(); err == nil {
			return filepath.Join(cwd, fileName)
		}

		return filepath.Join(".", fileName)
	}(),
	Sources:  cli.EnvVars("CONFIG_FILE"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
}

var shortMessageOnlyFlag = cli.BoolFlag{
	Name:     "short-message-only",
	Aliases:  []string{"s"},
	Usage:    "generate a short commit message (subject line) only",
	Sources:  cli.EnvVars("SHORT_MESSAGE_ONLY"),
	OnlyOnce: true,
}

var enableEmojiFlag = cli.BoolFlag{
	Name:     "enable-emoji",
	Aliases:  []string{"e"},
	Usage:    "enable emoji in the commit message",
	Sources:  cli.EnvVars("ENABLE_EMOJI"),
	OnlyOnce: true,
}

var maxOutputTokensFlag = cli.IntFlag{
	Name:     "max-output-tokens",
	Usage:    "maximum number of tokens in the output message",
	Sources:  cli.EnvVars("MAX_OUTPUT_TOKENS"),
	OnlyOnce: true,
	Value:    500, //nolint:mnd
}

var aiProviderNameFlag = cli.StringFlag{
	Name:     "ai-provider",
	Aliases:  []string{"ai"},
	Usage:    fmt.Sprintf("AI provider name (%s)", strings.Join(ai.SupportedProviders(), "/")),
	Sources:  cli.EnvVars("AI_PROVIDER"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
	Value:    ai.ProviderGemini, // due to its free
	Validator: func(s string) error {
		if !ai.IsProviderSupported(s) {
			return fmt.Errorf("unsupported AI provider: %s", s)
		}

		return nil
	},
}

var geminiApiKeyFlag = cli.StringFlag{
	Name:     "gemini-api-key",
	Aliases:  []string{"ga"},
	Usage:    "Gemini API key (https://bit.ly/4jZhiKI, as of February 2025 it's free)",
	Sources:  cli.EnvVars("GEMINI_API_KEY"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
}

var geminiModelNameFlag = cli.StringFlag{
	Name:     "gemini-model-name",
	Aliases:  []string{"gm"},
	Usage:    "Gemini model name (https://bit.ly/4i02ARR)",
	Sources:  cli.EnvVars("GEMINI_MODEL_NAME"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
	Value:    "gemini-2.0-flash",
}

var openAIApiKeyFlag = cli.StringFlag{
	Name:     "openai-api-key",
	Aliases:  []string{"oa"},
	Usage:    "OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account to access the API)",
	Sources:  cli.EnvVars("OPENAI_API_KEY"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
}

var openAIModelNameFlag = cli.StringFlag{
	Name:     "openai-model-name",
	Aliases:  []string{"om"},
	Usage:    "OpenAI model name (https://bit.ly/4hXCXkL)",
	Sources:  cli.EnvVars("OPENAI_MODEL_NAME"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
	Value:    "gpt-4o-mini",
}
