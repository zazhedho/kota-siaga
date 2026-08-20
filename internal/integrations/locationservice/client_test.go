package locationservice

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kota-siaga/pkg/config"
)

func TestClientUsesProviderPathsAndMapsDottedCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body string
		switch r.URL.Path {
		case "/api/locations/provinces":
			if len(r.URL.Query()) != 0 {
				t.Errorf("province request must not need a query, got %q", r.URL.RawQuery)
			}
			body = `{"data":[{"code":"32","full_code":"32","name":"Jawa Barat","level":"province"}]}`
		case "/api/locations/regencies":
			if got := r.URL.Query().Get("province_code"); got != "32" {
				t.Errorf("expected dotted province code 32, got %q", got)
			}
			body = `{"data":[{"code":"32.73","full_code":"32.73","name":"Kota Bandung","level":"regency","parent_code":"32"}]}`
		case "/api/locations/districts":
			if got := r.URL.Query().Get("regency_code"); got != "32.73" {
				t.Errorf("expected dotted regency code 32.73, got %q", got)
			}
			body = `{"data":[{"code":"32.73.01","full_code":"32.73.01","name":"Sukasari","level":"district","parent_code":"32.73"}]}`
		case "/api/locations/villages":
			if got := r.URL.Query().Get("district_code"); got != "32.73.01" {
				t.Errorf("expected dotted district code 32.73.01, got %q", got)
			}
			body = `{"data":[{"code":"32.73.01.1001","full_code":"32.73.01.1001","name":"Sukarasa","level":"village","parent_code":"32.73.01","postal_code":"40152"}]}`
		default:
			t.Errorf("unexpected provider path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = fmt.Fprint(w, body)
	}))
	defer server.Close()

	client, err := NewClient(config.LocationServiceConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	province, err := client.ListProvinces(ctx, 1, 100)
	if err != nil {
		t.Fatalf("ListProvinces() error = %v", err)
	}
	if province.Data[0].ID != "32" || province.Data[0].Code != "32" || !province.Data[0].IsActive {
		t.Fatalf("unexpected province mapping: %+v", province.Data[0])
	}

	city, err := client.ListCities(ctx, "32", 1, 100)
	if err != nil {
		t.Fatalf("ListCities() error = %v", err)
	}
	if city.Data[0].ID != "3273" || city.Data[0].ProvinceID != "32" {
		t.Fatalf("unexpected city mapping: %+v", city.Data[0])
	}

	district, err := client.ListDistricts(ctx, "3273", 1, 100)
	if err != nil {
		t.Fatalf("ListDistricts() error = %v", err)
	}
	if district.Data[0].ID != "327301" || district.Data[0].RegencyID != "3273" {
		t.Fatalf("unexpected district mapping: %+v", district.Data[0])
	}

	village, err := client.ListVillages(ctx, "327301", 1, 100)
	if err != nil {
		t.Fatalf("ListVillages() error = %v", err)
	}
	got := village.Data[0]
	if got.ID != "3273011001" || got.DistrictID != "327301" || got.Code != "32.73.01.1001" || got.PostalCode != "40152" {
		t.Fatalf("unexpected village mapping: %+v", got)
	}
}

func TestClientPaginatesProviderCollectionLocally(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/locations/provinces" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_, _ = fmt.Fprint(w, `{"data":[{"code":"11","name":"Aceh"},{"code":"12","name":"Sumatera Utara"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(config.LocationServiceConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListProvinces(context.Background(), 2, 1)
	if err != nil {
		t.Fatalf("ListProvinces() error = %v", err)
	}
	if got.Total != 2 || got.Page != 2 || got.PerPage != 1 || got.TotalPages != 2 || len(got.Data) != 1 || got.Data[0].ID != "12" {
		t.Fatalf("unexpected local page: %+v", got)
	}
}

func TestDottedCodePreservesAlreadyDottedValues(t *testing.T) {
	for _, code := range []string{"32", "32.73", "32.73.01", "32.73.01.1001"} {
		if got := dottedCode(code); got != code {
			t.Fatalf("dottedCode(%q) = %q", code, got)
		}
	}
}

func TestDottedCodeConvertsCompactValues(t *testing.T) {
	for _, test := range []struct {
		compact string
		want    string
	}{
		{compact: "32", want: "32"},
		{compact: "3273", want: "32.73"},
		{compact: "327301", want: "32.73.01"},
		{compact: "3273010", want: "32.73.01"},
	} {
		if got := dottedCode(test.compact); got != test.want {
			t.Fatalf("dottedCode(%q) = %q, want %q", test.compact, got, test.want)
		}
	}
}
