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
	case "typescript":
		return typescriptPresetConfigYAML(), nil
	case "python":
		return pythonPresetConfigYAML(), nil
	case "rust":
		return rustPresetConfigYAML(), nil
	case "java":
		return javaPresetConfigYAML(), nil
	case "dotnet":
		return dotnetPresetConfigYAML(), nil
	case "ruby":
		return rubyPresetConfigYAML(), nil
	case "swift":
		return swiftPresetConfigYAML(), nil
	case "php":
		return phpPresetConfigYAML(), nil
	case "generic":
		return genericPresetConfigYAML(), nil
	default:
		return "", fmt.Errorf("unsupported preset %q (expected one of: go, node, typescript, python, rust, java, dotnet, ruby, swift, php, generic)", preset)
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

func rustPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: cargo-fmt
#     checker: shell
#     command: "cargo fmt --check"
#     timeout: "60s"

pipeline:
  - name: cargo-clippy
    checker: shell
    command: "cargo clippy -- -D warnings"
    timeout: "120s"

  - name: cargo-build
    checker: shell
    command: "cargo build"
    timeout: "120s"

  - name: cargo-test
    checker: shell
    command: "cargo test"
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Rust diff for bugs, unsafe code, memory safety, and maintainability. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func typescriptPresetConfigYAML() string {
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
  - name: tsc
    checker: shell
    command: "npx tsc --noEmit"
    timeout: "120s"

  - name: eslint
    checker: shell
    command: "npx eslint ."
    timeout: "120s"

  - name: test
    checker: shell
    command: "npm test"
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this TypeScript diff for type safety, bugs, security issues, and maintainability. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func javaPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: checkstyle
#     checker: shell
#     command: "mvn checkstyle:check || gradle checkstyleMain"
#     timeout: "120s"

pipeline:
  - name: build
    checker: shell
    command: "mvn compile -q || gradle build"
    timeout: "180s"

  - name: test
    checker: shell
    command: "mvn test -q || gradle test"
    timeout: "180s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Java diff for bugs, security vulnerabilities, design issues, and maintainability. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func dotnetPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: dotnet-format
#     checker: shell
#     command: "dotnet format --verify-no-changes"
#     timeout: "120s"

pipeline:
  - name: dotnet-build
    checker: shell
    command: "dotnet build --no-restore"
    timeout: "180s"

  - name: dotnet-test
    checker: shell
    command: "dotnet test --no-build"
    timeout: "180s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this .NET/C# diff for bugs, security issues, SOLID violations, and maintainability. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func rubyPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: brakeman
#     checker: shell
#     command: "bundle exec brakeman"
#     timeout: "120s"

pipeline:
  - name: rubocop
    checker: shell
    command: "bundle exec rubocop"
    timeout: "120s"

  - name: rspec
    checker: shell
    command: "bundle exec rspec"
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Ruby diff for bugs, security issues, and maintainability. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func swiftPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: swift-format
#     checker: shell
#     command: "swift format lint --recursive ."
#     timeout: "120s"

pipeline:
  - name: swift-build
    checker: shell
    command: "swift build"
    timeout: "180s"

  - name: swift-test
    checker: shell
    command: "swift test"
    timeout: "180s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this Swift diff for bugs, memory management issues, and maintainability. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}

func phpPresetConfigYAML() string {
	return `version: "1"

global:
  fail_fast: true
  diff_max_bytes: 262144
  timeout: 2m
  verbose: false

# Example alternative shell step:
#   - name: php-cs-fixer
#     checker: shell
#     command: "vendor/bin/php-cs-fixer fix --dry-run"
#     timeout: "120s"

pipeline:
  - name: phpstan
    checker: shell
    command: "vendor/bin/phpstan analyse"
    timeout: "120s"

  - name: phpunit
    checker: shell
    command: "vendor/bin/phpunit"
    timeout: "120s"

  - name: ai-review
    checker: ai-review
    provider: claude
    prompt: "Review this PHP diff for bugs, security issues, and maintainability. Reply with PASS if the changes look good, or FAIL followed by your findings."
    timeout: "120s"
    max_diff_lines: 500
    skip_on_empty_diff: true
`
}
