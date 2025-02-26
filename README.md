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

- [OpenAI ChatGPT](https://openai.com/chatgpt/overview/)
- [Google Gemini](https://deepmind.google/technologies/gemini/)

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
- Takes the commit history into account for better context
- Runs as a standalone binary (only installed `git` is required)
- Available for **Linux**, **macOS**, **Windows**, and as a **Docker image**

> [!NOTE]
> Under the hood, this app does two things before returning the generated commit message:
>
> - Retrieves the `git diff` (and `git log` optionally) for the specified directory
> - Sends this diff to the AI provider with the provided special prompt
>
> Please keep in mind that when working with proprietary code, some parts of the code will be sent to the AI
> provider. You should ensure that this is permitted by your company's policy. Additionally, make sure that
> the AI provider does not store your data (or stores it securely).
>
> The author of this tool is not responsible for any data leaks or security issues.

## 🧩 Installation

### 📦 Debian/Ubuntu-based (.deb) systems

Execute the following commands in order:

```shell
# setup the repository automatically
curl -1sLf https://dl.cloudsmith.io/public/tarampampam/describe-commit/setup.deb.sh | sudo -E bash

# install the package
sudo apt install describe-commit
```

<details>
  <summary>Uninstalling</summary>

```shell
sudo apt remove describe-commit
rm /etc/apt/sources.list.d/tarampampam-describe-commit.list
```

</details>

### 📦 RedHat (.rpm) systems

```shell
# setup the repository automatically
curl -1sLf https://dl.cloudsmith.io/public/tarampampam/describe-commit/setup.rpm.sh | sudo -E bash

# install the package
sudo dnf install describe-commit # RedHat, CentOS, etc.
sudo yum install describe-commit # Fedora, etc.
sudo zypper install describe-commit # OpenSUSE, etc.
```

<details>
  <summary>Uninstalling</summary>

```shell
# RedHat, CentOS, Fedora, etc.
sudo dnf remove describe-commit
rm /etc/yum.repos.d/tarampampam-describe-commit.repo
rm /etc/yum.repos.d/tarampampam-describe-commit-source.repo

# OpenSUSE, etc.
sudo zypper remove describe-commit
zypper rr tarampampam-describe-commit
zypper rr tarampampam-describe-commit-source
```

</details>

### 📦 Alpine Linux

```shell
# bash is required for the setup script
sudo apk add --no-cache bash

# setup the repository automatically
curl -1sLf https://dl.cloudsmith.io/public/tarampampam/describe-commit/setup.alpine.sh | sudo -E bash

# install the package
sudo apk add describe-commit
```

<details>
  <summary>Uninstalling</summary>

```shell
sudo apk del describe-commit
$EDITOR /etc/apk/repositories # remove the line with the repository
```

</details>

### 📦 AUR (Arch Linux)

There are three packages available in the AUR:

- Build from source: [describe-commit](https://aur.archlinux.org/packages/describe-commit)
- Precompiled: [describe-commit-bin](https://aur.archlinux.org/packages/describe-commit-bin)
- Unstable: [describe-commit-git](https://aur.archlinux.org/packages/describe-commit-git)

```shell
pamac build describe-commit
```

<details>
  <summary>Uninstalling</summary>

```shell
pacman -Rs describe-commit
```

</details>

### 📦 Binary (Linux, macOS, Windows)

Download the latest binary for your architecture/OS from the [releases page][link_releases]. For example, to install
the latest version to the `/usr/local/bin` directory on an **amd64** system (e.g., Debian, Ubuntu), you can run:

```shell
# download and install the binary
curl -SsL \
  https://github.com/tarampampam/describe-commit/releases/latest/download/describe-commit-linux-amd64.gz | \
  gunzip -c | sudo tee /usr/local/bin/describe-commit > /dev/null

# make the binary executable
sudo chmod +x /usr/local/bin/describe-commit
```

<details>
  <summary>Uninstalling</summary>

```shell
sudo rm /usr/local/bin/describe-commit
```

</details>

> [!TIP]
> Each release includes binaries for **linux**, **darwin** (macOS) and **windows** (`amd64` and `arm64` architectures).
> You can download the binary for your system from the [releases page][link_releases] (section `Assets`). And - yes,
> all what you need is just download and run single binary file.

[link_releases]:https://github.com/tarampampam/describe-commit/releases

### 📦 Docker image

Also, you can use the Docker image:

| Registry                               | Image                                 |
|----------------------------------------|---------------------------------------|
| [GitHub Container Registry][link_ghcr] | `ghcr.io/tarampampam/describe-commit` |

> [!NOTE]
> It’s recommended to avoid using the `latest` tag, as **major** upgrades may include breaking changes.
> Instead, use specific tags in `:X.Y.Z` or only `:X` format for version consistency.

[link_ghcr]:https://github.com/tarampampam/describe-commit/pkgs/container/describe-commit

<details>
  <summary>Example</summary>

> ```shell
> docker run --rm \
>   -u "$(id -u):$(id -g)" \                                # to avoid problems with permissions
>   -v "$HOME/.config/describe-commit.yml:/config.yml:ro" \ # use your configuration file
>   -v "$(pwd):/rootfs:ro" \                                # mount current directory as read-only
>   -e "CONFIG_FILE=/config.yml" \                          # specify the configuration file path
>   -w "/rootfs" \                                          # set the working directory
>     ghcr.io/tarampampam/describe-commit ...
> ```

</details>

## ⚙ Configuration

You can configure `describe-commit` using a YAML file. Refer to [this example](describe-commit.example.yml) for
available options.

You can specify the configuration file's location using the `--config-file` option. By default, however, the
tool searches for the file in the user's configuration directory:

- **Linux**: `~/.configs/describe-commit.yml`
- **Windows**: `%APPDATA%\describe-commit.yml`
- **macOS**: `~/Library/Application Support/describe-commit.yml`

### Configuration Options Priority

Configuration options are applied in the following order, from highest to lowest priority:

1. Command-line options (e.g., `--ai-provider`, `--openai-api-key`, etc.)
2. Environment variables (e.g., `GEMINI_API_KEY`, `OPENAI_MODEL_NAME`, etc.)
3. A configuration file in the working directory or any parent directory, up to the root (the file can be
   named `.describe-commit.yml` or `describe-commit.yml`)
4. A configuration file in the user's configuration directory (e.g., `~/.configs/describe-commit.yml` for Linux)

This means you can store API tokens and other default settings in the global user's configuration file and override
them with command-line options or a configuration file in the working directory when needed (e.g., enabling emojis
only for specific projects, disable commits history analysis, etc.).

## 🚀 Use Cases (usage examples)

#### ☝ Commit the changes using an AI-generated commit message in a single command

```shell
git commit -m "$(describe-commit)"
```

> A Git repository must be initialized in the specified directory, and `git` must be installed on your system.
> Additionally, ensure that changes are staged (`git add -A`) before running the tool.

#### ☝ Integration with the git

Add this alias to your `~/.gitconfig` file:

```ini
[alias]
  # Stage all changes and commit them with a generated message
  wip = "!f() { git add -Av && git commit -m \"$(describe-commit)\"; }; f"
```

Now, in **any** repository, you can simply run:

```shell
git wip
```

And voilà! All changes will be staged and committed with a generated message.

<details>
  <summary><strong>☝ Get a Commit Message for a Specific Directory</strong></summary>

```shell
describe-commit /path/to/repo
```

Example output:

```markdown
docs: Add initial README with project description

This commit introduces the initial README file, providing a comprehensive
overview of the `describe-commit` project. It includes a project description,
features list, installation instructions, and usage examples.

- Provides a clear introduction to the project
- Guides users through installation and basic usage
- Highlights key features and functionalities
```

You are able to save the output to a file:

```shell
describe-commit /path/to/repo > /path/to/commit-message.txt
```

Or do wherever you want with it.

</details>

<details>
  <summary><strong>☝ Switch Between AI Providers</strong></summary>

Generate a commit message using OpenAI:

```shell
describe-commit --ai openai --openai-api-key "your-openai-api-key"
```

Will output something like this:

```markdown
docs(README): Update project description and installation instructions

Enhanced the README file to provide a clearer project overview and detailed installation instructions. The
changes aim to improve user understanding and accessibility of the `describe-commit` CLI tool.

- Added project description and AI provider support
- Included features list for better visibility
- Updated installation instructions with binary and Docker options
- Provided usage examples for generating commit messages
```

But if you want to use Gemini instead:

```shell
describe-commit --ai gemini --gemini-api-key "your-gemini-api-key"
```

</details>

<details>
  <summary><strong>☝ Generate a short commit message (only the first line) with emojis</strong></summary>

```shell
describe-commit -s -e
```

Will give you something like this:

```markdown
📝 docs(README): Update project description and installation instructions
```

</details>

<!--GENERATED:APP_README-->
## 💻 Command line interface

```
Description:
   This tool leverages AI to generate commit messages based on changes made in a Git repository.

Usage:
   describe-commit [<options>] [<git-dir-path>]

Version:
   0.0.0@undefined

Options:
   --config-file="…", -c="…"                        Path to the configuration file (default: depends/on/your-os/describe-commit.yml) [$CONFIG_FILE]
   --short-message-only, -s                         Generate a short commit message (subject line) only [$SHORT_MESSAGE_ONLY]
   --commit-history-length="…", --cl="…", --hl="…"  Number of previous commits from the Git history (0 = disabled) (default: 20) [$COMMIT_HISTORY_LENGTH]
   --enable-emoji, -e                               Enable emoji in the commit message [$ENABLE_EMOJI]
   --max-output-tokens="…"                          Maximum number of tokens in the output message (default: 500) [$MAX_OUTPUT_TOKENS]
   --ai-provider="…", --ai="…"                      AI provider name (gemini|openai) (default: gemini) [$AI_PROVIDER]
   --gemini-api-key="…", --ga="…"                   Gemini API key (https://bit.ly/4jZhiKI, as of February 2025 it's free) [$GEMINI_API_KEY]
   --gemini-model-name="…", --gm="…"                Gemini model name (https://bit.ly/4i02ARR) (default: gemini-2.0-flash) [$GEMINI_MODEL_NAME]
   --openai-api-key="…", --oa="…"                   OpenAI API key (https://bit.ly/4i03NbR, you need to add funds to your account) [$OPENAI_API_KEY]
   --openai-model-name="…", --om="…"                OpenAI model name (https://bit.ly/4hXCXkL) (default: gpt-4o-mini) [$OPENAI_MODEL_NAME]
   --help, -h                                       Show help
   --version, -v                                    Print the version
```
<!--/GENERATED:APP_README-->

## 📜 License

This is open-sourced software licensed under the [MIT License][link_license].

[link_license]:https://github.com/tarampampam/describe-commit/blob/master/LICENSE
