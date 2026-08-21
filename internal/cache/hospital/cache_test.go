package hospitalcache

import (
	"testing"
	"time"
)

func TestHospitalKeyIncludesAllInputs(t *testing.T) {
	if got := Key("003273", "", 2, 20); got != "hospital:003273::2:20" {
		t.Fatalf("unexpected hospital cache key: %q", got)
	}
	if Key("3273", "", 1, 20) == Key("3274", "", 1, 20) {
		t.Fatal("different kabupaten IDs shared a cache key")
	}
	if Key("3273", "", 1, 20) == Key("3273", "", 2, 20) {
		t.Fatal("different pages shared a cache key")
	}
	if Key("3273", "", 1, 20) == Key("3273", "", 1, 21) {
		t.Fatal("different page sizes shared a cache key")
	}
}

func TestHospitalKeyNormalizesAndEscapesSearch(t *testing.T) {
	if got := Key("003273", "  Hasan Sadikin ", 1, 20); got != "hospital:search:003273:hasan+sadikin:1:20" {
		t.Fatalf("unexpected normalized hospital cache key: %q", got)
	}
	if Key("3273", " Hasan ", 1, 20) != Key("3273", "hasan", 1, 20) {
		t.Fatal("equivalent searches used different cache keys")
	}
	if Key("3273", "hasan", 1, 20) == Key("3273", "hasan sadikin", 1, 20) {
		t.Fatal("different searches shared a cache key")
	}
	if got := Key("3273", "A+B/C", 1, 20); got != "hospital:search:3273:a%2Bb%2Fc:1:20" {
		t.Fatalf("unexpected escaped hospital cache key: %q", got)
	}
}

func TestHospitalSearchKeyUsesDistinctNamespaceFromUpstreamNameSearch(t *testing.T) {
	const previousUpstreamNameSearchKey = "hospital:3273:hasan:1:20"

	got := Key("3273", " HASAN ", 1, 20)
	if got == previousUpstreamNameSearchKey {
		t.Fatalf("non-empty search reused previous upstream-name-search cache key: %q", got)
	}
	if got != Key("3273", "hasan", 1, 20) {
		t.Fatalf("equivalent normalized searches used different cache keys: %q", got)
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
