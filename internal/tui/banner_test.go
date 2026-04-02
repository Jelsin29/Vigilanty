package tui

import (
	"strings"
	"testing"
)

func TestBannerWideReturnsASCIIArt(t *testing.T) {
	banner := Banner(80)

	if banner == "[ VIGILANTY ]" {
		t.Fatal("Banner(80) returned fallback text")
	}
	if banner == "" {
		t.Fatal("Banner(80) returned empty string")
	}
	if !strings.Contains(banner, "%@%") {
		t.Fatalf("Banner(80) = %q, want ASCII art marker", banner)
	}
}

func TestBannerNarrowReturnsFallback(t *testing.T) {
	if got := Banner(50); got != "[ VIGILANTY ]" {
		t.Fatalf("Banner(50) = %q, want %q", got, "[ VIGILANTY ]")
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
