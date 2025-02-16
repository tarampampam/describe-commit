<p align="center">
  <a href="https://github.com/tarampampam/describe-commit#readme">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://socialify.git.ci/tarampampam/describe-commit/image?description=1&font=Raleway&forks=1&issues=1&logo=https%3A%2F%2Fhabrastorage.org%2Fwebt%2F6v%2F7-%2Fil%2F6v7-iljmr7fo1ogj3uz0kvnkhaa.png&owner=1&pulls=1&pattern=Solid&stargazers=1&theme=Dark">
      <img align="center" src="https://socialify.git.ci/tarampampam/describe-commit/image?description=1&font=Raleway&forks=1&issues=1&logo=https%3A%2F%2Fhabrastorage.org%2Fwebt%2F6v%2F7-%2Fil%2F6v7-iljmr7fo1ogj3uz0kvnkhaa.png&owner=1&pulls=1&pattern=Solid&stargazers=1&theme=Light">
    </picture>
  </a>
</p>

# Describe Commit

`describe-commit` is a CLI tool that leverages AI to generate commit messages based on changes made in a Git repository.
Currently, it supports the following AI providers:

- OpenAI
- Google Gemini

It also allows users to select the desired model for content generating.

## 🦾 tl;dr

|       Turn this       |      Into this      |
|:---------------------:|:-------------------:|
| ![Before][before_img] | ![After][after_img] |

[before_img]:https://habrastorage.org/webt/ni/z-/d1/niz-d1zfmbf4pc9kpdbvtapvljy.png
[after_img]:https://habrastorage.org/webt/wo/8t/cd/wo8tcdgdm80fb6ayacgpse3_8hu.png

