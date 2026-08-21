package satusehat

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kota-siaga/internal/dto"
	"kota-siaga/pkg/config"
)

func TestListHospitalsAuthenticatesOnceAndMapsMasterFacilityPage(t *testing.T) {
	authCalls := 0
	hospitalCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/v1/accesstoken":
			authCalls++
			if r.Method != http.MethodPost {
				t.Errorf("expected token POST, got %s", r.Method)
			}
			if got := r.URL.Query().Get("grant_type"); got != "client_credentials" {
				t.Errorf("expected client_credentials grant, got %q", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse token form: %v", err)
			}
			if r.Form.Get("client_id") != "client-id" || r.Form.Get("client_secret") != "client-secret" {
				t.Errorf("unexpected token credentials")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"access_token":"access-token","token_type":"BearerToken","expires_in":"3600"}`)
		case "/masterdata/v1/mastersaranaindex/mastersarana":
			hospitalCalls++
			if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
				t.Errorf("expected bearer token, got %q", got)
			}
			expectedLimit, expectedPage := "20", "2"
			if hospitalCalls == 1 {
				expectedLimit, expectedPage = "2000", "1"
			}
			for key, want := range map[string]string{
				"limit":        expectedLimit,
				"page":         expectedPage,
				"jenis_sarana": "104",
				"kode_kabkota": "3273",
			} {
				if got := r.URL.Query().Get(key); got != want {
					t.Errorf("expected %s=%q, got %q", key, want, got)
				}
			}
			if _, ok := r.URL.Query()["nama"]; ok {
				t.Errorf("expected nama to be absent for local and empty searches, got %q", r.URL.Query().Get("nama"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{
  "status_code": 200,
  "message": "Success",
  "page": 1,
  "total_page": 1,
  "data": [{
    "kode_satusehat": "1000123456",
    "kode_sarana": "3273001",
    "nama": "Rumah Sakit Kota",
    "telp": "021-555000",
    "alamat": "Jalan Kesehatan No. 1",
    "provinsi": {"kode": "32", "nama": "Jawa Barat"},
    "kabkota": {"kode": "3273", "nama": "Kota Bandung"},
    "jenis_sarana": {"kode": "104", "nama": "Rumah Sakit"},
    "subjenis": {"kode": "10401", "nama": "Rumah Sakit Umum"},
    "kelas_sarana": {"kode": "B", "nama": "Kelas B"},
    "status_aktif": true
  }]
}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(config.SATUSEHATConfig{
		BaseURL:      server.URL + "/masterdata",
		AuthBaseURL:  server.URL + "/oauth2/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListHospitals(context.Background(), "003273", "  Rumah Sakit  ", 1, 20)
	if err != nil {
		t.Fatalf("ListHospitals() error = %v", err)
	}
	if _, err := client.ListHospitals(context.Background(), "003273", "", 2, 20); err != nil {
		t.Fatalf("second ListHospitals() error = %v", err)
	}
	if authCalls != 1 || hospitalCalls != 2 {
		t.Fatalf("expected one auth and two hospital requests, got auth=%d hospitals=%d", authCalls, hospitalCalls)
	}
	if got.Total != 1 || got.Page != 1 || got.PerPage != 20 || got.TotalPages != 1 || len(got.Data) != 1 {
		t.Fatalf("unexpected hospital page: %+v", got)
	}
	want := dto.Hospital{
		ID: "1000123456", Name: "Rumah Sakit Kota", Type: "Rumah Sakit Umum", Class: "Kelas B",
		Address: "Jalan Kesehatan No. 1", Phone: "021-555000", ProvinceName: "Jawa Barat",
		RegencyName: "Kota Bandung", IsActive: true,
	}
	if got.Data[0] != want {
		t.Fatalf("unexpected hospital mapping: got=%+v want=%+v", got.Data[0], want)
	}
}

func TestListHospitalsSupportsCaseInsensitivePartialSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/v1/accesstoken":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"access_token":"access-token","expires_in":"3600"}`)
		case "/masterdata/v1/mastersaranaindex/mastersarana":
			for key, want := range map[string]string{
				"limit":        "2000",
				"page":         "1",
				"jenis_sarana": "104",
				"kode_kabkota": "3273",
			} {
				if got := r.URL.Query().Get(key); got != want {
					t.Errorf("expected %s=%q, got %q", key, want, got)
				}
			}
			if got := r.URL.Query().Get("nama"); got != "" {
				t.Errorf("expected no upstream nama filter for local partial search, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{
  "page": 1,
  "total_page": 1,
  "data": [
    {
      "kode_satusehat": "1000123456",
      "nama": "RS Brawijaya Saharjo",
      "kabkota": {"kode": "3273", "nama": "Kota Jakarta Selatan"},
      "jenis_sarana": {"kode": "104", "nama": "Rumah Sakit"},
      "status_aktif": true
    },
    {
      "kode_satusehat": "1000654321",
      "nama": "RS Hasan Sadikin",
      "kabkota": {"kode": "3273", "nama": "Kota Jakarta Selatan"},
      "jenis_sarana": {"kode": "104", "nama": "Rumah Sakit"},
      "status_aktif": true
    }
  ]
}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(config.SATUSEHATConfig{
		BaseURL:      server.URL + "/masterdata",
		AuthBaseURL:  server.URL + "/oauth2/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListHospitals(context.Background(), "3273", "bra", 1, 20)
	if err != nil {
		t.Fatalf("ListHospitals() error = %v", err)
	}
	if got.Total != 1 || got.Page != 1 || got.PerPage != 20 || got.TotalPages != 1 || len(got.Data) != 1 || got.Data[0].Name != "RS Brawijaya Saharjo" {
		t.Fatalf("expected one case-insensitive partial match page, got %+v", got)
	}
}

func TestListHospitalsSearchesAllUpstreamPagesBeforePublicPagination(t *testing.T) {
	hospitalCalls := 0
	pages := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/v1/accesstoken":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"access_token":"access-token","expires_in":"3600"}`)
		case "/masterdata/v1/mastersaranaindex/mastersarana":
			hospitalCalls++
			page := r.URL.Query().Get("page")
			pages = append(pages, page)
			if got := r.URL.Query().Get("limit"); got != "2000" {
				t.Errorf("expected limit=2000, got %q", got)
			}
			if got := r.URL.Query().Get("jenis_sarana"); got != "104" {
				t.Errorf("expected jenis_sarana=104, got %q", got)
			}
			if got := r.URL.Query().Get("kode_kabkota"); got != "3273" {
				t.Errorf("expected kode_kabkota=3273, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			switch page {
			case "1":
				_, _ = fmt.Fprint(w, `{
  "page": 1,
  "total_page": 2,
  "data": [
    {"kode_satusehat": "1000123456", "nama": "RS Hasan Sadikin"},
    {"kode_satusehat": "1000654321", "nama": "RS Brawijaya"}
  ]
}`)
			case "2":
				_, _ = fmt.Fprint(w, `{
  "page": 2,
  "total_page": 2,
  "data": [
    {"kode_satusehat": "1000987654", "nama": "RS Hasan Santika"}
  ]
}`)
			default:
				http.Error(w, "unexpected upstream page", http.StatusBadRequest)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(config.SATUSEHATConfig{
		BaseURL:      server.URL + "/masterdata",
		AuthBaseURL:  server.URL + "/oauth2/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got, err := client.ListHospitals(context.Background(), "3273", "  Hasan  ", 2, 1)
	if err != nil {
		t.Fatalf("ListHospitals() error = %v", err)
	}
	if len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
		t.Fatalf("expected upstream pages 1 and 2, got %v", pages)
	}
	if got.Total != 2 || got.Page != 2 || got.PerPage != 1 || got.TotalPages != 2 || len(got.Data) != 1 {
		t.Fatalf("expected combined matches before public pagination, got %+v", got)
	}
	if got.Data[0].Name != "RS Hasan Santika" {
		t.Fatalf("expected public page 2 match, got %+v", got.Data[0])
	}
}

func TestClientMarksOAuth404AsNonResourceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/v1/accesstoken" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		http.Error(w, "private token body", http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(config.SATUSEHATConfig{
		BaseURL:      server.URL + "/masterdata",
		AuthBaseURL:  server.URL + "/oauth2/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var output struct{}
	err = client.GetJSON(context.Background(), "v1/resource", nil, &output)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("expected UpstreamError, got %T: %v", err, err)
	}
	if upstreamErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, upstreamErr.StatusCode)
	}
	if upstreamErr.IsResourceError {
		t.Fatal("expected OAuth error not to be marked as resource error")
	}
}

func TestClientMarksData404AsResourceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/v1/accesstoken":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"access_token":"access-token","expires_in":"3600"}`)
		case "/masterdata/v1/resource":
			http.Error(w, "private data body", http.StatusNotFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(config.SATUSEHATConfig{
		BaseURL:      server.URL + "/masterdata",
		AuthBaseURL:  server.URL + "/oauth2/v1",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var output struct{}
	err = client.GetJSON(context.Background(), "v1/resource", nil, &output)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("expected UpstreamError, got %T: %v", err, err)
	}
	if upstreamErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, upstreamErr.StatusCode)
	}
	if !upstreamErr.IsResourceError {
		t.Fatal("expected data error to be marked as resource error")
	}
}
func TestNewClientRejectsMissingSATUSEHATCredentials(t *testing.T) {
	_, err := NewClient(config.SATUSEHATConfig{
		BaseURL:     "https://api.example.test/masterdata",
		AuthBaseURL: "https://api.example.test/oauth2/v1",
	})
	if err == nil {
		t.Fatal("expected missing SATUSEHAT credentials to fail")
	}
}
