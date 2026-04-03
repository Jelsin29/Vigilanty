package ui

import (
	"strings"
	"testing"
)

func TestRunBannerContainsVersion(t *testing.T) {
	banner := RunBanner("v0.2.0-dev")

	if !strings.Contains(banner, "v0.2.0-dev") {
		t.Fatalf("RunBanner() = %q, want version", banner)
	}
}

func TestRunBannerContainsArt(t *testing.T) {
	banner := RunBanner("dev")

	if !strings.Contains(banner, "▀██▀") {
		t.Fatalf("RunBanner() = %q, want Vigilanty art", banner)
	}
	if !strings.Contains(banner, "▀▀▀▀▀ ▀▀▀▀▀") {
		t.Fatalf("RunBanner() = %q, want border art", banner)
	}
}
