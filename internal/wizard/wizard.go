// Package wizard provides the interactive init setup for Vigilanty.
// It handles terminal prompts for file patterns, exclude patterns,
// AI provider selection, and AGENTS.md generation.
package wizard

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// InitResult holds the user's choices from the interactive wizard.
type InitResult struct {
	ProjectType     string   // detected or selected preset
	FilePatterns    []string // e.g. ["*.py", "*.pyi"]
	ExcludePatterns []string // e.g. ["*_test.py", "*.spec.ts"]
	Provider        string   // e.g. "claude", "gemini", "ollama:llama3"
	RulesFile       string   // path to rules file, default "AGENTS.md"
	GenerateRules   bool     // whether to generate AGENTS.md
}

// Run executes the interactive wizard and returns the user's choices.
func Run(detectedPreset string) (*InitResult, error) {
	defaults := defaultResult(detectedPreset)

	isTTY, err := stdinIsTTY()
	if err != nil {
		return nil, err
	}
	if !isTTY {
		return defaults, nil
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Fprintln(os.Stdout, "🔧 Vigilanty Setup")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintf(os.Stdout, "Detected: %s\n\n", detectedPresetLabel(defaults.ProjectType))

	includePatterns, err := promptPatternList(reader, "include", defaults.FilePatterns)
	if err != nil {
		return nil, err
	}

	excludePatterns, err := promptPatternList(reader, "exclude", defaults.ExcludePatterns)
	if err != nil {
		return nil, err
	}

	provider, err := promptProvider(reader)
	if err != nil {
		return nil, err
	}

	generateRules, err := promptGenerateRules(reader)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "✅ Configuration complete!")

	return &InitResult{
		ProjectType:     defaults.ProjectType,
		FilePatterns:    includePatterns,
		ExcludePatterns: excludePatterns,
		Provider:        provider,
		RulesFile:       defaults.RulesFile,
		GenerateRules:   generateRules,
	}, nil
}

func defaultResult(detectedPreset string) *InitResult {
	projectType := strings.ToLower(strings.TrimSpace(detectedPreset))
	if projectType == "" {
		projectType = "generic"
	}

	includePatterns, excludePatterns := defaultPatternsForPreset(projectType)

	return &InitResult{
		ProjectType:     projectType,
		FilePatterns:    append([]string(nil), includePatterns...),
		ExcludePatterns: append([]string(nil), excludePatterns...),
		Provider:        "claude",
		RulesFile:       "AGENTS.md",
		GenerateRules:   true,
	}
}

func stdinIsTTY() (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("inspect stdin: %w", err)
	}

	return (info.Mode() & os.ModeCharDevice) != 0, nil
}

func defaultPatternsForPreset(preset string) ([]string, []string) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "go":
		return []string{"*.go"}, []string{"*_test.go", "vendor/"}
	case "typescript":
		return []string{"*.ts", "*.tsx"}, []string{"*.test.ts", "*.spec.ts", "node_modules/"}
	case "node":
		return []string{"*.js", "*.jsx", "*.ts", "*.tsx"}, []string{"*.test.js", "*.spec.js", "node_modules/", "dist/"}
	case "python":
		return []string{"*.py"}, []string{"*_test.py", "test_*", "__pycache__/"}
	case "rust":
		return []string{"*.rs"}, []string{"target/"}
	case "java":
		return []string{"*.java"}, []string{"*Test.java", "build/", "target/"}
	case "dotnet":
		return []string{"*.cs"}, []string{"*Tests.cs", "bin/", "obj/"}
	case "ruby":
		return []string{"*.rb"}, []string{"*_spec.rb", "spec/"}
	case "swift":
		return []string{"*.swift"}, []string{"*Tests.swift", ".build/"}
	case "php":
		return []string{"*.php"}, []string{"*Test.php", "vendor/"}
	default:
		return []string{"*"}, nil
	}
}

