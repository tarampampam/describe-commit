# describe-commit

> [!WARNING]
> This project is under active development and is not yet ready for use

## Gemini

- Generate your own Gemini API token: <https://aistudio.google.com/app/apikey>
- View usage data: <https://console.cloud.google.com/apis/api/generativelanguage.googleapis.com/metrics>

<!--GENERATED:CLI_DOCS-->
<!-- Documentation inside this block generated by github.com/urfave/cli-docs/v3; DO NOT EDIT -->
## CLI interface

To debug the application, set the DEBUG environment variable to `true`.

This tool uses AI to generate a commit message based on the changes made.

Usage:

```bash
$ describe-commit [GLOBAL FLAGS] [COMMAND] [COMMAND FLAGS] [dir-path]
```

Global flags:

| Name                               | Description                                                                                      |               Default value               | Environment variables |
|------------------------------------|--------------------------------------------------------------------------------------------------|:-----------------------------------------:|:---------------------:|
| `--config-file="…"` (`-c`)         | path to the configuration file                                                                   | `/depends/on/your-os/describe-commit.yml` |     `CONFIG_FILE`     |
| `--short-message-only` (`-s`)      | generate a short commit message (subject line) only                                              |                  `false`                  |  `SHORT_MESSAGE_ONLY` |
| `--enable-emoji` (`-e`)            | enable emoji in the commit message                                                               |                  `false`                  |    `ENABLE_EMOJI`     |
| `--ai-provider="…"` (`--ai`)       | AI provider name (gemini/openai)                                                                 |                 `gemini`                  |     `AI_PROVIDER`     |
| `--gemini-api-key="…"` (`--ga`)    | Gemini API key (https://bit.ly/4jZhiKI, as of February 2025 it's free)                           |                                           |   `GEMINI_API_KEY`    |
| `--gemini-model-name="…"` (`--gm`) | Gemini model name (https://bit.ly/4i02ARR)                                                       |            `gemini-2.0-flash`             |  `GEMINI_MODEL_NAME`  |
| `--openai-api-key="…"` (`--oa`)    | OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account to access the API) |                                           |   `OPENAI_API_KEY`    |
| `--openai-model-name="…"` (`--om`) | OpenAI model name (https://bit.ly/4hXCXkL)                                                       |               `gpt-4o-mini`               |  `OPENAI_MODEL_NAME`  |

<!--/GENERATED:CLI_DOCS-->
