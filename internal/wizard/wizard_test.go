package wizard

import (
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
	if got := indexForChoice("abc"); got != -1 {
		t.Errorf("indexForChoice('abc') = %d, want -1", got)
	}
	if got := indexForChoice("99"); got != -1 {
		t.Errorf("indexForChoice('99') = %d, want -1", got)
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
