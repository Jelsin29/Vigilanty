# Integrations

Vigilanty is most useful when it becomes part of your normal delivery path, not a one-off local command.

## See also

- [README](../README.md)
- [CLI reference](./cli.md)
- [Configuration](./configuration.md)
- [`run --json`](./run-json.md)
- [Examples](./examples.md)

## Git hook integration

Install the pre-commit hook with:

```bash
vigilanty install
```

Remove it with:

```bash
vigilanty uninstall
```

### What this gives you

- local verification before commits leave your machine
- one consistent entry point instead of ad-hoc shell glue
- a clean path from local workflow to CI workflow

### Tradeoff

Pre-commit hooks improve discipline, but they are not a replacement for CI. The serious setup is **both**: hook locally, CI remotely.

## GitHub Actions

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

### Why `--ci`

`--ci` evaluates the last commit diff, which matches common CI verification behavior without relying on staged local changes.

### Why `--json`

Because CI should consume a stable machine-readable contract. Human output is great locally; JSON is what automation wants.

## GitLab CI

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

## PR review workflows

For branch comparison instead of staged or CI mode:

```bash
vigilanty run --pr-mode --base main
```

Use this when:

- you want the full PR diff reviewed locally before opening a PR
- your workflow needs a branch comparison against a base branch
- you are validating integration changes rather than staged edits

## Automation and scripts

### Example: shell script gate

```bash
#!/usr/bin/env bash
set -euo pipefail

report_file="vigilanty-report.json"
vigilanty run --ci --json > "$report_file"
```

### Example: custom config path

```bash
vigilanty --config .ci/vigilanty.yml run --ci --json
```

## Config inspection in pipelines

When debugging CI behavior, this is the fastest visibility command:

```bash
vigilanty config
```

It tells you:

- which config file was actually loaded
- the global settings in effect
- the current pipeline order
- provider/model metadata for AI steps

## Cache management

Useful operational commands:

```bash
vigilanty cache status
vigilanty cache clear
```

Use cache inspection when you need to explain repeated results or verify that your pipeline is not silently reusing prior step outputs.

## Recommended rollout path

1. Start with `vigilanty init`
2. Review and adjust `.vigilanty.yml`
3. Install the pre-commit hook locally
4. Add `vigilanty run --ci --json` to CI
5. Tune prompts and `AGENTS.md` once the baseline is stable

That sequence keeps the product onboarding smooth while preserving a path to team-wide rigor.