Without any manual effort (there's no time to write commit messages, lazy developers unite)!

## 🔥 Features list

- Generates meaningful commit messages using AI
- Supports different AI providers
- Can generate short commit messages (subject line only)
- Optionally includes emojis (🐛✨📝🚀✅♻️⬆️🔧🌐💡) in commit messages
- Runs as a standalone binary (no external dependencies)
- Available as a **Docker image**

> [!NOTE]
> Under the hood, this app does two things before returning the generated commit message:
>
> - Retrieves the `git diff` for the specified directory
> - Sends this diff to the AI provider with the provided special prompt
>
> Please keep in mind that when working with proprietary code, some parts of the code will be sent to the AI
> provider. You should ensure that this is permitted by your company's policy. Additionally, make sure that
> the AI provider does not store your data (or stores it securely).
>
> The author of this tool is not responsible for any data leaks or security issues.

## 🧩 Installation

Download the latest binary for your architecture from the [releases page][link_releases]. For example, to install
on an **amd64** system (e.g., Debian, Ubuntu):

```shell
curl -SsL -o ./describe-commit https://github.com/tarampampam/describe-commit/releases/latest/download/describe-commit-linux-amd64
chmod +x ./describe-commit
./describe-commit --help
```

> [!TIP]
> Each release includes binaries for **linux**, **darwin** (macOS) and **windows** (`amd64` and `arm64` architectures).
> You can download the binary for your system from the [releases page][link_releases] (section `Assets`). And - yes,
> all what you need is just download and run single binary file.

[link_releases]:https://github.com/tarampampam/describe-commit/releases

Alternatively, you can use the Docker image:

| Registry                               | Image                                 |
|----------------------------------------|---------------------------------------|
| [GitHub Container Registry][link_ghcr] | `ghcr.io/tarampampam/describe-commit` |

[link_ghcr]:https://github.com/tarampampam/describe-commit/pkgs/container/describe-commit

> ```shell
> docker run --rm \
>   -u "$(id -u):$(id -g)" \                                # to avoid problems with permissions
>   -v "$HOME/.config/describe-commit.yml:/config.yml:ro" \ # use your configuration file
>   -v "$(pwd):/rootfs:ro" \                                # mount current directory as read-only
>   -e "CONFIG_FILE=/config.yml" \                          # specify the configuration file path
>   -w "/rootfs" \                                          # set the working directory
>     ghcr.io/tarampampam/describe-commit ...
> ```

> [!NOTE]
> It’s recommended to avoid using the `latest` tag, as **major** upgrades may include breaking changes.
> Instead, use specific tags in `X.Y.Z` format for version consistency.

## 🚀 Usage

To generate a commit message for the current Git repository, and commit the changes in a single command:

```shell
git commit -m "$(/path/to/describe-commit)"
```

> [!NOTE]
> A Git repository must be initialized in the specified directory, and `git` must be installed on your system.
> Additionally, ensure that changes are staged (`git add -A`) before running the tool.

Or just to get a commit message for a specific directory:

```shell
describe-commit /path/to/repo
```

Will output something like:

```markdown
docs: Add initial README with project description

This commit introduces the initial README file, providing a comprehensive
overview of the `describe-commit` project. It includes a project description,
features list, installation instructions, and usage examples.

- Provides a clear introduction to the project
- Guides users through installation and basic usage
- Highlights key features and functionalities
```

Generate a commit message with OpenAI:

```sh
describe-commit --ai openai --openai-api-key "your-api-key"
```

```markdown
docs(README): Update project description and installation instructions

Enhanced the README file to provide a clearer project overview and detailed installation instructions. The changes aim to improve user understanding and accessibility of the `describe-commit` CLI tool.

- Added project description and AI provider support
- Included features list for better visibility
- Updated installation instructions with binary and Docker options
- Provided usage examples for generating commit messages
```

Generate a short commit message (only the first line) with emojis:

```sh
describe-commit -s -e
```

```markdown
📝 docs(README): Update project description and installation instructions
```

## ⚙ Configuration

You can configure `describe-commit` using a YAML file. Use [describe-commit.yml](describe-commit.yml) as a reference
for the available options.

The configuration file's location can be specified using the `--config-file` option. However, by default, the file
is searched for in the user's configuration directory:

- **Linux**: `~/.configs/describe-commit.yml`
- **Windows**: `%APPDATA%\describe-commit.yml`
- **macOS**: `~/Library/Application Support/describe-commit.yml`

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

| Name                                           | Description                                                                                            |               Default value               |  Environment variables  |
|------------------------------------------------|--------------------------------------------------------------------------------------------------------|:-----------------------------------------:|:-----------------------:|
| `--config-file="…"` (`-c`)                     | path to the configuration file                                                                         | `/depends/on/your-os/describe-commit.yml` |      `CONFIG_FILE`      |
| `--short-message-only` (`-s`)                  | generate a short commit message (subject line) only                                                    |                  `false`                  |  `SHORT_MESSAGE_ONLY`   |
| `--commit-history-length="…"` (`--cl`, `--hl`) | number of previous commits from the Git history to consider as context for the AI model (0 = disabled) |                   `20`                    | `COMMIT_HISTORY_LENGTH` |
| `--enable-emoji` (`-e`)                        | enable emoji in the commit message                                                                     |                  `false`                  |     `ENABLE_EMOJI`      |
| `--max-output-tokens="…"`                      | maximum number of tokens in the output message                                                         |                   `500`                   |   `MAX_OUTPUT_TOKENS`   |
| `--ai-provider="…"` (`--ai`)                   | AI provider name (gemini/openai)                                                                       |                 `gemini`                  |      `AI_PROVIDER`      |
| `--gemini-api-key="…"` (`--ga`)                | Gemini API key (https://bit.ly/4jZhiKI, as of February 2025 it's free)                                 |                                           |    `GEMINI_API_KEY`     |
| `--gemini-model-name="…"` (`--gm`)             | Gemini model name (https://bit.ly/4i02ARR)                                                             |            `gemini-2.0-flash`             |   `GEMINI_MODEL_NAME`   |
| `--openai-api-key="…"` (`--oa`)                | OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account to access the API)       |                                           |    `OPENAI_API_KEY`     |
| `--openai-model-name="…"` (`--om`)             | OpenAI model name (https://bit.ly/4hXCXkL)                                                             |               `gpt-4o-mini`               |   `OPENAI_MODEL_NAME`   |

<!--/GENERATED:CLI_DOCS-->

## License

This is open-sourced software licensed under the [MIT License][link_license].

[link_license]:https://github.com/tarampampam/describe-commit/blob/master/LICENSE
