# AGENTS.md

> Read this file AND the global rules before making any code changes -
> https://tarampampam.github.io/.github/ai/AGENTS.md (mirror -
> <https://raw.githubusercontent.com/tarampampam/.github/refs/heads/master/ai/AGENTS.md>).

## Instruction Priority

1. This file (`AGENTS.md` in this repository)
2. Global rules (external URLs)
3. Other documentation

If rules conflict, follow the highest priority source.

## Project Overview

`describe-commit` is a CLI tool that uses AI to generate Git commit messages from staged changes. It fetches
`git diff` and optionally `git log`, then sends them to a configured AI provider.

## Key Commands

```bash
# Build
go build ./cmd/describe-commit/

# Run tests
go test ./...

# Run a single test
go test ./internal/ai/... -run TestFunctionName -v

# Lint
golangci-lint run ./...

# Update README CLI docs (regenerates the <!--GENERATED:APP_README--> section)
go generate -skip readme ./...
# To also regenerate the readme section:
go generate ./...
```

> **Before making Go changes**: read `.golangci.yml` - it defines all active linters and their configuration.

Before submitting changes:

1. Regenerate code if needed
2. Run linters
3. Run tests

## Architecture

```
cmd/describe-commit/main.go       # entry point - signal handling, calls cli.NewApp
internal/
├── cli/
│   ├── app.go                    # main orchestrator: flags, config loading, provider selection,
│   │                             #   concurrent git fetch, calls provider.Query
│   ├── cmd/
│   │   ├── command.go            # custom CLI framework (no deps): flag parsing, help, version
│   │   └── flag.go               # generic Flag[T] with env var support, validation, IsSet()
│   └── generate/
│       └── readme.go             # go:generate target (build tag "readme"): rewrites
│                                 #   <!--GENERATED:APP_README--> block in README.md
├── ai/
│   ├── providers.go              # Provider interface, provider name constants, SupportedProviders()
│   ├── prompt.go                 # GeneratePrompt(): builds the system prompt for all providers
│   ├── options.go                # shared query options (ShortMessageOnly, EnableEmoji, MaxOutputTokens)
│   ├── gemini.go                 # Gemini HTTP implementation
│   ├── openai.go                 # OpenAI HTTP implementation
│   ├── openrouter.go             # OpenRouter HTTP implementation
│   └── anthropic.go              # Anthropic HTTP implementation
├── config/
│   ├── path.go                   # DefaultDirPath(), FindIn() - walks dir tree for config files
│   ├── path_linux.go             # OS-specific config dir (~/.config/)
│   ├── path_darwin.go            # OS-specific config dir (~/Library/Application Support/)
│   ├── path_windows.go           # OS-specific config dir (%APPDATA%)
│   └── config_test.go            # covers FromFile() and multi-file merge behaviour
├── git/
│   ├── diff.go                   # thin wrapper around `git diff`
│   └── log.go                    # thin wrapper around `git log`
├── errgroup/
│   └── errgroup.go               # context-aware errgroup used for concurrent git calls
├── version/
│   └── version.go                # version string injected at build time via -ldflags
├── debug/
│   └── debug.go                  # debug Printf (enabled by DEBUG env var)
└── yaml/                         # vendored YAML parser - excluded from linting
```

## Key Design Patterns

**Config merging**: `options.UpdateFromConfigFile()` in `app.go` accepts a slice of config paths (global user config +
any found in the working directory hierarchy). Each file only overrides fields that are explicitly set
(pointer-based: `nil` means "not set"). Config priority: CLI flags > env vars > per-project config files > global
user config.
**Flag tracking**: `Flag[T].IsSet()` distinguishes an explicit flag from its default value, enabling the "override
only what was explicitly set" pattern.
**Concurrent git operations**: `git diff` and `git log` run concurrently via `internal/errgroup`.
**README generation**: `go generate ./...` in `internal/cli/` runs `generate/readme.go` (build tag `readme`) which
replaces the `<!--GENERATED:APP_README-->...<!--/GENERATED:APP_README-->` section in README.md with the live CLI
help text. Run this when changing flags.

## Adding a New AI Provider

1. Add a constant in `internal/ai/providers.go` and add it to `SupportedProviders()`.
2. Implement the `Provider` interface in a new file `internal/ai/<name>.go`.
3. Add the provider struct and options to `internal/config/` (following the existing pattern).
4. Add corresponding `cmd.Flag` entries in `internal/cli/app.go` and wire them in the `switch` block.

## Changes That Require Confirmation

Ask before:

- Modifying `api/*` files
- Changing storage interfaces or implementations
- Introducing new external dependencies
- Changing public HTTP APIs
- Refactoring large parts of the codebase

## Linting Notes

- Line length limit: 120 characters.
- Import grouping enforced by `gci`: stdlib → external → `gh.tarampamp.am/describe-commit`.
- `fmt.Print*` and bare `print*` are forbidden (use `fmt.Fprint*` with an explicit writer).
- `internal/yaml/` is excluded from all linting.
- Test files are exempt from `dupl`, `funlen`, `gocognit`, `goconst`, `lll`, `nlreturn`.

## Offline Fallback Rules

> Apply these only if the external rule URLs above are inaccessible. The external rules are authoritative.

### Go

- Wrap errors with context: `fmt.Errorf("operation: %w", err)`. Return sentinel errors directly when they are unlikely.
- Use `xErr` naming when multiple errors are in scope (e.g. `readErr`, `writeErr`); use `if err := ...; err != nil`
  for single short-lived errors.
- Interfaces in the consumer package; keep them minimal; add `var _ Interface = (*Impl)(nil)` compile-time assertions.
- Exported declarations must have a doc comment starting with the identifier name, ending with a period.
- No `fmt.Print*` / `print` / `println`; no global variables; no `init()` without justification.
- Line length ≤ 120 characters.
- Test files: `package foo_test` (external); one `_test.go` per tested file; both outer and inner `t.Parallel()`;
  map-based table-driven tests with `give*` / `want*` keys.
