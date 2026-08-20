package weatherhandler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kota-siaga/internal/dto"

	"github.com/gin-gonic/gin"
)

type serviceFake struct {
	forecasts []dto.WeatherForecast
	err       error
	calls     int
}

func (f *serviceFake) GetWeather(_ context.Context, _ string) ([]dto.WeatherForecast, error) {
	f.calls++
	return f.forecasts, f.err
}

func TestHandlerRejectsInvalidADM4BeforeService(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/api/weather", handler.GetWeather)

	for _, query := range []string{"", "adm4=abc", "adm4=123456789012345678901"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/weather?"+query, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("query %q: expected 400, got %d: %s", query, recorder.Code, recorder.Body.String())
		}
		if strings.Contains(recorder.Body.String(), "invalid adm4:") {
			t.Fatalf("response exposed internal validation detail: %s", recorder.Body.String())
		}
	}
	if service.calls != 0 {
		t.Fatalf("invalid adm4 reached service %d times", service.calls)
	}
}

func TestHandlerWritesThreeDaySuccessEnvelope(t *testing.T) {
	service := &serviceFake{forecasts: []dto.WeatherForecast{
		{ID: "1", Adm4: "32.73.01.1001", Province: "Jawa Barat", WeatherDescription: "Cerah", TemperatureC: 24.5},
		{ID: "2", Adm4: "32.73.01.1001", Province: "Jawa Barat", WeatherDescription: "Hujan", TemperatureC: 23.25},
		{ID: "3", Adm4: "32.73.01.1001", Province: "Jawa Barat", WeatherDescription: "Berawan", TemperatureC: 25},
	}}
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/api/weather", handler.GetWeather)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/weather?adm4=32.73.01.1001", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, fragment := range []string{`"status":true`, `"data":[`, `"province":"Jawa Barat"`, `"weather_description":"Cerah"`, `"temperature_c":24.5`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("response missing %s: %s", fragment, body)
		}
	}
	if strings.Contains(body, `"provinsi"`) || strings.Contains(body, `"weather_desc"`) {
		t.Fatalf("response exposed upstream field names: %s", body)
	}
	if service.calls != 1 {
		t.Fatalf("expected one service call, got %d", service.calls)
	}
}

func TestHandlerMapsUpstreamFailureToSafeBadGateway(t *testing.T) {
	const (
		apiKey       = "SECRET_API_KEY"
		upstreamBody = "private upstream body"
	)
	handler := NewHandler(&serviceFake{err: fmt.Errorf("upstream response %s with key %s", upstreamBody, apiKey)})
	router := gin.New()
	router.GET("/api/weather", handler.GetWeather)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/weather?adm4=32.73.01.1001", nil))

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, secret := range []string{apiKey, upstreamBody, "upstream response"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response exposed upstream detail %q: %s", secret, body)
		}
	}
}

func TestHandlerReturnsServiceUnavailableWhenDependencyMissing(t *testing.T) {
	handler := NewHandler(nil)
	router := gin.New()
	router.GET("/api/weather", handler.GetWeather)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/weather?adm4=32.73.01.1001", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "not configured") || strings.Contains(body, "weather upstream client") {
		t.Fatalf("response exposed internal dependency error: %s", body)
	}
}

func TestHandlerMapsCanceledRequestToSafeBadGateway(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	handler := NewHandler(&serviceFake{err: context.Canceled})
	router := gin.New()
	router.GET("/api/weather", handler.GetWeather)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/weather?adm4=32.73.01.1001", nil).WithContext(ctx)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), context.Canceled.Error()) {
		t.Fatalf("response exposed cancellation detail: %s", recorder.Body.String())
	}
}

func TestRegisterAddsWeatherRouteAndHandlesMissingClient(t *testing.T) {
	router := gin.New()
	Register(router, nil, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/weather?adm4=32.73.01.1001", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected registered route to return 503 without client, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

var _ Service = (*serviceFake)(nil)
