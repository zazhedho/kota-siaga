package apiindonesia

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"kota-siaga/internal/dto"
)

func TestListWarningsUsesExactEndpointQueryAndMapsFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/peringatan-dini" {
			t.Errorf("expected /api/v1/peringatan-dini, got %q", r.URL.Path)
		}
		if r.URL.RawQuery != "provinsi=Jawa+Barat" {
			t.Errorf("expected exact province query, got %q", r.URL.RawQuery)
		}
		_, _ = fmt.Fprint(w, `{"data":[{"id":"warning-1","alert_id":"alert-1","event":"Heavy Rain","urgency":"Immediate","severity":"Severe","certainty":"Observed","area":"Bogor Selatan","province":"Jawa Barat","effective":"2026-08-20T10:00:00+07:00","expires":"20 Aug 2026 16:00 WIB","headline":"Official warning headline","description":"Official warning description","instruction":"Official warning instruction","source":"BMKG","is_active":1},{"id":"warning-2","alert_id":"alert-2","event":"Strong Wind","urgency":"Future","severity":"Moderate","certainty":"Possible","area":"Bogor","province":"Jawa Barat","effective":"2026-08-21","expires":"2026-08-21","headline":"Second headline","description":"Second description","instruction":"Second instruction","source":"BMKG","is_active":0}]}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListWarnings(context.Background(), "Jawa Barat")
	if err != nil {
		t.Fatalf("ListWarnings() error = %v", err)
	}

	want := []dto.Warning{
		{
			ID:          "warning-1",
			AlertID:     "alert-1",
			Event:       "Heavy Rain",
			Urgency:     "Immediate",
			Severity:    "Severe",
			Certainty:   "Observed",
			Area:        "Bogor Selatan",
			Province:    "Jawa Barat",
			Effective:   "2026-08-20T10:00:00+07:00",
			Expires:     "20 Aug 2026 16:00 WIB",
			Headline:    "Official warning headline",
			Description: "Official warning description",
			Instruction: "Official warning instruction",
			Source:      "BMKG",
			IsActive:    true,
		},
		{
			ID:          "warning-2",
			AlertID:     "alert-2",
			Event:       "Strong Wind",
			Urgency:     "Future",
			Severity:    "Moderate",
			Certainty:   "Possible",
			Area:        "Bogor",
			Province:    "Jawa Barat",
			Effective:   "2026-08-21",
			Expires:     "2026-08-21",
			Headline:    "Second headline",
			Description: "Second description",
			Instruction: "Second instruction",
			Source:      "BMKG",
			IsActive:    false,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized warnings:\n got: %#v\nwant: %#v", got, want)
	}
}
