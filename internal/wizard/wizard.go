// Package wizard provides the interactive init setup for Vigilanty.
// It handles terminal prompts for file patterns, exclude patterns,
// AI provider selection, and AGENTS.md generation.
package wizard

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
// Stub — implementation in sub-feature branch.
func Run(detectedPreset string) (*InitResult, error) {
	return nil, nil
}
