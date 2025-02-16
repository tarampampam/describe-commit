package ai_test

import (
	"strings"
	"testing"

	"gh.tarampamp.am/describe-commit/internal/ai"
)

func TestGeneratePrompt(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		giveOpts     []ai.Option
		wantContains []string
		wantNot      []string
	}{
		"short without emoji": {
			giveOpts: []ai.Option{
				ai.WithShortMessageOnly(false),
				ai.WithEmoji(false),
			},
			wantContains: []string{
				// role
				"Role", "Git commit messages",

				// task
				"Task", "well-structured **SINGLE** Git commit", "based on the provided input",

				// input
				"Input", "will receive", "git diff", "git log",

				// output
				"Output", "commit message in plain text without wrapping",

				// guidelines
				"Guidelines",
				"Format", "`<type>(<scope>): <message>`", "`<type>`", "`<scope>`", "`<message>`",
				"Commit Message Structure", "Summarize what was changed", "Use present tense",
				"Commit Body", "Start with a single-line summary", "Exclude the provided diff", "add a detailed description",
				"Example", "feat(api): Add rate-limiting to endpoints", "Implemented rate-limiting", "Enforces request limits",

				// security
				"Security", "Exclude sensitive data", "or code snippets",

				// instructions
				"Instructions for the AI", "Analyze the provided", "Synthesize this information",
			},
			wantNot: []string{
				// guidelines
				"<emoji>", "`<emoji>`", "🐛", "✨", "📝", "🚀", "✅", "♻️", "⬆️", "🔧", "🌐", "💡",
				"Focus on the primary purpose", "Summarize all changes in a single", "Explain why the changes were made",
			},
		},
		"long with emoji": {
			giveOpts: []ai.Option{
				ai.WithShortMessageOnly(true),
				ai.WithEmoji(true),
			},
			wantContains: []string{
				// role
				"Role", "Git commit messages",

				// task
				"Task", "well-structured **SINGLE** Git commit", "based on the provided input",

				// input
				"Input", "will receive", "git diff", "git log",

				// output
				"Output", "commit message in plain text without wrapping",

				// guidelines
				"Guidelines",
				"Format", "`<emoji> <type>(<scope>): <message>`", "`<emoji>`", "`<type>`", "`<scope>`", "`<message>`",
				"🐛", "✨", "📝", "🚀", "✅", "♻️", "⬆️", "🔧", "🌐", "💡",
				"Example", "✨ feat(api): Add rate-limiting to endpoints",
				"Focus on the primary purpose", "Summarize all changes in a single", "Explain why the changes were made",

				// security
				"Security", "Exclude sensitive data", "or code snippets",

				// instructions
				"Instructions for the AI", "Analyze the provided", "Synthesize this information",
			},
			wantNot: []string{
				// guidelines
				"Commit Message Structure", "Summarize what was changed", "Use present tense",
				"Commit Body", "Start with a single-line summary", "Exclude the provided diff", "add a detailed description",
				"Implemented rate-limiting", "Enforces request limits",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ai.GeneratePrompt(tc.giveOpts...)

			for _, want := range tc.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("want %q to contain %q", got, want)
				}
			}

			for _, want := range tc.wantNot {
				if strings.Contains(got, want) {
					t.Errorf("want %q to not contain %q", got, want)
				}
			}
		})
	}
}
