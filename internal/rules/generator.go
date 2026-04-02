// Package rules handles AGENTS.md generation and discovery.
// It generates language-specific coding standards and discovers
// existing rules files for the AI review checker.
package rules

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenerateOptions configures AGENTS.md generation.
type GenerateOptions struct {
	ProjectType     string   // detected preset (go, typescript, python, etc.)
	FilePatterns    []string // included file patterns
	ExcludePatterns []string // excluded file patterns
	DetectedTools   []string // detected linter/formatter configs
}

// Generate creates an AGENTS.md content string based on the project type.
func Generate(opts GenerateOptions) string {
	include := joinPatterns(opts.FilePatterns)
	exclude := joinPatterns(opts.ExcludePatterns)
	rules := presetRules(opts.ProjectType)

	var builder strings.Builder
	builder.WriteString("# Code Review Rules\n\n")
	builder.WriteString("## File Scope\n\n")
	builder.WriteString("**Include**: ")
	builder.WriteString(include)
	builder.WriteString("\n")
	builder.WriteString("**Exclude**: ")
	builder.WriteString(exclude)
	builder.WriteString("\n\n---\n\n")
	builder.WriteString("## ")
	builder.WriteString(rules.Title)
	builder.WriteString(" Rules\n\n")
	builder.WriteString("REJECT if:\n")
	for _, rule := range rules.Reject {
		builder.WriteString("- ")
		builder.WriteString(rule)
		builder.WriteString("\n")
	}
	builder.WriteString("\nREQUIRE:\n")
	for _, rule := range rules.Require {
		builder.WriteString("- ")
		builder.WriteString(rule)
		builder.WriteString("\n")
	}
	builder.WriteString("\nPREFER:\n")
	for _, rule := range rules.Prefer {
		builder.WriteString("- ")
		builder.WriteString(rule)
		builder.WriteString("\n")
	}

	if len(opts.DetectedTools) > 0 {
		builder.WriteString("\n---\n\n## Detected Tools\n\n")
		builder.WriteString("The following tools are configured in this project: ")
		builder.WriteString(strings.Join(sortedUnique(opts.DetectedTools), ", "))
		builder.WriteString("\nRespect their configurations during review.\n")
	}

	builder.WriteString("\n---\n\n## Response Format\n\n")
	builder.WriteString("FIRST LINE must be exactly:\n")
	builder.WriteString("STATUS: PASSED\n")
	builder.WriteString("or\n")
	builder.WriteString("STATUS: FAILED\n\n")
	builder.WriteString("If FAILED, list each issue as:\n")
	builder.WriteString("- file:line - rule violated - description\n")

	return builder.String()
}

// Discover finds the rules file for the current project.
// Checks: AGENTS.md in project root, then rules_file from config.
// Stub — implementation in sub-feature branch.
func Discover(root string, configRulesFile string) (string, bool) {
	return "", false
}

// DetectTools scans the project root for known linter/formatter configs.
// Returns a list of detected tool names (e.g. ["eslint", "prettier", "ruff"]).
func DetectTools(root string) []string {
	tools := make(map[string]struct{})

	addIfMatched(root, tools, "eslint", []string{".eslintrc*", "eslint.config.*"})
	addIfMatched(root, tools, "prettier", []string{".prettierrc*", "prettier.config.*"})
	addIfMatched(root, tools, "golangci-lint", []string{".golangci.yml", ".golangci.yaml"})
	addIfMatched(root, tools, "rustfmt", []string{"rustfmt.toml", ".rustfmt.toml"})
	addIfMatched(root, tools, "clippy", []string{"clippy.toml", ".clippy.toml"})
	addIfMatched(root, tools, "rubocop", []string{".rubocop.yml"})
	addIfMatched(root, tools, "phpstan", []string{"phpstan.neon*"})
	addIfMatched(root, tools, "swiftlint", []string{".swiftlint.yml"})
	addIfMatched(root, tools, "checkstyle", []string{"checkstyle.xml"})
	addIfMatched(root, tools, "ruff", []string{"ruff.toml", ".ruff.toml"})

	pyprojectPath := filepath.Join(root, "pyproject.toml")
	if info, err := os.Stat(pyprojectPath); err == nil && !info.IsDir() {
		content, readErr := os.ReadFile(pyprojectPath)
		if readErr == nil && strings.Contains(string(content), "[tool.ruff]") {
			tools["ruff"] = struct{}{}
		}
	}

	return sortedKeys(tools)
}

type languageRules struct {
	Title   string
	Reject  []string
	Require []string
	Prefer  []string
}

