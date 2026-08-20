package warningcache

import (
	"testing"
	"time"
)

func TestKeyTrimsAndLowercasesProvince(t *testing.T) {
	if got := Key("  Jawa Barat  "); got != "warning:jawa barat" {
		t.Fatalf("unexpected warning cache key: %q", got)
	}
}

func TestTTLDefaultsToOneHourAndUsesOnlyPositiveOverride(t *testing.T) {
	t.Setenv("WARNING_CACHE_TTL", "")
	if got := TTL(); got != time.Hour {
		t.Fatalf("expected one-hour default TTL, got %v", got)
	}

	t.Setenv("WARNING_CACHE_TTL", "90m")
	if got := TTL(); got != 90*time.Minute {
		t.Fatalf("expected configured TTL, got %v", got)
	}

	for _, value := range []string{"invalid", "0", "-1s"} {
		t.Setenv("WARNING_CACHE_TTL", value)
		if got := TTL(); got != time.Hour {
			t.Fatalf("expected invalid TTL %q to use one-hour default, got %v", value, got)
		}
	}
}
