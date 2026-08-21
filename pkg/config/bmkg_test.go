package config

import (
	"testing"
	"time"
)

func TestLoadBMKGConfigUsesDefaults(t *testing.T) {
	t.Setenv("BMKG_BASE_URL", "")
	t.Setenv("BMKG_TIMEOUT", "")

	got := LoadBMKGConfig()

	if got.BaseURL != "https://data.bmkg.go.id" {
		t.Fatalf("expected default BMKG base URL, got %q", got.BaseURL)
	}
	if got.Timeout != 5*time.Second {
		t.Fatalf("expected default timeout 5s, got %v", got.Timeout)
	}
}

func TestLoadBMKGConfigReadsOverrides(t *testing.T) {
	t.Setenv("BMKG_BASE_URL", " https://bmkg.example.test/root ")
	t.Setenv("BMKG_TIMEOUT", "750ms")

	got := LoadBMKGConfig()

	if got.BaseURL != "https://bmkg.example.test/root" {
		t.Fatalf("expected configured BMKG base URL, got %q", got.BaseURL)
	}
	if got.Timeout != 750*time.Millisecond {
		t.Fatalf("expected configured timeout, got %v", got.Timeout)
	}
}

func TestValidateBMKGBaseURLRejectsUnsafeURLs(t *testing.T) {
	for _, baseURL := range []string{
		"ftp://data.bmkg.go.id",
		"https://data.bmkg.go.id?tenant=city",
		"https://data.bmkg.go.id/root#fragment",
	} {
		if err := ValidateBMKGBaseURL(baseURL); err == nil {
			t.Fatalf("expected unsafe BMKG base URL %q to fail validation", baseURL)
		}
	}
}
