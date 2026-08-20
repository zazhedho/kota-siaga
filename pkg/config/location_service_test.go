package config

import (
	"testing"
	"time"
)

func TestLoadLocationServiceConfigUsesDefaults(t *testing.T) {
	t.Setenv("LOCATION_SERVICE_BASE_URL", "")
	t.Setenv("LOCATION_SERVICE_TIMEOUT", "")

	got := LoadLocationServiceConfig()

	if got.BaseURL != "https://indonesia.imyourz.com" {
		t.Fatalf("expected default location service URL, got %q", got.BaseURL)
	}
	if got.Timeout != 5*time.Second {
		t.Fatalf("expected default timeout 5s, got %v", got.Timeout)
	}
}

func TestLoadLocationServiceConfigReadsOverrides(t *testing.T) {
	t.Setenv("LOCATION_SERVICE_BASE_URL", " https://locations.example.test/root ")
	t.Setenv("LOCATION_SERVICE_TIMEOUT", "750ms")

	got := LoadLocationServiceConfig()

	if got.BaseURL != "https://locations.example.test/root" {
		t.Fatalf("expected configured location service URL, got %q", got.BaseURL)
	}
	if got.Timeout != 750*time.Millisecond {
		t.Fatalf("expected configured timeout, got %v", got.Timeout)
	}
}

func TestValidateLocationServiceBaseURLRejectsUnsafeURL(t *testing.T) {
	for _, baseURL := range []string{
		"ftp://locations.example.test",
		"https://locations.example.test?tenant=city",
		"https://locations.example.test/root#fragment",
	} {
		if err := ValidateLocationServiceBaseURL(baseURL); err == nil {
			t.Fatalf("expected unsafe location service URL %q to fail validation", baseURL)
		}
	}
}
