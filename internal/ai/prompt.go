package ai

import (
	"strings"
)

func GeneratePrompt(opts ...Option) string { //nolint:funlen
	var (
		opt = options{}.Apply(opts...)
		b   strings.Builder
	)

	b.Grow(2560) //nolint:mnd // pre-allocate memory for the string builder

	b.WriteString("## **Role**: You are an AI assistant acting as a Git commit message author.")
	b.WriteString("## **Task**: I will provide the output of `git diff --staged`. Your job is to convert it ")
	b.WriteString("into a concise, informative, and well-structured **SINGLE** Git commit message as plain text, ")
	b.WriteString("without wrapping it in backticks, quotes, or code blocks")

	b.WriteString("## **Security**: Never include sensitive data (passwords, API keys, personal information, etc.) or ")
	b.WriteString("code snippets in the commit message")

	b.WriteString("## **Guidelines**: \n")
	b.WriteString("### Follow the **Conventional Commit** format: ")

	const (
		convFormat = "<type>(<scope>): <message>"
		convDesc   = "  - `<type>`: Choose from 'feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', 'ci', " +
			"'chore' (lowercase only).\n" +
			"  - `<scope>`: (optional but recommended) Specify the affected module (e.g., 'auth', 'api', 'ui'). If the " +
			"changes span multiple areas, omit this.\n" +
			"  - `<message>`: Use **imperative tone** (max 72 characters), describe **WHAT** was changed and **WHY**. " +
			"No periods at the end of the message."
	)

	if !opt.EnableEmoji {
		b.WriteString("`" + convFormat + "`\n\n")
		b.WriteString(convDesc)
		b.WriteString("\n  **Examples**:\n")
		b.WriteString("  - `fix(auth): Resolve login token expiration issue`\n" +
			"  - `feat(api): Add caching to improve response times`\n" +
			"  - `refactor(ui): Simplify navbar component logic`")
	} else {
		b.WriteString("`<emoji> " + convFormat + "`\n\n")
		b.WriteString("  - `<emoji>`: Use GitMoji to preface the commit. Choose from:\n" +
			"    - 🐛 fix: Fix a bug\n" +
			"    - ✨ feat: Introduce new features\n" +
			"    - 📝 docs: Add or update documentation\n" +
			"    - 🚀 ci: Deploy-related changes\n" +
			"    - ✅ test: Add, update, or pass tests\n" +
			"    - ♻️ refactor: Improve code without changing functionality\n" +
			"    - ⬆️ chore: Upgrade dependencies\n" +
			"    - 🔧 chore: Update configuration files\n" +
			"    - 🌐 chore: Internationalization & localization\n" +
			"    - 💡 chore: Add or update comments\n" + convDesc)
		b.WriteString("\n\n  **Examples**:\n\n")
		b.WriteString("  - `🐛 fix(auth): Resolve login token expiration issue`\n" +
			"  - `✨ feat(api): Add caching to improve response times`\n" +
			"  - `♻️ refactor(ui): Simplify navbar component logic`")
	}

	b.WriteString("\n\n")

	if opt.ShortMessageOnly {
		b.WriteString(`### Focus on the primary purpose of the commit:

- Summarize all changes in a single, meaningful message.
- Explain why the changes were made, not just what was modified.`)

		b.WriteString("\n\nExample:\n")

		if !opt.EnableEmoji {
			b.WriteString("feat(api): Add rate-limiting to endpoints")
		} else {
			b.WriteString(`✨ feat(api): Add rate-limiting to endpoints`)
		}
	} else {
		b.WriteString(`### Commit Message Structure:

- **WHAT & WHY**: Focus on summarizing **what** was changed and **why** the change was needed.
- **Avoid**: Vague messages like "Updated files" or "Fixed bugs". Be specific.
- **Tense**: Use present tense (Fix, Add, Refactor), not past tense (Fixed, Added).
- **Format**: The first line should follow the Conventional Commit format, followed by a blank line, and then a
  detailed description.
- **No periods**: Keep commit messages clean and concise without period at the end of each line.

### Commit Body (if necessary):

- Start the commit message with a single line summary.
- Never include the provided diff output in the commit message.
- If the change is more complex, add a detailed description in bullet points after a blank line.
  - Explain additional context or implementation details.
  - Include a **summary** and a list of **key points** when necessary.
  - Avoid excessive detail – keep it to what is needed for understanding.
- Don't start it with "This commit", just describe the changes.`)

		b.WriteString("\n\n#### Example:\n\n")

		if opt.EnableEmoji {
			b.WriteString("✨ ")
		}

		b.WriteString(`feat(api): Add rate-limiting to endpoints

Security has been enhanced by implementing rate-limiting on all API endpoints,
preventing abuse by enforcing request limits. Additionally, Redis has been
integrated to track API requests, which may be useful for future analytics.
The configuration can be adjusted via environment variables.

- Prevents abuse by enforcing request limits
- Uses Redis for tracking API requests
- Configurable via environment variables`)
	}

	return b.String()
}
