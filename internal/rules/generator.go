// Package rules handles AGENTS.md generation and discovery.
// It generates language-specific coding standards and discovers
// existing rules files for the AI review checker.
package rules

// GenerateOptions configures AGENTS.md generation.
type GenerateOptions struct {
	ProjectType     string   // detected preset (go, typescript, python, etc.)
	FilePatterns    []string // included file patterns
	ExcludePatterns []string // excluded file patterns
	DetectedTools   []string // detected linter/formatter configs
}

// Generate creates an AGENTS.md content string based on the project type.
// Stub — implementation in sub-feature branch.
func Generate(opts GenerateOptions) string {
	return ""
}

// Discover finds the rules file for the current project.
// Checks: AGENTS.md in project root, then rules_file from config.
// Stub — implementation in sub-feature branch.
func Discover(root string, configRulesFile string) (string, bool) {
	return "", false
}

// DetectTools scans the project root for known linter/formatter configs.
// Returns a list of detected tool names (e.g. ["eslint", "prettier", "ruff"]).
// Stub — implementation in sub-feature branch.
func DetectTools(root string) []string {
	return nil
}
