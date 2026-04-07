<p align="center">
  <img width="1773" height="886" alt="Image" src="https://github.com/user-attachments/assets/df88e02c-930d-4123-bfa0-fb757ebac293" />
</p>

<p align="center">
  <img src="https://img.shields.io/badge/version-v0.2.0-blue" alt="Version" />
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License" />
  <img src="https://img.shields.io/badge/go-1.22.2-00ADD8?logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/platforms-linux%20%7C%20macOS%20%7C%20windows-lightgrey" alt="Platforms" />
  <img src="https://img.shields.io/badge/CI-JSON%20ready-brightgreen" alt="CI JSON Ready" />
</p>

# Vigilanty

**Vigilanty is a pre-commit verification hub**: one CLI that runs shell checks, staged-diff analysis, and AI review before bad changes leave your machine or your CI pipeline.

## Why Vigilanty

Most teams still glue verification together with separate linters, test runners, hooks, CI jobs, and ad-hoc AI prompts.

Vigilanty gives you one pipeline for all of that:

- **Shell checks** for lint, build, test, security, formatting, or custom scripts
- **AI review** as a first-class pipeline step, not a separate ritual
- **Repo-local config** in `.vigilanty.yml`
- **Project auto-detection** during `vigilanty init`
- **Human-friendly local output** by default
- **Stable `--json` output** for CI and automation

## Differential Positioning

Vigilanty is strongest when you need a tool that is:

| Capability | Vigilanty stance |
| --- | --- |
| Verification workflow | A **hub** that orchestrates multiple checks in order |
| AI integration | **Provider-agnostic bridge** across Claude, Gemini, Codex, OpenCode, Ollama, LM Studio, and GitHub |
| Setup | **Zero-friction onboarding** with presets and project detection |
| Git scope | Supports **staged**, **PR**, and **CI last-commit** review modes |
| Automation | Explicit **machine-readable JSON contract** with unchanged human default output |

## How it works

```text
staged changes / PR diff / CI commit diff
                  |
                  v
            vigilanty run
                  |
                  v
      +---------------------------+
      | verification pipeline     |
      |---------------------------|
      | 1. lint / format          |
      | 2. build / typecheck      |
      | 3. test / security        |
      | 4. AI review              |
      +---------------------------+
                  |
          +-------+-------+
          |               |
          v               v
       pass           fail / error
   exit code 0        exit code 1
```

## Installation

### Homebrew

```bash
brew install Jelsin29/tap/vigilanty
```

### Go install

```bash
go install github.com/jelsin29/vigilanty@v0.2.0
```

### Linux quick install

```bash
curl -fsSL https://raw.githubusercontent.com/Jelsin29/Vigilanty/main/scripts/autoinstall.sh | bash
```

Specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/Jelsin29/Vigilanty/main/scripts/autoinstall.sh | VIGILANTY_VERSION=v0.2.0 bash
```

### From source

```bash
git clone https://github.com/Jelsin29/Vigilanty.git vigilanty
cd vigilanty
make build
./vigilanty version
```

## Quick start

```bash
cd /path/to/repo
vigilanty init
vigilanty install
vigilanty run
```

For CI or scripts:

```bash
vigilanty run --json
```

## Interactive setup

`vigilanty init` can auto-detect the project type, walk you through file scope, help choose an AI provider, and optionally generate `AGENTS.md` so your review rules start with real structure instead of an empty file.

<p align="center">
  <img width="1400" alt="Vigilanty interactive provider selection" src="https://github.com/user-attachments/assets/dbb63929-bc77-4b33-a476-e67f8c58f359" />
</p>

Want the full setup flow? See [`docs/providers.md`](docs/providers.md), [`docs/configuration.md`](docs/configuration.md), and [`docs/rules-file.md`](docs/rules-file.md).

## Example output

Default runs stay optimized for humans: you get a clear pipeline summary, step-by-step progress, and a direct pass/fail outcome without giving up the option to switch to JSON for CI.

<p align="center">
  <img width="404" height="1016" alt="Vigilanty successful run output" src="https://github.com/user-attachments/assets/fea85988-e32b-4aef-b133-bd7178f84249" />
</p>

For more real-world samples, including failure cases and JSON payloads, see [`docs/examples.md`](docs/examples.md) and [`docs/run-json.md`](docs/run-json.md).

## Commands

See the full command reference in [`docs/cli.md`](docs/cli.md).

Core commands:

- `vigilanty init`
- `vigilanty run`
- `vigilanty install`
- `vigilanty uninstall`
- `vigilanty cache status`
- `vigilanty cache clear`
- `vigilanty config`
- `vigilanty version`

## Run modes

| Mode | Command | Input |
| --- | --- | --- |
| Local staged review | `vigilanty run` | staged diff |
| PR review | `vigilanty run --pr-mode --base main` | diff vs base branch |
| CI review | `vigilanty run --ci` | last commit diff |
| Machine-readable CI | `vigilanty run --json` | staged diff as JSON |

JSON output details live in [`docs/run-json.md`](docs/run-json.md).

## Providers

Vigilanty currently supports these AI providers:

| Provider | Model behavior | Notes |
| --- | --- | --- |
| `claude` | Usually no explicit model needed | Anthropic Claude CLI |
| `gemini` | Optional in config, provider CLI decides actual model | Gemini CLI |
| `codex` | Usually no explicit model needed | OpenAI Codex CLI |
| `opencode` | Optional, passed with `--model` when set | SST OpenCode CLI |
| `ollama` | **Required** | Local Ollama models |
| `lmstudio` | Optional in config | LM Studio local runtime |
| `github` | Optional in config | GitHub CLI / Copilot-backed workflows |

Example Ollama step:

```yaml
- name: ai-review
  checker: ai-review
  provider: ollama
  model: llama3.1
  prompt: "Review this diff for bugs and maintainability issues."
