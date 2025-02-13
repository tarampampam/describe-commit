package cli

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
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

var geminiApiKeyFlag = cli.StringFlag{
	Name:     "gemini-api-key",
	Aliases:  []string{"ga"},
	Usage:    "Gemini API key",
	Sources:  cli.EnvVars("GEMINI_API_KEY"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
}

var geminiModelNameFlag = cli.StringFlag{
	Name:     "gemini-model-name",
	Aliases:  []string{"gm"},
	Usage:    "Gemini model name",
	Sources:  cli.EnvVars("GEMINI_MODEL_NAME"),
	OnlyOnce: true,
	Config:   cli.StringConfig{TrimSpace: true},
	Value:    "gemini-2.0-flash",
}
