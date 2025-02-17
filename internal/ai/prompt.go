package ai

import (
	"fmt"
	"strings"
)

const (
	gitDiffBegin, gitDiffEnd = "[---GIT-DIFF-BEGIN---]", "[---GIT-DIFF-END---]"
	gitLogBegin, gitLogEnd   = "[---GIT-LOG-BEGIN---]", "[---GIT-LOG-END---]"
)

// wrapChanges wraps the provided diff output between the specified markers (to help the AI identify the changes).
func wrapChanges(diff string) string {
	return fmt.Sprintf("%s\n%s\n%s", gitDiffBegin, diff, gitDiffEnd)
}

// wrapCommits wraps the provided log output between the specified markers (to help the AI too).
func wrapCommits(log string) string {
	return fmt.Sprintf("%s\n%s\n%s", gitLogBegin, log, gitLogEnd)
}

func GeneratePrompt(opts ...Option) string { //nolint:funlen
	var (
		opt = options{}.Apply(opts...)
		b   strings.Builder
	)

	b.Grow(2560) //nolint:mnd // pre-allocate memory for the string builder

	{ // role
		b.WriteString("## Role\n")
		b.WriteString("You are an AI assistant specializing in crafting Git commit messages.\n")

		b.WriteRune('\n')
	}

	{ // task
		b.WriteString("## Task\n")
		b.WriteString("Generate a concise, informative, and well-structured **SINGLE** Git commit ")
		b.WriteString("message based on the provided input.\n")

		b.WriteRune('\n')
	}

	{ // input
		b.WriteString("## Input\n")
		b.WriteString("You will receive:\n")
		b.WriteString(fmt.Sprintf(
			"1. The output of `git diff`, showing the staged changes, is wrapped between `%s` and `%s`.\n",
			gitDiffBegin, gitDiffEnd,
		))
		b.WriteString(fmt.Sprintf(
			"2. The output of `git log`, presenting recent commit history, is wrapped between `%s` and `%s`.\n",
			gitLogBegin, gitLogEnd,
		))
		b.WriteRune('\n')
	}

	{ // output
		b.WriteString("## Output\n")
		b.WriteString("Produce a commit message in plain text without wrapping it in backticks, ")
		b.WriteString("quotes, or code blocks.\n")

		b.WriteRune('\n')
	}

	{ // guidelines
		b.WriteString("## Guidelines\n")
		b.WriteString("### Format\n")

		const (
			convFormat = "<type>(<scope>): <message>"
			convDesc   = "- `<type>`: Choose from 'feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', " +
				"'ci', 'chore'. Carefully analyze **ALL** changes made across all files in the provided diff to " +
				"determine the primary impact. Use the lowercase form of the type.\n" +
				"- `<scope>`: (optional but recommended) Specify the affected module (e.g., 'auth', " +
				"'api', 'ui'). If the changes span multiple areas, omit this.\n" +
				"- `<message>`: Use **imperative tone** (max 72 characters), describe **WHAT** " +
				"was changed and **WHY**. No periods at the end of the message.\n"
		)

		b.WriteString("Follow the Conventional Commit format: `")

		if !opt.EnableEmoji {
			b.WriteString(convFormat)
			b.WriteString("`\n")
			b.WriteString(convDesc)
		} else {
			b.WriteString("<emoji> ")
			b.WriteString(convFormat)
			b.WriteString("`\n")
			b.WriteString("- `<emoji>`: Use GitMoji convention to preface the commit. Choose from (emoji, description):\n")
			b.WriteString("  - 🐛, Fix a bug\n")
			b.WriteString("  - ✨, Introduce new features\n")
			b.WriteString("  - 📝, Add or update documentation\n")
			b.WriteString("  - 🚀, Deploy-related changes\n")
			b.WriteString("  - ✅, Add, update, or pass tests\n")
			b.WriteString("  - ♻️, Refactor code\n")
			b.WriteString("  - ⬆️, Upgrade dependencies\n")
			b.WriteString("  - 🔧, Add or update configuration files\n")
			b.WriteString("  - 🌐, Internationalization and localization\n")
			b.WriteString("  - 💡, Add or update comments in source code\n")
			b.WriteString(convDesc)
		}

		if !opt.ShortMessageOnly {
			b.WriteString("### Commit Message Structure\n")
			b.WriteString("- **WHAT** and **WHY**: Summarize what was changed and why the change was needed.\n")
			b.WriteString("- **Avoid**: Vague messages like \"Updated files\" or \"Fixed bugs.\" Be specific.\n")
			b.WriteString("- **Tense**: Use present tense (e.g., Fix, Add, Refactor), not past tense ")
			b.WriteString("(e.g., Fixed, Added).\n")
			b.WriteString("- **Format**: The first line should follow the Conventional Commit format, followed ")
			b.WriteString("by a blank line, then a detailed description if necessary.\n")
			b.WriteString("- **No periods**: Omit periods at the end of each line.\n")
			b.WriteString("### Commit Body (if necessary)\n")
			b.WriteString("- Start with a single-line summary.\n")
			b.WriteString("- Exclude the provided diff output from the commit message.\n")
			b.WriteString("- For complex changes, add a detailed description after a blank line:\n")
			b.WriteString("  - Explain additional context or implementation details.\n")
			b.WriteString("  - Include a summary and key points when necessary.\n")
			b.WriteString("  - Avoid excessive detail; provide only what's needed for understanding.\n")
			b.WriteString("- Avoid starting with \"This commit\"; directly describe the changes.\n")
		} else {
			b.WriteString("### Focus on the primary purpose of the commit\n")
			b.WriteString("- Summarize all changes in a single, meaningful message.\n")
			b.WriteString("- Explain why the changes were made, not just what was modified.\n")
		}

		b.WriteRune('\n')
		b.WriteString("**Example**:\n")
		b.WriteRune('\n')
		b.WriteString("```\n")

		if opt.EnableEmoji {
			b.WriteString("✨ ")
		}

		b.WriteString("feat(api): Add rate-limiting to endpoints\n")

		if !opt.ShortMessageOnly {
			b.WriteRune('\n')
			b.WriteString("Implemented rate-limiting on all API endpoints to enhance security by preventing abuse ")
			b.WriteString("through request limits. Integrated Redis to track API requests, aiding future analytics. ")
			b.WriteString("Configuration is adjustable via environment variables.\n")
			b.WriteRune('\n')
			b.WriteString("- Enforces request limits to prevent abuse\n")
			b.WriteString("- Utilizes Redis for tracking API requests\n")
			b.WriteString("- Configurable through environment variables\n")
		}

		b.WriteString("```\n")

		b.WriteRune('\n')
	}

	{ // security
		b.WriteString("## Security\n")
		b.WriteString("- Exclude sensitive data (passwords, API keys, personal information, etc.) ")
		b.WriteString("or code snippets from the commit message.\n")

		b.WriteRune('\n')
	}

	{ // instructions
		b.WriteString("## Instructions for the AI\n")
		b.WriteString("- Analyze the provided `git diff` to understand the current changes.\n")
		b.WriteString("- Analyze the provided `git log` output to better understand the codebase functionally, ")
		b.WriteString("features, and recent changes, but do not include this information in the commit message ")
		b.WriteString("or use it as a template.\n")
		b.WriteString("- Synthesize this information to generate a commit message that accurately reflects ")
		b.WriteString("the current changes in the context of the project's history.\n")
	}

	return b.String()
}
