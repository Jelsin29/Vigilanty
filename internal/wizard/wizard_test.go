package wizard

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"*.go, *.mod", 2},
		{"", 0},
		{"*.py", 1},
		{" , , ", 0}, // all whitespace/empty
		{"*.ts, *.tsx, *.js", 3},
	}

	for _, tt := range tests {
		got := parsePatterns(tt.input)
		if len(got) != tt.want {
			t.Errorf("parsePatterns(%q) = %d items, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestParsePatternsTrimsWhitespace(t *testing.T) {
	got := parsePatterns("  *.go  ,  *.mod  ")
	if len(got) != 2 {
		t.Fatalf("parsePatterns() = %v, want 2 items", got)
	}
	if got[0] != "*.go" || got[1] != "*.mod" {
		t.Fatalf("parsePatterns() = %v, want [*.go, *.mod]", got)
	}
}

func TestMergePatterns(t *testing.T) {
	base := []string{"*.go", "*.mod"}
	extra := []string{"*.go", "*.sum"} // *.go is a dupe

	got := mergePatterns(base, extra)
	if len(got) != 3 {
		t.Fatalf("mergePatterns() = %v, want 3 unique items", got)
	}

	// check order is preserved (base first, then new from extra)
	expected := []string{"*.go", "*.mod", "*.sum"}
	for i, want := range expected {
		if got[i] != want {
			t.Errorf("mergePatterns()[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestMergePatternsEmptyInputs(t *testing.T) {
	got := mergePatterns(nil, nil)
	if len(got) != 0 {
		t.Fatalf("mergePatterns(nil, nil) = %v, want empty", got)
	}

	got = mergePatterns([]string{"*.go"}, nil)
	if len(got) != 1 || got[0] != "*.go" {
		t.Fatalf("mergePatterns([*.go], nil) = %v, want [*.go]", got)
	}
}

func TestDefaultResultSetsProjectType(t *testing.T) {
	result := defaultResult("  Go  ")
	if result.ProjectType != "go" {
		t.Fatalf("defaultResult('Go').ProjectType = %q, want 'go'", result.ProjectType)
	}
	if result.Provider != "claude" {
		t.Fatalf("defaultResult().Provider = %q, want 'claude'", result.Provider)
	}
}

func TestDefaultResultEmptyPreset(t *testing.T) {
	result := defaultResult("")
	if result.ProjectType != "generic" {
		t.Fatalf("defaultResult('').ProjectType = %q, want 'generic'", result.ProjectType)
	}
}

func TestDefaultPatternsForPreset(t *testing.T) {
	tests := []struct {
		preset      string
		wantInclude string
		wantExclude string
	}{
		{"go", "*.go", "*_test.go"},
		{"python", "*.py", "*_test.py"},
		{"typescript", "*.ts", "*.test.ts"},
		{"rust", "*.rs", "target/"},
		{"nope", "*", ""},
	}

	for _, tt := range tests {
		inc, exc := defaultPatternsForPreset(tt.preset)
		if len(inc) == 0 {
			t.Errorf("defaultPatternsForPreset(%q) include is empty", tt.preset)
			continue
		}
		if inc[0] != tt.wantInclude {
			t.Errorf("defaultPatternsForPreset(%q) include[0] = %q, want %q", tt.preset, inc[0], tt.wantInclude)
		}
		if tt.wantExclude != "" && (len(exc) == 0 || exc[0] != tt.wantExclude) {
			t.Errorf("defaultPatternsForPreset(%q) exclude = %v, want first = %q", tt.preset, exc, tt.wantExclude)
		}
	}
}

func TestIndexForChoiceValid(t *testing.T) {
	for i := 1; i <= 7; i++ {
		choice := string(rune('0' + i))
		got := indexForChoice(choice)
		if got != i-1 {
			t.Errorf("indexForChoice(%q) = %d, want %d", choice, got, i-1)
		}
	}
}

func TestIndexForChoiceInvalid(t *testing.T) {
	if got := indexForChoice("0"); got != -1 {
		t.Errorf("indexForChoice('0') = %d, want -1", got)
	}
	if got := indexForChoice("-1"); got != -1 {
		t.Errorf("indexForChoice('-1') = %d, want -1", got)
	}
	if got := indexForChoice("abc"); got != -1 {
		t.Errorf("indexForChoice('abc') = %d, want -1", got)
	}
	// "99" returns 98 — valid int, bounds checked by caller
	if got := indexForChoice("99"); got != 98 {
		t.Errorf("indexForChoice('99') = %d, want 98", got)
	}
}

func TestDetectedPresetLabel(t *testing.T) {
	label := detectedPresetLabel("go")
	if label != "Go project (found go.mod)" {
		t.Fatalf("detectedPresetLabel('go') = %q", label)
	}

	label = detectedPresetLabel("something-weird")
	if label != "Generic project" {
		t.Fatalf("detectedPresetLabel('something-weird') = %q, want 'Generic project'", label)
	}
}

func TestRunReturnsDefaultsForNonTTYInput(t *testing.T) {
	originalStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	defer func() {
		os.Stdin = originalStdin
		r.Close()
		w.Close()
	}()
	os.Stdin = r

	result, err := Run("go")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ProjectType != "go" {
		t.Fatalf("Run().ProjectType = %q, want %q", result.ProjectType, "go")
	}
	if result.Provider != "claude" {
		t.Fatalf("Run().Provider = %q, want %q", result.Provider, "claude")
	}
	if !result.GenerateRules {
		t.Fatal("Run().GenerateRules = false, want true")
	}
}

func TestErrCancelledIsSentinel(t *testing.T) {
	wrapped := errors.Join(ErrCancelled, errors.New("wrapped"))
	if !errors.Is(wrapped, ErrCancelled) {
		t.Fatal("errors.Is should match ErrCancelled")
	}
}

func TestDefaultResultIsAccessible(t *testing.T) {
	result := DefaultResult("go")
	if result == nil {
		t.Fatal("DefaultResult() returned nil")
	}
	if result.ProjectType != "go" {
		t.Fatalf("DefaultResult().ProjectType = %q, want %q", result.ProjectType, "go")
	}
}

func TestPromptProviderKeepsSelectionWhenModelIsEmpty(t *testing.T) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	defer func() {
		os.Stdout = originalStdout
		r.Close()
		w.Close()
	}()
	os.Stdout = w

	reader := bufio.NewReader(strings.NewReader("2\n\ngemini-2.5-pro\n"))
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()

	provider, err := promptProvider(reader)
	_ = w.Close()
	output := <-done
	if err != nil {
		t.Fatalf("promptProvider() error = %v", err)
	}
	if provider != "gemini:gemini-2.5-pro" {
		t.Fatalf("promptProvider() = %q, want %q", provider, "gemini:gemini-2.5-pro")
	}
	if strings.Count(output, "? Select your AI provider:") != 1 {
		t.Fatalf("provider menu rendered %d times, want 1", strings.Count(output, "? Select your AI provider:"))
	}
}
