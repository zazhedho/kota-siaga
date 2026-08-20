package config

import (
	"testing"
	"time"
)

func TestLoadAPIIndonesiaConfigUsesDefaults(t *testing.T) {
	t.Setenv("API_INDONESIA_BASE_URL", "")
	t.Setenv("API_INDONESIA_KEY", "test-api-key")
	t.Setenv("API_INDONESIA_TIMEOUT", "")

	got := LoadAPIIndonesiaConfig()

	if got.BaseURL != "https://use.apiindonesia.id" {
		t.Fatalf("expected default base URL, got %q", got.BaseURL)
	}
	if got.APIKey != "test-api-key" {
		t.Fatalf("expected API key from environment, got %q", got.APIKey)
	}
	if got.Timeout != 5*time.Second {
		t.Fatalf("expected default timeout 5s, got %v", got.Timeout)
	}
}

func TestLoadAPIIndonesiaConfigReadsOverrides(t *testing.T) {
	t.Setenv("API_INDONESIA_BASE_URL", " https://api.example.test/root ")
	t.Setenv("API_INDONESIA_KEY", " test-api-key ")
	t.Setenv("API_INDONESIA_TIMEOUT", "750ms")

	got := LoadAPIIndonesiaConfig()

	if got.BaseURL != "https://api.example.test/root" {
		t.Fatalf("expected configured base URL, got %q", got.BaseURL)
	}
	if got.APIKey != "test-api-key" {
		t.Fatalf("expected configured API key, got %q", got.APIKey)
	}
	if got.Timeout != 750*time.Millisecond {
		t.Fatalf("expected configured timeout, got %v", got.Timeout)
	}
}

func TestValidateAPIIndonesiaBaseURLRejectsQueryAndFragment(t *testing.T) {
	for _, baseURL := range []string{
		"https://api.example.test?tenant=city",
		"https://api.example.test/root#fragment",
		"https://api.example.test?",
	} {
		if err := ValidateAPIIndonesiaBaseURL(baseURL); err == nil {
			t.Fatalf("expected unsafe base URL %q to fail validation", baseURL)
		}
	}
}