func detectedPresetLabel(preset string) string {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "go":
		return "Go project (found go.mod)"
	case "typescript":
		return "TypeScript project (found tsconfig.json)"
	case "node":
		return "Node project (found package.json)"
	case "python":
		return "Python project"
	case "rust":
		return "Rust project (found Cargo.toml)"
	case "java":
		return "Java project"
	case "dotnet":
		return ".NET project"
	case "ruby":
		return "Ruby project (found Gemfile)"
	case "swift":
		return "Swift project (found Package.swift)"
	case "php":
		return "PHP project (found composer.json)"
	default:
		return "Generic project"
	}
}

func promptPatternList(reader *bufio.Reader, mode string, defaults []string) ([]string, error) {
	var title string
	if mode == "include" {
		title = "? Select file patterns to include in review:"
	} else {
		title = "? Select file patterns to exclude from review:"
	}

	defaultLabel := strings.Join(defaults, ", ")
	if defaultLabel == "" {
		defaultLabel = "(none)"
	}

	fmt.Fprintln(os.Stdout, title)
	fmt.Fprintf(os.Stdout, "  Suggested: %s\n", defaultLabel)
	fmt.Fprint(os.Stdout, "  Add more (comma-separated, or press Enter to keep): ")

	input, err := readLine(reader)
	if err != nil {
		return nil, err
	}
	if input == "" {
		return append([]string(nil), defaults...), nil
	}

	return mergePatterns(defaults, parsePatterns(input)), nil
}

func promptProvider(reader *bufio.Reader) (string, error) {
	providers := []string{"claude", "gemini", "ollama", "codex", "opencode", "lmstudio", "github"}

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "? Select your AI provider:")
	for index, provider := range providers {
		fmt.Fprintf(os.Stdout, "  %d) %s\n", index+1, provider)
	}

	for {
		fmt.Fprint(os.Stdout, "  Choice [1]: ")
		input, err := readLine(reader)
		if err != nil {
			return "", err
		}

		choice := strings.TrimSpace(input)
		if choice == "" {
			choice = "1"
		}

		idx := indexForChoice(choice)
		if idx < 0 || idx >= len(providers) {
			fmt.Fprintf(os.Stdout, "  Invalid choice. Please enter a number between 1 and %d.\n", len(providers))
			continue
		}
		provider := providers[idx]

		// claude and codex work without specifying a model
		if provider == "claude" || provider == "codex" {
			return provider, nil
		}

		// everything else needs a model — gemini, ollama, opencode, lmstudio, github
		for {
			fmt.Fprintf(os.Stdout, "  Model for %s: ", provider)
			model, err := readLine(reader)
			if err != nil {
				return "", err
			}
			model = strings.TrimSpace(model)
			if model == "" {
				fmt.Fprintln(os.Stdout, "  Please enter a model name.")
				continue
			}
			return provider + ":" + model, nil
		}
	}
}

func promptGenerateRules(reader *bufio.Reader) (bool, error) {
	fmt.Fprintln(os.Stdout)
	fmt.Fprint(os.Stdout, "? Generate AGENTS.md with coding standards? [Y/n]: ")

	input, err := readLine(reader)
	if err != nil {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(input)) {
	case "", "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return true, nil
	}
}

func indexForChoice(choice string) int {
	n, err := strconv.Atoi(choice)
	if err != nil || n < 1 {
		return -1
	}
	return n - 1
}

func readLine(reader *bufio.Reader) (string, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return strings.TrimSpace(input), nil
		}
		return "", fmt.Errorf("read input: %w", err)
	}

	return strings.TrimSpace(input), nil
}

func parsePatterns(input string) []string {
	parts := strings.Split(input, ",")
	patterns := make([]string, 0, len(parts))
	for _, part := range parts {
		pattern := strings.TrimSpace(part)
		if pattern == "" {
			continue
		}
		patterns = append(patterns, pattern)
	}
	return patterns
}

func mergePatterns(base []string, extra []string) []string {
	merged := make([]string, 0, len(base)+len(extra))
	seen := make(map[string]struct{}, len(base)+len(extra))

	for _, pattern := range append(append([]string(nil), base...), extra...) {
		if _, exists := seen[pattern]; exists {
			continue
		}
		seen[pattern] = struct{}{}
		merged = append(merged, pattern)
	}

	return merged
}
