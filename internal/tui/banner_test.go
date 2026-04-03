package tui

import (
	"strings"
	"testing"
)

func TestBannerWideReturnsBrailleArt(t *testing.T) {
	banner := Banner(80)

	if banner == "[ VIGILANTY ]" {
		t.Fatal("Banner(80) returned fallback text")
	}
	if banner == "" {
		t.Fatal("Banner(80) returned empty string")
	}
	// block art uses these distinctive characters
	if !strings.Contains(banner, "█") {
		t.Fatal("Banner(80) missing block art content")
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

func TestCenterBrailleLinesCentersBasedOnContentWidth(t *testing.T) {
	lines := []string{
		brailleSpace + "⣿⣿⣿" + brailleSpace + brailleSpace, // content=3, padded both sides
		"⣿",                                // content=1
		brailleSpace + brailleSpace + "⣿⣿", // content=2, leading spaces
	}

	got := centerBrailleLines(lines)

	// max content width = 3 (from line 0: ⣿⣿⣿)
	// line 0: (3-3)/2=0 padding → "⣿⣿⣿"
	// line 1: (3-1)/2=1 padding → "⠀⣿"
	// line 2: (3-2)/2=0 padding → "⣿⣿" (integer division truncates)
	want := []string{
		"⣿⣿⣿",
		brailleSpace + "⣿",
		"⣿⣿",
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("centerBrailleLines()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRenderVersionContainsVersionText(t *testing.T) {
	rendered := RenderVersion("v1.2.3")

	if !strings.Contains(stripANSI(rendered), "Vigilanty v1.2.3") {
		t.Fatalf("RenderVersion() missing visible version text: %q", rendered)
	}
}