```

Provider-specific setup guidance lives in [`docs/providers.md`](docs/providers.md).

## Presets

`vigilanty init --preset <name>` scaffolds starter pipelines for **11 presets**:

| Preset | Typical steps |
| --- | --- |
| `go` | `golangci-lint`, `go build`, `go test`, AI review |
| `node` | `npm run lint`, `npm run typecheck`, `npm test`, AI review |
| `typescript` | `npx tsc --noEmit`, `npx eslint .`, `npm test`, AI review |
| `python` | `ruff`, `pytest`, AI review |
| `rust` | `cargo clippy`, `cargo build`, `cargo test`, AI review |
| `java` | `mvn compile`/`gradle build`, `mvn test`/`gradle test`, AI review |
| `dotnet` | `dotnet build`, `dotnet test`, AI review |
| `ruby` | `rubocop`, `rspec`, AI review |
| `swift` | `swift build`, `swift test`, AI review |
| `php` | `phpstan`, `phpunit`, AI review |
| `generic` | AI review starter with commented shell examples |

## CI integration

### GitHub Actions

```yaml
name: vigilanty

on:
  pull_request:
  push:
    branches: [main]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - run: go install github.com/jelsin29/vigilanty@v0.2.0
      - run: vigilanty init --preset go --no-interactive --force
      - run: vigilanty run --ci --json > vigilanty-report.json
```

### GitLab CI

```yaml
stages:
  - verify

vigilanty:
  stage: verify
  image: golang:1.22
  script:
    - go install github.com/jelsin29/vigilanty@v0.2.0
    - vigilanty init --preset go --no-interactive --force
    - vigilanty run --ci --json > vigilanty-report.json
  artifacts:
    when: always
    paths:
      - vigilanty-report.json
```

More integration patterns are documented in [`docs/integrations.md`](docs/integrations.md).

## Example config

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

  - name: test
    checker: shell
    command: "go test ./..."
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    rules_file: "AGENTS.md"
    prompt: "Review this diff for bugs, maintainability issues, and security risks."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
```

Full config field guidance lives in [`docs/configuration.md`](docs/configuration.md).

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Pipeline passed |
| `1` | Pipeline failed |
| `2` | Configuration error |
| `3` | Git error |
| `4` | Internal error |

## Documentation

| Guide | What it covers |
| --- | --- |
| [`docs/cli.md`](docs/cli.md) | Canonical command reference and flags |
| [`docs/configuration.md`](docs/configuration.md) | `.vigilanty.yml` structure, field meanings, and patterns |
| [`docs/providers.md`](docs/providers.md) | Provider setup, model behavior, and interactive selection flow |
| [`docs/integrations.md`](docs/integrations.md) | Git hooks, CI pipelines, and automation entry points |
| [`docs/examples.md`](docs/examples.md) | Successful runs, failure runs, and sample configurations |
| [`docs/troubleshooting.md`](docs/troubleshooting.md) | Common setup, provider, git, and config failure modes |
| [`docs/rules-file.md`](docs/rules-file.md) | How `AGENTS.md` and `rules_file` shape AI review behavior |
| [`docs/architecture.md`](docs/architecture.md) | Internal system design for contributors and advanced users |
| [`docs/run-json.md`](docs/run-json.md) | `--json` schema and compatibility rules |

## Contributing

PRs are welcome. Open an issue for large changes first.

## License

MIT — see [LICENSE](./LICENSE).
