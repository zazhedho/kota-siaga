package earthquakehandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kota-siaga/internal/dto"

	"github.com/gin-gonic/gin"
)

type serviceFake struct {
	earthquakes []dto.Earthquake
	err         error
	calls       int
}

func (f *serviceFake) ListLatest(_ context.Context) ([]dto.Earthquake, error) {
	f.calls++
	return f.earthquakes, f.err
}

func TestHandlerWritesSuccessEnvelopeAndIgnoresQueryParameters(t *testing.T) {
	service := &serviceFake{earthquakes: []dto.Earthquake{{
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
	}}}
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/api/earthquakes/latest", handler.GetLatest)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/earthquakes/latest?unexpected=query", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, fragment := range []string{`"status":true`, `"data":[`, `"date_time":"2026-08-20T03:15:00Z"`, `"depth_km":10.5`, `"latitude":-6.1234`, `"longitude":106.9876`, `"is_felt":true`, `"felt_areas":null`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("response missing %s: %s", fragment, body)
		}
	}
	if strings.Contains(body, `"datetime"`) || strings.Contains(body, `"lat"`) || strings.Contains(body, `"lng"`) {
		t.Fatalf("response exposed upstream field names: %s", body)
	}
	if service.calls != 1 {
		t.Fatalf("expected one service call, got %d", service.calls)
	}
}

func TestHandlerReturnsServiceUnavailableWhenDependencyMissing(t *testing.T) {
	handler := NewHandler(nil)
	router := gin.New()
	router.GET("/api/earthquakes/latest", handler.GetLatest)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/earthquakes/latest", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "not configured") || strings.Contains(body, "earthquake upstream client") {
		t.Fatalf("response exposed internal dependency error: %s", body)
	}
}

func TestHandlerMapsMalformedPayloadFailureToSafeBadGateway(t *testing.T) {
	const upstreamErrorDetails = "malformed upstream payload details"
	handler := NewHandler(&serviceFake{err: errors.New(upstreamErrorDetails)})
	router := gin.New()
	router.GET("/api/earthquakes/latest", handler.GetLatest)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/earthquakes/latest", nil))

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), upstreamErrorDetails) {
		t.Fatalf("response exposed malformed payload detail: %s", recorder.Body.String())
	}
}

func TestRegisterAddsEarthquakeRouteAndHandlesMissingClient(t *testing.T) {
	router := gin.New()
	Register(router, nil, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/earthquakes/latest", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected registered route to return 503 without client, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

var _ Service = (*serviceFake)(nil)
