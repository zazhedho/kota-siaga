package locationcache

import (
	"testing"
	"time"
)

func TestLocationKeysIncludeAllInputs(t *testing.T) {
	if got := ProvinceKey(2, 20); got != "locations:province:2:20" {
		t.Fatalf("unexpected province key: %q", got)
	}
	if got := CityKey("32", 2, 20); got != "locations:city:32:2:20" {
		t.Fatalf("unexpected city key: %q", got)
	}
	if got := DistrictKey("3273", 2, 20); got != "locations:district:3273:2:20" {
		t.Fatalf("unexpected district key: %q", got)
	}
	if got := VillageKey("3273010", 2, 20); got != "locations:village:3273010:2:20" {
		t.Fatalf("unexpected village key: %q", got)
	}
}

func TestLocationTTLDefaultsTo24HoursAndReadsOverride(t *testing.T) {
	t.Setenv("LOCATION_CACHE_TTL", "")
	if got := TTL(); got != 24*time.Hour {
		t.Fatalf("expected 24-hour default TTL, got %v", got)
	}

	t.Setenv("LOCATION_CACHE_TTL", "45m")
	if got := TTL(); got != 45*time.Minute {
		t.Fatalf("expected configured TTL, got %v", got)
	}

	t.Setenv("LOCATION_CACHE_TTL", "-1s")
	if got := TTL(); got != 24*time.Hour {
		t.Fatalf("expected invalid TTL fallback, got %v", got)
	}
}
