package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jelsin29/vigilanty/internal/config"
)

func TestPrintPresetPreviewIncludesGeneratedSteps(t *testing.T) {
	content, err := config.ConfigYAMLForPreset("go")
	if err != nil {
		t.Fatalf("ConfigYAMLForPreset() error = %v", err)
	}

	var out bytes.Buffer
	if err := printPresetPreview(&out, "go", content); err != nil {
		t.Fatalf("printPresetPreview() error = %v", err)
	}

	printed := out.String()
	if !strings.Contains(printed, "Preset preview (go):") {
		t.Fatalf("preview = %q, want preset header", printed)
	}
	if !strings.Contains(printed, "golangci-lint — golangci-lint run ./...") {
		t.Fatalf("preview = %q, want lint step", printed)
	}
	if !strings.Contains(printed, "ai-review — claude") {
		t.Fatalf("preview = %q, want provider step", printed)
	}
}

func TestPrintInitNextStepsMentionsJSONMode(t *testing.T) {
	var out bytes.Buffer
	printInitNextSteps(&out)

	printed := out.String()
	for _, snippet := range []string{"Next Steps:", "vigilanty install", "vigilanty run", "vigilanty run --json"} {
		if !strings.Contains(printed, snippet) {
			t.Fatalf("next steps = %q, missing %q", printed, snippet)
		}
	}
}