func joinPatterns(patterns []string) string {
	clean := sortedUnique(patterns)
	if len(clean) == 0 {
		return "(none)"
	}
	return strings.Join(clean, ", ")
}

func presetRules(projectType string) languageRules {
	switch strings.ToLower(strings.TrimSpace(projectType)) {
	case "", "default", "go":
		return languageRules{
			Title:   "Go",
			Reject:  []string{"Exported functions without doc comments", "Naked returns in functions longer than 10 lines", "panic() in library code", "interface{} instead of any"},
			Require: []string{"Wrap errors with fmt.Errorf(\"...: %w\", err)", "Use context.Context as the first parameter when context is required"},
			Prefer:  []string{"Table-driven tests", "Early returns over deep nesting"},
		}
	case "typescript":
		return languageRules{
			Title:   "TypeScript",
			Reject:  []string{"any type without justification", "console.log in production code", "var declarations", "Type assertions without a comment"},
			Require: []string{"Explicit return types on exported functions", "Error handling in async functions"},
			Prefer:  []string{"const over let", "Discriminated unions over type guards", "Named exports"},
		}
	case "python":
		return languageRules{
			Title:   "Python",
			Reject:  []string{"Bare except:", "print() instead of logging", "Mutable default arguments", "import *"},
			Require: []string{"Type hints on public functions", "Docstrings on public classes and methods"},
			Prefer:  []string{"f-strings over .format()", "pathlib over os.path", "dataclasses for data containers"},
		}
	case "node", "javascript", "js":
		return languageRules{
			Title:   "Node/JavaScript",
			Reject:  []string{"var declarations", "== instead of ===", "console.log in production code", "Callback hell deeper than 3 levels"},
			Require: []string{"Error handling in promises and async functions", "Input validation on API endpoints"},
			Prefer:  []string{"const over let", "Destructuring", "Template literals"},
		}
	case "rust":
		return languageRules{
			Title:   "Rust",
			Reject:  []string{"unwrap() in library code", "unsafe blocks without a safety comment", "clone() without justification"},
			Require: []string{"Error types implementing std::error::Error", "Doc comments on public items"},
			Prefer:  []string{"? operator over match for error propagation", "Iterators over manual loops"},
		}
	case "java":
		return languageRules{
			Title:   "Java",
			Reject:  []string{"Empty catch blocks", "System.out.println in production", "Raw types", "null returns without documentation"},
			Require: []string{"@Override annotations where applicable", "Resource management with try-with-resources"},
			Prefer:  []string{"Streams over manual loops", "Immutable collections", "Builder pattern for constructors with more than 3 parameters"},
		}
	case "dotnet", "c#", "csharp":
		return languageRules{
			Title:   "C#/.NET",
			Reject:  []string{"Empty catch blocks", "Console.WriteLine in production", "public fields instead of properties"},
			Require: []string{"async/await for I/O operations", "Nullable reference type annotations"},
			Prefer:  []string{"Pattern matching", "LINQ over manual loops", "Records for DTOs"},
		}
	case "ruby":
		return languageRules{
			Title:   "Ruby",
			Reject:  []string{"eval", "send with user input", "rescue without a specific exception"},
			Require: []string{"frozen_string_literal comment", "Keyword arguments for functions with more than 2 parameters"},
			Prefer:  []string{"Guard clauses over nested conditionals", "Symbols over strings for keys"},
		}
	case "swift":
		return languageRules{
			Title:   "Swift",
			Reject:  []string{"Force unwrap ! without justification", "Any type", "Implicitly unwrapped optionals"},
			Require: []string{"guard for early exits", "Access control on all declarations"},
			Prefer:  []string{"let over var", "Protocol-oriented design", "Value types over reference types"},
		}
	case "php":
		return languageRules{
			Title:   "PHP",
			Reject:  []string{"eval()", "$$ variable variables", "@ error suppression", "SQL without prepared statements"},
			Require: []string{"Type declarations on functions", "Namespace usage"},
			Prefer:  []string{"Early returns", "Readonly properties (PHP 8.1+)", "match expressions"},
		}
	default:
		return languageRules{
			Title:   "Generic",
			Reject:  []string{"Hardcoded secrets or credentials", "Empty error handling", "Code duplication"},
			Require: []string{"Descriptive naming", "Error messages that help debugging"},
			Prefer:  []string{"Small functions (under 30 lines)", "Single responsibility"},
		}
	}
}

func addIfMatched(root string, tools map[string]struct{}, name string, patterns []string) {
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			if info, statErr := os.Stat(match); statErr == nil && !info.IsDir() {
				tools[name] = struct{}{}
				return
			}
		}
	}
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}

func sortedUnique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}
	sort.Strings(unique)
	return unique
}
