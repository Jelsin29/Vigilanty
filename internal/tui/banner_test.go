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
