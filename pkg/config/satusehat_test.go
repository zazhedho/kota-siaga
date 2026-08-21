package config

import (
	"testing"
	"time"
)

func TestLoadSATUSEHATConfigUsesProductionDefaults(t *testing.T) {
	t.Setenv("SATUSEHAT_BASE_URL", "")
	t.Setenv("SATUSEHAT_AUTH_BASE_URL", "")
	t.Setenv("SATUSEHAT_CLIENT_ID", "client-id")
	t.Setenv("SATUSEHAT_CLIENT_SECRET", "client-secret")
	t.Setenv("SATUSEHAT_TIMEOUT", "")

	got := LoadSATUSEHATConfig()

	if got.BaseURL != "https://api-satusehat.kemkes.go.id/masterdata" {
		t.Fatalf("expected production master data URL, got %q", got.BaseURL)
	}
	if got.AuthBaseURL != "https://api-satusehat.kemkes.go.id/oauth2/v1" {
		t.Fatalf("expected production auth URL, got %q", got.AuthBaseURL)
	}
	if got.ClientID != "client-id" || got.ClientSecret != "client-secret" {
		t.Fatalf("expected SATUSEHAT credentials from environment")
	}
	if got.Timeout != 5*time.Second {
		t.Fatalf("expected default timeout 5s, got %v", got.Timeout)
	}
}

func TestLoadSATUSEHATConfigReadsOverrides(t *testing.T) {
	t.Setenv("SATUSEHAT_BASE_URL", " https://satusehat.example.test/masterdata ")
	t.Setenv("SATUSEHAT_AUTH_BASE_URL", " https://satusehat.example.test/oauth2/v1 ")
	t.Setenv("SATUSEHAT_CLIENT_ID", " client-id ")
	t.Setenv("SATUSEHAT_CLIENT_SECRET", " client-secret ")
	t.Setenv("SATUSEHAT_TIMEOUT", "750ms")

	got := LoadSATUSEHATConfig()

	if got.BaseURL != "https://satusehat.example.test/masterdata" || got.AuthBaseURL != "https://satusehat.example.test/oauth2/v1" {
		t.Fatalf("expected configured SATUSEHAT URLs, got master=%q auth=%q", got.BaseURL, got.AuthBaseURL)
	}
	if got.ClientID != "client-id" || got.ClientSecret != "client-secret" {
		t.Fatalf("expected trimmed SATUSEHAT credentials")
	}
	if got.Timeout != 750*time.Millisecond {
		t.Fatalf("expected configured timeout, got %v", got.Timeout)
	}
}
