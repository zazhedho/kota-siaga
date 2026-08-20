package hospitalcache

import (
	"testing"
	"time"
)

func TestHospitalKeyIncludesAllInputs(t *testing.T) {
	if got := Key("003273", 2, 20); got != "hospital:003273:2:20" {
		t.Fatalf("unexpected hospital cache key: %q", got)
	}
	if Key("3273", 1, 20) == Key("3274", 1, 20) {
		t.Fatal("different kabupaten IDs shared a cache key")
	}
	if Key("3273", 1, 20) == Key("3273", 2, 20) {
		t.Fatal("different pages shared a cache key")
	}
	if Key("3273", 1, 20) == Key("3273", 1, 21) {
		t.Fatal("different page sizes shared a cache key")
	}
}

func TestHospitalTTLDefaultsTo24HoursAndUsesOnlyPositiveOverride(t *testing.T) {
	t.Setenv("HOSPITAL_CACHE_TTL", "")
	if got := TTL(); got != 24*time.Hour {
		t.Fatalf("expected 24-hour default TTL, got %v", got)
	}

	t.Setenv("HOSPITAL_CACHE_TTL", "90m")
	if got := TTL(); got != 90*time.Minute {
		t.Fatalf("expected configured TTL, got %v", got)
	}

	for _, value := range []string{"invalid", "0", "-1s"} {
		t.Setenv("HOSPITAL_CACHE_TTL", value)
		if got := TTL(); got != 24*time.Hour {
			t.Fatalf("expected invalid TTL %q to use 24-hour default, got %v", value, got)
		}
	}
}
