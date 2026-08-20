package hospitalhandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/integrations/apiindonesia"

	"github.com/gin-gonic/gin"
)

type serviceFake struct {
	page        dto.Page[dto.Hospital]
	err         error
	calls       int
	kabupatenID string
	pageNumber  int
	perPage     int
}

func (f *serviceFake) ListHospitals(_ context.Context, kabupatenID string, page, perPage int) (dto.Page[dto.Hospital], error) {
	f.calls++
	f.kabupatenID = kabupatenID
	f.pageNumber = page
	f.perPage = perPage
	return f.page, f.err
}

func TestHandlerRejectsInvalidHospitalQueryBeforeService(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/api/hospitals", handler.GetHospitals)

	for _, query := range []string{
		"",
		"kabupaten_id=",
		"kabupaten_id=32x&page=1&per_page=20",
		"kabupaten_id=3273&page=0&per_page=20",
		"kabupaten_id=3273&page=1&per_page=201",
		"kabupaten_id=3273&page=nope&per_page=20",
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/hospitals?"+query, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("query %q: expected 400, got %d: %s", query, recorder.Code, recorder.Body.String())
		}
		if strings.Contains(recorder.Body.String(), "invalid kabupaten") || strings.Contains(recorder.Body.String(), "must be") {
			t.Fatalf("response exposed internal validation detail: %s", recorder.Body.String())
		}
	}
	if service.calls != 0 {
		t.Fatalf("invalid hospital query reached service %d times", service.calls)
	}
}

func TestHandlerWritesPaginatedHospitalResponseWithApplicationFields(t *testing.T) {
	service := &serviceFake{page: dto.Page[dto.Hospital]{
		Data: []dto.Hospital{{
			ID: "hospital-1", Name: "Official Hospital", Type: "RSU", Class: "B", Ownership: "Public",
			Address: "Official Address", PostalCode: "16151", Phone: "021-555000", BedsTotal: 120, ICUBeds: 12,
			ProvinceName: "Jawa Barat", RegencyName: "Bogor", IsActive: true,
		}},
		Total: 21, Page: 2, PerPage: 20, TotalPages: 2,
	}}
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/api/hospitals", handler.GetHospitals)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/hospitals?kabupaten_id=003273&page=2&per_page=20", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, fragment := range []string{
		`"status":true`, `"total_data":21`, `"current_page":2`, `"limit":20`,
		`"id":"hospital-1"`, `"type":"RSU"`, `"class":"B"`, `"postal_code":"16151"`,
		`"beds_total":120`, `"icu_beds":12`, `"province_name":"Jawa Barat"`, `"regency_name":"Bogor"`, `"is_active":true`,
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("response missing %s: %s", fragment, body)
		}
	}
	if strings.Contains(body, `"jenis"`) || strings.Contains(body, `"kelas"`) || strings.Contains(body, `"province_id"`) {
		t.Fatalf("response exposed source field names: %s", body)
	}
	if service.calls != 1 || service.kabupatenID != "003273" || service.pageNumber != 2 || service.perPage != 20 {
		t.Fatalf("unexpected service call: calls=%d id=%q page=%d per_page=%d", service.calls, service.kabupatenID, service.pageNumber, service.perPage)
	}
}

func TestHospitalDTOUsesEnglishSnakeCaseJSON(t *testing.T) {
	encoded, err := json.Marshal(dto.Hospital{
		ID: "1", Name: "Name", Type: "Type", Class: "Class", Ownership: "Ownership", Address: "Address",
		PostalCode: "12345", Phone: "Phone", BedsTotal: 10, ICUBeds: 2, ProvinceName: "Province", RegencyName: "Regency", IsActive: true,
	})
	if err != nil {
		t.Fatalf("marshal hospital: %v", err)
	}
	body := string(encoded)
	for _, field := range []string{"id", "name", "type", "class", "ownership", "address", "postal_code", "phone", "beds_total", "icu_beds", "province_name", "regency_name", "is_active"} {
		if !strings.Contains(body, `"`+field+`"`) {
			t.Fatalf("JSON missing application field %q: %s", field, body)
		}
	}
}

func TestHandlerReturnsServiceUnavailableWhenDependencyMissing(t *testing.T) {
	handler := NewHandler(nil)
	router := gin.New()
	router.GET("/api/hospitals", handler.GetHospitals)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/hospitals?kabupaten_id=3273&page=1&per_page=20", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Hospital service unavailable") || strings.Contains(body, "not configured") || strings.Contains(body, "hospital upstream client") {
		t.Fatalf("unexpected dependency response: %s", body)
	}
}

func TestHandlerMapsTypedUpstreamNotFoundToSafeNotFound(t *testing.T) {
	const (
		apiKey       = "SECRET_API_KEY"
		upstreamBody = "private upstream body"
	)
	handler := NewHandler(&serviceFake{err: fmt.Errorf("upstream response %s with key %s: %w", upstreamBody, apiKey, &apiindonesia.UpstreamError{
		StatusCode: http.StatusNotFound,
		Code:       "HOSPITAL_NOT_FOUND",
	})})
	router := gin.New()
	router.GET("/api/hospitals", handler.GetHospitals)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/hospitals?kabupaten_id=3273&page=1&per_page=20", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Hospital not found") {
		t.Fatalf("expected safe not-found response: %s", body)
	}
	for _, secret := range []string{apiKey, upstreamBody, "upstream response", "HOSPITAL_NOT_FOUND"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response exposed upstream detail %q: %s", secret, body)
		}
	}
}

func TestHandlerMapsOtherUpstreamAndMalformedFailuresToSafeBadGateway(t *testing.T) {
	for _, wantErr := range []error{
		fmt.Errorf("upstream response private body with key SECRET_API_KEY: %w", &apiindonesia.UpstreamError{StatusCode: http.StatusBadGateway, Code: "UPSTREAM_ERROR"}),
		errors.New("malformed upstream payload details"),
	} {
		handler := NewHandler(&serviceFake{err: wantErr})
		router := gin.New()
		router.GET("/api/hospitals", handler.GetHospitals)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/hospitals?kabupaten_id=3273&page=1&per_page=20", nil))

		if recorder.Code != http.StatusBadGateway {
			t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
		}
		body := recorder.Body.String()
		for _, secret := range []string{"private body", "SECRET_API_KEY", "upstream response", "UPSTREAM_ERROR", "malformed upstream payload details"} {
			if strings.Contains(body, secret) {
				t.Fatalf("response exposed upstream detail %q: %s", secret, body)
			}
		}
	}
}

func TestRegisterAddsHospitalRouteAndHandlesMissingClient(t *testing.T) {
	router := gin.New()
	Register(router, nil, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/hospitals?kabupaten_id=3273&page=1&per_page=20", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected registered route to return 503 without client, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

var _ Service = (*serviceFake)(nil)
