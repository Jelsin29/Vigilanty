# Examples

This page shows what a real Vigilanty workflow looks like once the product is installed, configured, and running against actual repo changes.

## See also

- [README](../README.md)
- [Configuration](./configuration.md)
- [Providers](./providers.md)
- [Troubleshooting](./troubleshooting.md)
- [`run --json`](./run-json.md)

## Example: successful run output

This is the kind of product experience Vigilanty should feel like locally: direct, readable, and pipeline-oriented.

<p align="center">
  <img width="404" height="1016" alt="Vigilanty successful run output" src="https://github.com/user-attachments/assets/fea85988-e32b-4aef-b133-bd7178f84249" />
</p>

### What it demonstrates

- a human-first pipeline summary
- clear step ordering
- pass/fail visibility without parsing JSON
- readiness for local pre-commit use

## Example: failure output

Failure output matters just as much as success output, because this is where teams decide whether a tool is useful or annoying.

<p align="center">
  <img width="962" height="563" alt="Vigilanty failed run output" src="https://github.com/user-attachments/assets/92b6aa0e-c486-407c-aef1-95bd62e55858" />
</p>

### What strong failure output should communicate

- which step failed
- whether it failed as a checker failure or internal error
- enough provider/tool context to fix the problem quickly
- confidence that the pipeline stopped for a real reason

## Example: repo-local config for a Go project

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
  - name: golangci-lint
    checker: shell
    command: "golangci-lint run ./..."

  - name: go-build
    checker: shell
    command: "go build ./..."

  - name: go-test
    checker: shell
    command: "go test ./..."

  - name: ai-review
    checker: ai-review
    provider: claude
    rules_file: "AGENTS.md"
    prompt: "Review this Go diff for bugs, security issues, performance regressions, and maintainability concerns."
    max_diff_lines: 500
    skip_on_empty_diff: true
```

## Example: generic project starter

```yaml
version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m

pipeline:
  - name: ai-review
    checker: ai-review
    provider: opencode
    model: openai/gpt-5.4
    rules_file: "AGENTS.md"
    prompt: "Review this diff for correctness, maintainability, and security regressions."
    max_diff_lines: 500
    skip_on_empty_diff: true
```

This is the fastest path when you want to start with AI review first and layer shell checks later.

## Example: PR mode

```bash
vigilanty run --pr-mode --base main
```

Use PR mode when your review target is the branch diff, not just staged files.

## Example: CI JSON output

```bash
vigilanty run --ci --json > vigilanty-report.json
```

For the schema and compatibility guarantees, read [`run-json.md`](./run-json.md).

## Interpreting the results

### Passed

- exit code `0`
- every step either passed or was intentionally skipped
- safe to continue to commit or CI promotion

### Failed

- exit code `1`
- at least one verification step failed or errored
- inspect the failing step first, not the whole config at once

### Config/Git/Internal errors

- exit code `2`: configuration problem
- exit code `3`: git context problem
- exit code `4`: internal runtime problem

When those happen, jump to [Troubleshooting](./troubleshooting.md).
