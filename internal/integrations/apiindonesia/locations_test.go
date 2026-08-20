package apiindonesia

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLocationMethodsForwardExactEndpointQueriesAndMapFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/wilayah/provinsi" {
			t.Errorf("expected province path, got %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Errorf("expected page=2, got %q", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "20" {
			t.Errorf("expected per_page=20, got %q", got)
		}
		_, _ = fmt.Fprint(w, `{"data":[{"id":"11","code":"11","name":"Aceh","alt_name":"Aceh Official","lat":4.695135,"lng":96.749397,"is_active":1}],"meta":{"total":1,"page":2,"per_page":20,"total_pages":1}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListProvinces(context.Background(), 2, 20)
	if err != nil {
		t.Fatalf("ListProvinces() error = %v", err)
	}
	if got.Total != 1 || got.Page != 2 || got.PerPage != 20 || len(got.Data) != 1 {
		t.Fatalf("unexpected page: %+v", got)
	}
	province := got.Data[0]
	if province.ID != "11" || province.Code != "11" || province.Name != "Aceh" || province.AlternateName != "Aceh Official" {
		t.Fatalf("unexpected province mapping: %+v", province)
	}
	if province.Latitude != 4.695135 || province.Longitude != 96.749397 || !province.IsActive {
		t.Fatalf("unexpected province coordinates/flag: %+v", province)
	}
}

func TestScopedLocationMethodsUseParentQueryKeys(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		parentKey string
		parentID  string
		call      func(*Client) error
	}{
		{
			name: "city", path: "/api/v1/wilayah/kabupaten", parentKey: "provinsi_id", parentID: "32",
			call: func(client *Client) error { _, err := client.ListCities(context.Background(), "32", 1, 20); return err },
		},
		{
			name: "district", path: "/api/v1/wilayah/kecamatan", parentKey: "kabupaten_id", parentID: "3273",
			call: func(client *Client) error {
				_, err := client.ListDistricts(context.Background(), "3273", 1, 20)
				return err
			},
		},
		{
			name: "village", path: "/api/v1/wilayah/kelurahan", parentKey: "kecamatan_id", parentID: "3273010",
			call: func(client *Client) error {
				_, err := client.ListVillages(context.Background(), "3273010", 1, 20)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("expected path %q, got %q", tt.path, r.URL.Path)
				}
				if got := r.URL.Query().Get(tt.parentKey); got != tt.parentID {
					t.Errorf("expected %s=%s, got %q", tt.parentKey, tt.parentID, got)
				}
				if got := r.URL.Query().Get("page"); got != "1" {
					t.Errorf("expected page=1, got %q", got)
				}
				if got := r.URL.Query().Get("per_page"); got != "20" {
					t.Errorf("expected per_page=20, got %q", got)
				}
				_, _ = fmt.Fprint(w, `{"data":[],"meta":{"total":0,"page":1,"per_page":20,"total_pages":0}}`)
			}))

			client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
			if err != nil {
				server.Close()
				t.Fatalf("NewClient() error = %v", err)
			}
			if err := tt.call(client); err != nil {
				t.Fatalf("location call error = %v", err)
			}
			server.Close()
		})
	}
}

func TestLocationMethodsDecodeStringIDsWithoutLosingLeadingZeros(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body string
		switch r.URL.Path {
		case "/api/v1/wilayah/provinsi":
			body = `{"data":[{"id":"00011","code":"11","name":"Aceh","alt_name":"Aceh Official","lat":4.6,"lng":96.7,"is_active":1}],"meta":{"total":1,"page":1,"per_page":20,"total_pages":1}}`
		case "/api/v1/wilayah/kabupaten":
			body = `{"data":[{"id":"0003201","province_id":"00032","code":"3201","name":"Bogor","alt_name":"Bogor Official","is_city":1,"lat":-6.6,"lng":106.8,"is_active":1}],"meta":{"total":1,"page":1,"per_page":20,"total_pages":1}}`
		case "/api/v1/wilayah/kecamatan":
			body = `{"data":[{"id":"000327301","regency_id":"0003273","code":"327301","name":"Bogor Selatan","alt_name":"Bogor Selatan Official","lat":-6.6,"lng":106.8,"is_active":1}],"meta":{"total":1,"page":1,"per_page":20,"total_pages":1}}`
		case "/api/v1/wilayah/kelurahan":
			body = `{"data":[{"id":"0003273010001","district_id":"0003273010","code":"3273010001","name":"Bondongan","alt_name":"Bondongan Official","postal_code":"00123","is_courier_support":1,"lat":-6.6,"lng":106.8,"is_active":1}],"meta":{"total":1,"page":1,"per_page":20,"total_pages":1}}`
		default:
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, body)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	provinces, err := client.ListProvinces(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("ListProvinces() error = %v", err)
	}
	assertStringLocationID(t, provinces.Data[0].ID, "00011")
	if provinces.Data[0].Code != "11" || provinces.Data[0].Name != "Aceh" || provinces.Data[0].AlternateName != "Aceh Official" {
		t.Fatalf("province fields changed: %+v", provinces.Data[0])
	}

	cities, err := client.ListCities(context.Background(), "00032", 1, 20)
	if err != nil {
		t.Fatalf("ListCities() error = %v", err)
	}
	assertStringLocationID(t, cities.Data[0].ID, "0003201")
	assertStringLocationID(t, cities.Data[0].ProvinceID, "00032")
	if cities.Data[0].Code != "3201" || cities.Data[0].Name != "Bogor" || cities.Data[0].AlternateName != "Bogor Official" {
		t.Fatalf("city fields changed: %+v", cities.Data[0])
	}

	districts, err := client.ListDistricts(context.Background(), "0003273", 1, 20)
	if err != nil {
		t.Fatalf("ListDistricts() error = %v", err)
	}
	assertStringLocationID(t, districts.Data[0].ID, "000327301")
	assertStringLocationID(t, districts.Data[0].RegencyID, "0003273")
	if districts.Data[0].Code != "327301" || districts.Data[0].Name != "Bogor Selatan" || districts.Data[0].AlternateName != "Bogor Selatan Official" {
		t.Fatalf("district fields changed: %+v", districts.Data[0])
	}

	villages, err := client.ListVillages(context.Background(), "0003273010", 1, 20)
	if err != nil {
		t.Fatalf("ListVillages() error = %v", err)
	}
	assertStringLocationID(t, villages.Data[0].ID, "0003273010001")
	assertStringLocationID(t, villages.Data[0].DistrictID, "0003273010")
	if villages.Data[0].Code != "3273010001" || villages.Data[0].Name != "Bondongan" || villages.Data[0].AlternateName != "Bondongan Official" || villages.Data[0].PostalCode != "00123" || !villages.Data[0].IsCourierSupport || !villages.Data[0].IsActive {
		t.Fatalf("village fields changed: %+v", villages.Data[0])
	}
}

func assertStringLocationID(t *testing.T, value any, want string) {
	t.Helper()
	got, ok := value.(string)
	if !ok || got != want {
		t.Fatalf("location ID = %#v (%T), want string %q", value, value, want)
	}
}
