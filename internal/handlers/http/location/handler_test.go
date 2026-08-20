package locationhandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kota-siaga/internal/dto"
	"kota-siaga/internal/integrations/apiindonesia"
	locationservice "kota-siaga/internal/services/location"

	"github.com/gin-gonic/gin"
)

type serviceFake struct {
	provincePage dto.Page[dto.Province]
	err          error
	calls        int
}

func (f *serviceFake) ListProvinces(_ context.Context, _, _ int) (dto.Page[dto.Province], error) {
	f.calls++
	return f.provincePage, f.err
}
func (f *serviceFake) ListCities(_ context.Context, _ string, _, _ int) (dto.Page[dto.City], error) {
	f.calls++
	return dto.Page[dto.City]{}, f.err
}
func (f *serviceFake) ListDistricts(_ context.Context, _ string, _, _ int) (dto.Page[dto.District], error) {
	f.calls++
	return dto.Page[dto.District]{}, f.err
}
func (f *serviceFake) ListVillages(_ context.Context, _ string, _, _ int) (dto.Page[dto.Village], error) {
	f.calls++
	return dto.Page[dto.Village]{}, f.err
}

func TestHandlerRejectsMissingOrInvalidParentID(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)

	for _, path := range []string{
		"/api/locations/city?page=1&per_page=20",
		"/api/locations/district?kabupaten_id=abc&page=1&per_page=20",
		"/api/locations/village?kecamatan_id=&page=1&per_page=20",
	} {
		router := gin.New()
		router.GET("/api/locations/city", handler.GetCity)
		router.GET("/api/locations/district", handler.GetDistrict)
		router.GET("/api/locations/village", handler.GetVillage)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d: %s", path, recorder.Code, recorder.Body.String())
		}
	}
	if service.calls != 0 {
		t.Fatalf("invalid parent reached service %d times", service.calls)
	}
}

func TestHandlerRejectsInvalidPagination(t *testing.T) {
	handler := NewHandler(&serviceFake{})
	for _, query := range []string{"", "page=1", "per_page=20", "page=0&per_page=20", "page=1&per_page=101", "page=nope&per_page=20"} {
		router := gin.New()
		router.GET("/api/locations/province", handler.GetProvince)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/locations/province?"+query, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d: %s", query, recorder.Code, recorder.Body.String())
		}
	}
}

func TestHandlerWritesNormalizedPaginatedResponse(t *testing.T) {
	service := &serviceFake{provincePage: dto.Page[dto.Province]{
		Data:  []dto.Province{{ID: "11", Code: "11", Name: "Aceh", AlternateName: "Aceh Official", Latitude: 4.6, Longitude: 96.7, IsActive: true}},
		Total: 21, Page: 2, PerPage: 20, TotalPages: 2,
	}}
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/api/locations/province", handler.GetProvince)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/locations/province?page=2&per_page=20", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, fragment := range []string{`"total_data":21`, `"current_page":2`, `"limit":20`, `"alternate_name":"Aceh Official"`, `"latitude":4.6`, `"longitude":96.7`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("response missing %s: %s", fragment, body)
		}
	}
	if strings.Contains(body, `"alt_name"`) || strings.Contains(body, `"lat"`) || strings.Contains(body, `"lng"`) {
		t.Fatalf("response exposed upstream field names: %s", body)
	}
}

func TestHandlerMapsTypedUpstreamNotFoundToSafeNotFound(t *testing.T) {
	const (
		apiKey       = "SECRET_API_KEY"
		upstreamBody = "private upstream body"
	)
	handler := NewHandler(&serviceFake{err: fmt.Errorf("upstream response %s with key %s: %w", upstreamBody, apiKey, &apiindonesia.UpstreamError{
		StatusCode: http.StatusNotFound,
		Code:       "LOCATION_NOT_FOUND",
	})})
	router := gin.New()
	router.GET("/api/locations/province", handler.GetProvince)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/locations/province?page=1&per_page=20", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Location not found") {
		t.Fatalf("expected safe English not-found response: %s", body)
	}
	for _, secret := range []string{apiKey, upstreamBody, "upstream response"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response exposed upstream detail %q: %s", secret, body)
		}
	}
}

func TestHandlerMapsMissingLocationClientToSafeServiceUnavailable(t *testing.T) {
	handler := NewHandler(locationservice.NewService(nil, nil))
	router := gin.New()
	router.GET("/api/locations/province", handler.GetProvince)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/locations/province?page=1&per_page=20", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "Location service unavailable") {
		t.Fatalf("expected safe English service-unavailable response: %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "location upstream client is not configured") {
		t.Fatalf("response exposed internal client error: %s", recorder.Body.String())
	}
}

func TestHandlerKeepsTypedUpstreamDetailsOutOfResponse(t *testing.T) {
	const (
		apiKey       = "SECRET_API_KEY"
		upstreamBody = "private upstream body"
	)
	handler := NewHandler(&serviceFake{err: fmt.Errorf("upstream response %s with key %s: %w", upstreamBody, apiKey, &apiindonesia.UpstreamError{
		StatusCode: http.StatusBadGateway,
		Code:       "UPSTREAM_ERROR",
	})})
	router := gin.New()
	router.GET("/api/locations/province", handler.GetProvince)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/locations/province?page=1&per_page=20", nil))

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, secret := range []string{apiKey, upstreamBody, "upstream response", "UPSTREAM_ERROR"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response exposed upstream detail %q: %s", secret, body)
		}
	}
}

func TestHandlerMapsUpstreamFailureToSafeBadGateway(t *testing.T) {
	handler := NewHandler(&serviceFake{err: errors.New("upstream body and api key must stay private")})
	router := gin.New()
	router.GET("/api/locations/province", handler.GetProvince)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/locations/province?page=1&per_page=20", nil))

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "api key") || strings.Contains(recorder.Body.String(), "upstream body") {
		t.Fatalf("response exposed upstream details: %s", recorder.Body.String())
	}
}

func TestRegisterAddsIndependentLocationRoutes(t *testing.T) {
	router := gin.New()
	Register(router, nil, nil)

	for _, path := range []string{
		"/api/locations/province?page=1&per_page=20",
		"/api/locations/city?province_id=32&page=1&per_page=20",
		"/api/locations/district?kabupaten_id=3273&page=1&per_page=20",
		"/api/locations/village?kecamatan_id=3273010&page=1&per_page=20",
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code == http.StatusNotFound {
			t.Fatalf("route was not registered for %s", path)
		}
	}
}
