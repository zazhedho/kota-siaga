package bmkg

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"kota-siaga/internal/dto"
	"kota-siaga/pkg/config"
)

func TestListLatestUsesAutogempaEndpointAndMapsFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/DataMKG/TEWS/autogempa.json" {
			t.Errorf("expected /DataMKG/TEWS/autogempa.json, got %q", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query parameters, got %q", r.URL.RawQuery)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("expected JSON Accept header, got %q", got)
		}
		_, _ = fmt.Fprint(w, `{"Infogempa":{"gempa":{"Tanggal":"21 Agu 2026","Jam":"08:18:05 WIB","DateTime":"2026-08-21T01:18:05+00:00","Coordinates":"-8.23,120.62","Lintang":"8.23 LS","Bujur":"120.62 BT","Magnitude":"4.6","Kedalaman":"7 km","Wilayah":"Pusat gempa berada di laut 45 km Utara Ruteng-Manggarai","Potensi":"Gempa ini dirasakan untuk diteruskan pada masyarakat","Dirasakan":"II-III MMI Kab. Manggarai","Shakemap":"20260821081805.mmi.jpg"}}}`)
	}))
	defer server.Close()

	client, err := NewClient(config.BMKGConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListLatest(context.Background())
	if err != nil {
		t.Fatalf("ListLatest() error = %v", err)
	}

	feltAreas := "II-III MMI Kab. Manggarai"
	want := []dto.Earthquake{{
		ID:        "bmkg:2026-08-21T01:18:05+00:00",
		DateTime:  "2026-08-21T01:18:05+00:00",
		Magnitude: 4.6,
		DepthKM:   7,
		Latitude:  -8.23,
		Longitude: 120.62,
		Region:    "Pusat gempa berada di laut 45 km Utara Ruteng-Manggarai",
		Potential: "Gempa ini dirasakan untuk diteruskan pada masyarakat",
		IsFelt:    true,
		FeltAreas: &feltAreas,
		Source:    "BMKG",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized earthquake:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestListLatestRejectsMalformedOrIncompletePayload(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: "not-json"},
		{name: "missing event", body: `{"Infogempa":{"gempa":{}}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, test.body)
			}))
			defer server.Close()

			client, err := NewClient(config.BMKGConfig{BaseURL: server.URL, Timeout: time.Second})
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			if _, err := client.ListLatest(context.Background()); err == nil {
				t.Fatal("expected malformed or incomplete payload error")
			}
		})
	}
}
