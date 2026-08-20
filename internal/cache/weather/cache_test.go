package weathercache

import (
	"testing"
	"time"
)

func TestWeatherKeyUsesADM4(t *testing.T) {
	if got := Key("32.73.01.1001"); got != "weather:32.73.01.1001" {
		t.Fatalf("unexpected weather cache key: %q", got)
	}
}

func TestWeatherTTLDefaultsToSixHoursAndReadsValidOverride(t *testing.T) {
	t.Setenv("WEATHER_CACHE_TTL", "")
	if got := TTL(); got != 6*time.Hour {
		t.Fatalf("expected six-hour default TTL, got %v", got)
	}

	t.Setenv("WEATHER_CACHE_TTL", "90m")
	if got := TTL(); got != 90*time.Minute {
		t.Fatalf("expected configured TTL, got %v", got)
	}

	for _, value := range []string{"invalid", "0", "-1s"} {
		t.Setenv("WEATHER_CACHE_TTL", value)
		if got := TTL(); got != 6*time.Hour {
			t.Fatalf("expected invalid TTL %q to use six-hour default, got %v", value, got)
		}
	}
}
