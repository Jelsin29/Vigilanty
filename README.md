# Vigilanty — Configurable verification pipeline for git commits

![Version](https://img.shields.io/badge/version-v0.1.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Go Version](https://img.shields.io/badge/go-1.22.2-00ADD8?logo=go)
![Platforms](https://img.shields.io/badge/platforms-linux%20%7C%20macOS%20%7C%20windows-lightgrey)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)

Vigilanty runs a configurable pre-commit verification pipeline for staged changes, combining shell-based checks and AI review in a single CLI.

## Why

Modern commits often pass through too many disconnected tools:

| Problem | What Vigilanty does |
| --- | --- |
| Commits go unchecked until CI | Runs checks before the commit is created |
| AI-generated code still needs verification | Adds AI review as a pipeline step, not a separate manual ritual |
| Linters, builds, tests, and review live in different tools | Centralizes them in one ordered pipeline |
| Teams want repo-local rules | Stores the workflow in `.vigilanty.yml` |

## How it works

```text
staged git changes
        |
        v
  vigilanty run
        |
        v
+-------------------------------+
| verification pipeline         |
|-------------------------------|
| 1. lint                       |
| 2. build                      |
| 3. test                       |
| 4. AI review                  |
+-------------------------------+
        |
        +--> all pass  -> commit continues
        |
        +--> one fails -> commit blocked
```

Vigilanty reads the staged diff, executes steps sequentially, and returns a non-zero exit code when the pipeline fails.

## Installation

### Go install

```bash
go install github.com/jelsin29/vigilanty@v0.1.0
```

### Homebrew

Coming soon.

### Build from source

```bash
git clone https://github.com/jelsin29/vigilanty.git vigilanty
cd vigilanty
make build
./vigilanty version
```

## Quick Start

```bash
# 1) Enter your repository
cd /path/to/your/repo

# 2) Scaffold a config
vigilanty init --preset go

# 3) Edit the generated file
$EDITOR .vigilanty.yml

# 4) Install the git pre-commit hook
vigilanty install

# 5) You're done
# Every commit now runs the pipeline automatically
```

## Configuration

Example `.vigilanty.yml`:

```yaml
version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

pipeline:
  - name: lint
    checker: shell
    command: "golangci-lint run ./..."
    timeout: "120s"

  - name: build
    checker: shell
    command: "go build ./..."

  - name: test
    checker: shell
    command: "go test ./..."

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Go diff for bugs, security issues, performance regressions, and maintainability concerns. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
```

### Top-level fields

| Field | Required | Description |
| --- | --- | --- |
| `version` | No | Config schema version. Current supported value is `"1"`. |
| `global.fail_fast` | No | Stops at the first failed or errored step. |
| `global.diff_max_bytes` | No | Maximum staged diff size read before truncation. Default: `262144`. |
| `global.timeout` | No | Default timeout applied to steps that do not define their own timeout. Default: `2m`. |
| `global.verbose` | No | Prints all checker output, not only failures. |
| `pipeline` | Yes | Ordered list of steps to run. |

### Pipeline step fields

| Field | Required | Description |
| --- | --- | --- |
| `name` | Yes | Unique step name shown in output. |
| `checker` | Yes | Checker type. Common values: `shell`, `ai-review`. |
| `command` | For `shell` | Command executed in the repository root. |
| `provider` | For `ai-review` | AI CLI to use: `claude`, `gemini`, or `ollama`. |
| `prompt` | For `ai-review` | Review instructions sent together with the staged diff. |
| `model` | For `ollama` | Model name required by the Ollama provider. |
| `timeout` | No | Step-specific timeout such as `60s` or `2m`. |
| `enabled` | No | Set to `false` to skip a step without deleting it. |
| `env` | No | Environment variables for shell-based steps. |
| `skip_on_empty_diff` | No | Skips AI review when nothing is staged. |
| `max_diff_lines` | No | Limits how many diff lines are sent to the AI reviewer. |
| `pass_pattern` | No | Regex used to detect a passing AI response. |
| `fail_pattern` | No | Regex used to detect a failing AI response. |
| `config` | No | Extra checker-specific options. |

## Presets

Use `vigilanty init --preset <name>` to scaffold a starting config.

| Preset | Command | Includes |
| --- | --- | --- |
| Go | `vigilanty init --preset go` | `golangci-lint`, `go build`, `go test`, AI review |
| Node | `vigilanty init --preset node` | `npm run lint`, `npm run typecheck`, `npm test`, AI review |
| Python | `vigilanty init --preset python` | `ruff`, `pytest`, AI review |
| Generic | `vigilanty init --preset generic` | AI review with commented shell examples |

## Providers

Supported AI CLI providers:

| Provider | Install | Notes |
| --- | --- | --- |
| `claude` | https://docs.anthropic.com/en/docs/claude-code | Uses the local Claude CLI |
| `gemini` | https://github.com/google-gemini/gemini-cli | Uses the Gemini CLI with prompt input |
| `ollama` | https://ollama.ai/download | Requires `model` in config |

If the selected CLI is not installed, Vigilanty fails the step and prints the provider install URL.

## Commands

| Command | Description |
| --- | --- |
| `vigilanty init` | Create `.vigilanty.yml` in the current repository |
| `vigilanty init --preset go` | Create config from a Go-focused preset |
| `vigilanty install` | Install the `pre-commit` git hook |
| `vigilanty uninstall` | Remove the installed git hook |
| `vigilanty run` | Execute the verification pipeline for staged changes |
| `vigilanty version` | Print the current version |
| `vigilanty --config /path/to/config.yml run` | Use an explicit config file |
| `vigilanty --verbose run` | Print verbose checker output |

## Pipeline behavior

| Behavior | Details |
| --- | --- |
| Execution model | Steps run sequentially in the order defined in `pipeline` |
| Fail-fast | When `global.fail_fast: true`, remaining steps are marked as skipped after the first failure or error |
| Git input | The pipeline evaluates staged files and the staged diff |
| Empty pipeline | Prints a warning and exits successfully |
| AI diff handling | Diff bytes are capped globally and AI review lines can be capped per step |

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Pipeline passed |
| `1` | Checker failure or general command failure |
| `2` | Configuration error |
| `3` | Git error |
| `4` | Internal error |

## Comparison with GGA

| Aspect | Vigilanty | GGA |
| --- | --- | --- |
| Core model | Multi-step verification pipeline | Single AI-oriented review flow |
| Implementation | Go CLI | Bash-based tooling |
| Configuration | YAML file (`.vigilanty.yml`) | Sourced shell configuration |
| Review input | Staged git diffs with truncation controls | More full-file oriented workflows |
| Use case | Block bad commits with ordered checks | Fast AI review utility |

## Contributing

PRs are welcome.

- Open an issue: https://github.com/jelsin29/vigilanty/issues
- Discuss a change before large work
- Keep contributions focused and documented

## License

MIT — see [LICENSE](./LICENSE).
