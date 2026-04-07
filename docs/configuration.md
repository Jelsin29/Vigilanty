# Configuration Guide

Vigilanty reads configuration from a repo-local `.vigilanty.yml` by default, or from `~/.config/vigilanty/config.yml` when a local file is not present. You can always override the source with `--config <path>`.

## See also

- [README](../README.md)
- [CLI reference](./cli.md)
- [Providers](./providers.md)
- [Rules file](./rules-file.md)
- [Troubleshooting](./troubleshooting.md)

## Configuration resolution order

When you run `vigilanty run` or `vigilanty config`, Vigilanty resolves config in this order:

1. `--config <path>`
2. `.vigilanty.yml` in the current repository
3. `~/.config/vigilanty/config.yml`

If no config file is found, the CLI returns a configuration error and tells you to run `vigilanty init`.

## Top-level structure

```yaml
version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false
  file_patterns:
    - "*.go"
  exclude_patterns:
    - "*_test.go"
    - "vendor/"

pipeline:
  - name: lint
    checker: shell
    command: "golangci-lint run ./..."

  - name: ai-review
    checker: ai-review
    provider: claude
    rules_file: "AGENTS.md"
    prompt: "Review this diff for bugs, security issues, and maintainability problems."
```

## `version`

- Current supported value: `"1"`
- If omitted, Vigilanty defaults it to `"1"`
- Any other value is rejected during config validation

## `global`

Global settings apply to the whole pipeline unless a step overrides them.

| Field | Type | Default | Notes |
| --- | --- | --- | --- |
| `fail_fast` | boolean | `true` in generated presets | Stops after the first failed/error step |
| `diff_max_bytes` | integer | `262144` | Maximum diff payload size before truncation |
| `timeout` | duration string | `2m` | Default timeout inherited by steps |
| `verbose` | boolean | `false` | Enables more detailed checker output |
| `file_patterns` | string[] | preset-specific in wizard | Defines the include scope shown in the banner |
| `exclude_patterns` | string[] | preset-specific in wizard | Defines exclusions for review scope |

### Timeout format

Timeouts use Go duration syntax, for example:

- `30s`
- `90s`
- `2m`
- `5m`

Invalid durations fail config validation before the pipeline starts.

## `pipeline`

`pipeline` must contain at least one named step. Step names must be unique.

Supported checker types:

- `shell`
- `command`
- `ai-review`
- `ai`
- `prompt`

In practice, the product is centered around two workflows:

1. **Shell steps** for lint, tests, formatting, or security
2. **AI review steps** for staged diff analysis against project rules

## Step fields

| Field | Required | Applies to | Notes |
| --- | --- | --- | --- |
| `name` | Yes | all steps | Human-readable label shown in output |
| `checker` | Yes | all steps | Usually `shell` or `ai-review` |
| `type` | Optional if `checker` is set | all steps | Aliased internally with `checker` |
| `command` | Yes for `shell` / `command` | shell steps | Command executed in the repo root |
| `provider` | Yes for AI steps | AI steps | Supported providers documented in [providers.md](./providers.md) |
| `model` | Optional except Ollama | AI steps | Required for `ollama`; optional elsewhere |
| `prompt` | Yes for AI steps | AI steps | Additional review instructions appended after coding standards |
| `rules_file` | Optional | AI steps | Explicit rules file path; falls back to root `AGENTS.md` |
| `timeout` | Optional | all steps | Inherits `global.timeout` when omitted |
| `enabled` | Optional | all steps | Can be used to toggle a step on/off |
| `env` | Optional | all steps | Extra environment variables for a step |
| `skip_on_empty_diff` | Optional | AI steps | Skips review when the evaluated diff is empty |
| `max_diff_lines` | Optional | AI steps | Truncates large diffs before sending to the provider |
| `pass_pattern` | Optional | AI steps | Custom regex to detect pass responses |
| `fail_pattern` | Optional | AI steps | Custom regex to detect fail responses |
| `config` | Optional | all steps | Free-form checker configuration map |

## Shell step example

```yaml
- name: test
  checker: shell
  command: "go test ./..."
  timeout: "120s"
```

Use shell steps for deterministic, tool-driven verification. Put lint/build/test/security here first, then finish with AI review.

## AI review step example

```yaml
- name: ai-review
  checker: ai-review
  provider: opencode
  model: openai/gpt-5.4
  rules_file: "AGENTS.md"
  prompt: "Review this diff for correctness, maintainability, and security regressions."
  timeout: "120s"
  max_diff_lines: 500
  skip_on_empty_diff: true
```

The effective prompt sent to the provider is composed from:

1. A built-in review instruction
2. The contents of `rules_file` or root `AGENTS.md`
3. Your custom `prompt`
4. The current diff under review

## Preset behavior

`vigilanty init` can generate a starter config from presets such as `go`, `node`, `typescript`, `python`, and `generic`.

Preset output includes:

- global defaults
- starter shell steps for the detected ecosystem
- one AI review step
- guidance comments at the top of the generated file

If you run the interactive wizard, the generated config also reflects your selected provider, file patterns, and exclusions.

## Recommended pipeline shape

For most teams, this ordering works best:

1. Fast formatting/linting
2. Build or typecheck
3. Tests or security scanning
4. AI review on the resulting diff

Why this order? Because expensive or subjective review should not run before deterministic failures are already known.

## Local vs global config

Use **repo-local config** when:

- the pipeline belongs to the repository
- you want the same rules in CI and local development
- teammates should share the same verification contract

Use **global config** when:

- you want a personal default for experiments
- you are testing Vigilanty outside a committed project

For serious team usage, repo-local `.vigilanty.yml` is the better default.

## Validating configuration quickly

Use:

```bash
vigilanty config
```

This prints:

- the resolved config source
- global settings
- pipeline step list
- provider/model settings when relevant

## Common mistakes

- Using an unsupported `version`
- Forgetting `prompt` on an AI review step
- Forgetting `provider` on an AI review step
- Forgetting `model` for `ollama`
- Writing invalid durations such as `120` instead of `120s`
- Defining duplicate step names

For fixes, jump to [Troubleshooting](./troubleshooting.md).
