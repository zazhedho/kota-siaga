package apiindonesia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/mappers"
)

func TestListHospitalsForwardsExactEndpointQueryAndMapsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/rumah-sakit" {
			t.Errorf("expected hospital path, got %q", r.URL.Path)
		}
		for key, want := range map[string]string{
			"kabupaten_id": "003273",
			"page":         "2",
			"per_page":     "20",
		} {
			if got := r.URL.Query().Get(key); got != want {
				t.Errorf("expected %s=%q, got %q", key, want, got)
			}
		}
		if got := r.Header.Get("x-api-key"); got != "hospital-test-key" {
			t.Errorf("expected API key header, got %q", got)
		}
		_, _ = fmt.Fprint(w, `{"data":[{"id":"0001","name":"RS Official","jenis":"Rumah Sakit Umum","kelas":"B","ownership":"Pemerintah","address":"Jalan Official","postal_code":"00123","phone":"021-555000","beds_total":120,"icu_beds":12,"province_name":"Jawa Barat","regency_name":"Bogor","is_active":1,"province_id":"32","regency_id":"3273","is_blu":0}],"meta":{"total":1,"page":2,"per_page":20,"total_pages":1}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, APIKey: "hospital-test-key", Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListHospitals(context.Background(), "003273", 2, 20)
	if err != nil {
		t.Fatalf("ListHospitals() error = %v", err)
	}
	if got.Total != 1 || got.Page != 2 || got.PerPage != 20 || got.TotalPages != 1 || len(got.Data) != 1 {
		t.Fatalf("unexpected hospital page: %+v", got)
	}
	want := dto.Hospital{
		ID: "0001", Name: "RS Official", Type: "Rumah Sakit Umum", Class: "B", Ownership: "Pemerintah",
		Address: "Jalan Official", PostalCode: "00123", Phone: "021-555000", BedsTotal: 120, ICUBeds: 12,
		ProvinceName: "Jawa Barat", RegencyName: "Bogor", IsActive: true,
	}
	if got.Data[0] != want {
		t.Fatalf("unexpected hospital mapping: got=%+v want=%+v", got.Data[0], want)
	}
}

func TestMapHospitalPreservesApplicationFields(t *testing.T) {
	got := mappers.MapHospital(mappers.HospitalSource{
		ID: "1", Name: "Name", Jenis: "Type", Kelas: "Class", Ownership: "Ownership",
		Address: "Address", PostalCode: "12345", Phone: "Phone", BedsTotal: 10, ICUBeds: 2,
		ProvinceName: "Province", RegencyName: "Regency", IsActive: 1,
	})
	want := dto.Hospital{
		ID: "1", Name: "Name", Type: "Type", Class: "Class", Ownership: "Ownership",
		Address: "Address", PostalCode: "12345", Phone: "Phone", BedsTotal: 10, ICUBeds: 2,
		ProvinceName: "Province", RegencyName: "Regency", IsActive: true,
	}
	if got != want {
		t.Fatalf("unexpected mapped hospital: got=%+v want=%+v", got, want)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal hospital: %v", err)
	}
	body := string(encoded)
	for _, field := range []string{"id", "name", "type", "class", "ownership", "address", "postal_code", "phone", "beds_total", "icu_beds", "province_name", "regency_name", "is_active"} {
		if !strings.Contains(body, `"`+field+`"`) {
			t.Fatalf("hospital JSON missing application field %q: %s", field, body)
		}
	}
	for _, field := range []string{"jenis", "kelas", "province_id", "regency_id", "is_blu"} {
		if strings.Contains(body, `"`+field+`"`) {
			t.Fatalf("hospital JSON exposed source field %q: %s", field, body)
		}
	}
}

func TestListHospitalsReturnsErrorForMalformedSourceField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":[{"id":"1","beds_total":"not-a-number"}],"meta":{"total":1,"page":1,"per_page":20,"total_pages":1}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if _, err := client.ListHospitals(context.Background(), "3273", 1, 20); err == nil {
		t.Fatal("expected malformed hospital response error")
	}
}
