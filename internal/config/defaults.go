package config

import (
	"fmt"
	"strings"
)

func DefaultConfigYAML() string {
	content, err := ConfigYAMLForPreset("")
	if err != nil {
		panic(err)
	}

	return content
}

func ConfigYAMLForPreset(preset string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "", "default", "go":
		return goPresetConfigYAML(), nil
	case "node":
		return nodePresetConfigYAML(), nil
	case "python":
		return pythonPresetConfigYAML(), nil
	case "generic":
		return genericPresetConfigYAML(), nil
	default:
		return "", fmt.Errorf("unsupported preset %q (expected one of: go, node, python, generic)", preset)
	}
}

func genericPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example steps you can adapt:
#   - name: lint
#     checker: shell
#     command: "make lint"
#     timeout: "60s"
#
#   - name: test
#     checker: shell
#     command: "make test"
#     timeout: "120s"

pipeline:
  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this diff for bugs, security issues, and maintainability problems. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func goPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: gofmt
#     checker: shell
#     command: "gofmt -w ."
#     timeout: "60s"

pipeline:
  - name: golangci-lint
    checker: shell
    command: "golangci-lint run ./..."
    timeout: "120s"

  - name: go-build
    checker: shell
    command: "go build ./..."
    timeout: "120s"

  - name: go-test
    checker: shell
    command: "go test ./..."
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Go diff for bugs, security issues, performance regressions, and maintainability concerns. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func nodePresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: prettier
#     checker: shell
#     command: "npx prettier --check ."
#     timeout: "60s"

pipeline:
  - name: eslint
    checker: shell
    command: "npm run lint"
    timeout: "120s"

  - name: tsc
    checker: shell
    command: "npm run typecheck"
    timeout: "120s"

  - name: npm-test
    checker: shell
    command: "npm test"
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Node.js diff for bugs, security issues, and maintainability problems. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func pythonPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: mypy
#     checker: shell
#     command: "mypy ."
#     timeout: "120s"

pipeline:
  - name: ruff
    checker: shell
    command: "ruff check ."
    timeout: "120s"

  - name: pytest
    checker: shell
    command: "pytest"
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Python diff for bugs, security issues, and maintainability problems. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}
