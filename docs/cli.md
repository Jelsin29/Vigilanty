# Vigilanty CLI Reference

## Global flags

| Flag | Description |
| --- | --- |
| `--config <path>` | Use an explicit config file instead of `.vigilanty.yml` or the global config |
| `--verbose` | Print verbose checker output |

## Commands

### `vigilanty init`

Create `.vigilanty.yml` in the current repository.

Flags:

| Flag | Description |
| --- | --- |
| `--preset <name>` | One of `go`, `node`, `typescript`, `python`, `rust`, `java`, `dotnet`, `ruby`, `swift`, `php`, `generic` |
| `--no-interactive` | Skip the wizard and generate the preset directly |
| `--force` | Overwrite `.vigilanty.yml` without prompting |

Behavior:

- Auto-detects project type when `--preset` is omitted
- Prints a preset preview before writing the config
- Prints a **Next Steps** section after creation

### `vigilanty run`

Run the verification pipeline.

Flags:

| Flag | Description |
| --- | --- |
| `--no-cache` | Bypass checker cache for the current run |
| `--pr-mode` | Review the full PR diff instead of staged changes |
| `--base <branch>` | Base branch to compare against in PR mode |
| `--ci` | Review the last commit diff for CI/CD flows |
| `--json` | Print stable machine-readable JSON to stdout |

Modes:

- Default: staged diff
- `--pr-mode`: PR diff vs base branch
- `--ci`: last commit diff
- `--json`: stable JSON payload, stdout reserved for JSON only

### `vigilanty install`

Install the `pre-commit` git hook.

### `vigilanty uninstall`

Remove the installed `pre-commit` hook.

### `vigilanty cache status`

Show cache location, number of cache entries, and total size.

### `vigilanty cache clear`

Delete all Vigilanty cache files.

### `vigilanty config`

Print the active config file source, global settings, and pipeline steps.

### `vigilanty version`

Print the current Vigilanty version.

## Providers

| Provider | Model required? | Notes |
| --- | --- | --- |
| `claude` | No | Anthropic Claude CLI |
| `gemini` | No | Gemini CLI |
| `codex` | No | OpenAI Codex CLI |
| `opencode` | Optional | OpenCode CLI |
| `ollama` | Yes | Local Ollama runtime |
| `lmstudio` | Optional | Local LM Studio runtime |
| `github` | Optional | GitHub CLI / Copilot-backed workflows |

Environment requirements depend on the provider CLI you install. For hosted providers, authenticate the official CLI first.

## Presets

| Preset | Pipeline skeleton |
| --- | --- |
| `go` | `golangci-lint`, `go build`, `go test`, `ai-review` |
| `node` | `eslint`, `tsc`, `npm-test`, `ai-review` |
| `typescript` | `tsc`, `eslint`, `test`, `ai-review` |
| `python` | `ruff`, `pytest`, `ai-review` |
| `rust` | `cargo-clippy`, `cargo-build`, `cargo-test`, `ai-review` |
| `java` | `build`, `test`, `ai-review` |
| `dotnet` | `dotnet-build`, `dotnet-test`, `ai-review` |
| `ruby` | `rubocop`, `rspec`, `ai-review` |
| `swift` | `swift-build`, `swift-test`, `ai-review` |
| `php` | `phpstan`, `phpunit`, `ai-review` |
| `generic` | `ai-review` starter plus commented examples |

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Passed |
| `1` | Checker failure |
| `2` | Config error |
| `3` | Git error |
| `4` | Internal error |
