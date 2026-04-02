package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBannerWideReturnsBrailleArt(t *testing.T) {
	banner := Banner(80)

	if banner == "[ VIGILANTY ]" {
		t.Fatal("Banner(80) returned fallback text")
	}
	if banner == "" {
		t.Fatal("Banner(80) returned empty string")
	}
	// braille eye has these distinctive characters
	if !strings.Contains(banner, "⣿") {
		t.Fatal("Banner(80) missing braille art content")
	}
}

func TestBannerNarrowReturnsFallback(t *testing.T) {
	if got := Banner(30); got != "[ VIGILANTY ]" {
		t.Fatalf("Banner(30) = %q, want %q", got, "[ VIGILANTY ]")
	}
}

func TestBannerZeroWidthDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Banner(0) panicked: %v", r)
		}
	}()

	_ = Banner(0)
}

func TestCenterBrailleLinesCentersBasedOnTrimmedWidth(t *testing.T) {
	lines := []string{
		"⣿⣿⣿" + brailleSpace + brailleSpace,
		"⣿",
		brailleSpace + "⣿⣿",
	}

	got := centerBrailleLines(lines)
	want := []string{
		"⣿⣿⣿",
		brailleSpace + "⣿",
		brailleSpace + "⣿⣿",
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("centerBrailleLines()[%d] = %q, want %q", i, got[i], want[i])
		}
		if strings.HasSuffix(got[i], brailleSpace) {
			t.Fatalf("centerBrailleLines()[%d] still has trailing braille spaces: %q", i, got[i])
		}
	}
}

func TestRenderVersionUsesBackToTheFutureGradient(t *testing.T) {
	rendered := RenderVersion("v1.2.3")

	if !strings.Contains(stripANSI(rendered), "Vigilanty v1.2.3 — Pre-commit verification pipeline") {
		t.Fatalf("RenderVersion() missing visible version text: %q", rendered)
	}

	runes := []rune("Vigilanty v1.2.3 — Pre-commit verification pipeline")
	checks := []struct {
		index int
		want  lipgloss.Color
	}{
		{index: 0, want: ColorBTTFOrange},
		{index: len(runes) / 2, want: ColorBTTFGold},
		{index: len(runes) - 1, want: ColorBTTFRed},
	}

	for _, check := range checks {
		if got := interpolateGradient(versionGradientStops, check.index, len(runes)); got != check.want {
			t.Fatalf("interpolateGradient(..., %d, %d) = %s, want %s", check.index, len(runes), got, check.want)
		}
	}
}
