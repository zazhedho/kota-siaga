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

func TestListLatestUsesExactEndpointWithoutQueryAndMapsFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/gempa/terkini" {
			t.Errorf("expected /api/v1/gempa/terkini, got %q", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query parameters, got %q", r.URL.RawQuery)
		}
		_, _ = fmt.Fprint(w, `{"data":[{"id":"earthquake-1","datetime":"2026-08-20T03:15:00Z","magnitude":5.4,"depth_km":10.5,"lat":-6.1234,"lng":106.9876,"region":"South of Java","potential":"No tsunami potential","is_felt":1,"felt_areas":null,"source":"BMKG"},{"id":"earthquake-2","datetime":"2026-08-20T02:00:00Z","magnitude":2.1,"depth_km":4,"lat":-7.25,"lng":110.5,"region":"Central Java","potential":"No potential","is_felt":0,"felt_areas":"II-III Yogyakarta","source":"BMKG"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListLatest(context.Background())
	if err != nil {
		t.Fatalf("ListLatest() error = %v", err)
	}

	want := []dto.Earthquake{
		{
			ID:        "earthquake-1",
			DateTime:  "2026-08-20T03:15:00Z",
			Magnitude: 5.4,
			DepthKM:   10.5,
			Latitude:  -6.1234,
			Longitude: 106.9876,
			Region:    "South of Java",
			Potential: "No tsunami potential",
			IsFelt:    true,
			FeltAreas: nil,
			Source:    "BMKG",
		},
		{
			ID:        "earthquake-2",
			DateTime:  "2026-08-20T02:00:00Z",
			Magnitude: 2.1,
			DepthKM:   4,
			Latitude:  -7.25,
			Longitude: 110.5,
			Region:    "Central Java",
			Potential: "No potential",
			IsFelt:    false,
			FeltAreas: func() *string { value := "II-III Yogyakarta"; return &value }(),
			Source:    "BMKG",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized earthquakes:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestListLatestHonorsCanceledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":[]}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ListLatest(ctx); err == nil {
		t.Fatal("expected canceled request error")
	}
}
