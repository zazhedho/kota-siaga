package config

import (
	"strings"
	"testing"
)

func TestValidateStartupConfigAcceptsPublicRuntime(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "aip_live_test")
	t.Setenv("API_INDONESIA_BASE_URL", "https://use.apiindonesia.id")

	if err := ValidateStartupConfig("8080"); err != nil {
		t.Fatalf("expected public runtime config to pass: %v", err)
	}
}

func TestValidateStartupConfigUsesDefaultAPIBaseURLWhenBlank(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "aip_live_test")
	t.Setenv("API_INDONESIA_BASE_URL", "")

	if err := ValidateStartupConfig("8080"); err != nil {
		t.Fatalf("expected blank API base URL to use default: %v", err)
	}
}

func TestValidateStartupConfigRequiresAPIKey(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "")

	if err := ValidateStartupConfig("8080"); err == nil {
		t.Fatal("expected missing API key to fail")
	}
}

func TestValidateStartupConfigRejectsNonHTTPAPIBaseURL(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "aip_live_test")
	t.Setenv("API_INDONESIA_BASE_URL", "ftp://api.example.com")

	err := ValidateStartupConfig("8080")
	if err == nil || !strings.Contains(err.Error(), "API_INDONESIA_BASE_URL must be a valid URL") {
		t.Fatalf("expected non-HTTP API base URL to fail validation, got %v", err)
	}
}

func TestValidateStartupConfigRejectsAPIBaseURLQueryOrFragment(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "aip_live_test")
	for _, baseURL := range []string{
		"https://api.example.test?tenant=city",
		"https://api.example.test/root#fragment",
	} {
		t.Setenv("API_INDONESIA_BASE_URL", baseURL)
		if err := ValidateStartupConfig("8080"); err == nil {
			t.Fatalf("expected unsafe API base URL %q to fail", baseURL)
		}
	}
}

func TestValidateStartupConfigRejectsBMKGBaseURL(t *testing.T) {
	t.Setenv("API_INDONESIA_KEY", "aip_live_test")
	t.Setenv("API_INDONESIA_BASE_URL", "https://use.apiindonesia.id")
	t.Setenv("BMKG_BASE_URL", "ftp://data.bmkg.go.id")

	err := ValidateStartupConfig("8080")
	if err == nil || !strings.Contains(err.Error(), "BMKG_BASE_URL must be a valid URL") {
		t.Fatalf("expected invalid BMKG base URL to fail, got %v", err)
	}
}
