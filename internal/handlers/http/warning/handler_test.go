package warninghandler

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

	"github.com/gin-gonic/gin"
)

type serviceFake struct {
	warnings []dto.Warning
	err      error
	calls    int
	province string
}

func (f *serviceFake) ListWarnings(_ context.Context, province string) ([]dto.Warning, error) {
	f.calls++
	f.province = province
	return f.warnings, f.err
}

func TestHandlerRejectsMissingOrWhitespaceProvinceBeforeService(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)

	for _, query := range []string{"", "provinsi=", "provinsi=%20%20"} {
		router := gin.New()
		router.GET("/api/warnings", handler.GetWarnings)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/warnings?"+query, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("query %q: expected 400, got %d: %s", query, recorder.Code, recorder.Body.String())
		}
		if strings.Contains(recorder.Body.String(), "invalid province:") {
			t.Fatalf("response exposed internal validation detail: %s", recorder.Body.String())
		}
	}
	if service.calls != 0 {
		t.Fatalf("invalid province reached service %d times", service.calls)
	}
}

func TestHandlerWritesWarningSuccessEnvelopeAndPreservesOfficialText(t *testing.T) {
	service := &serviceFake{warnings: []dto.Warning{{
		ID:          "warning-1",
		AlertID:     "alert-1",
		Event:       "Heavy Rain",
		Urgency:     "Immediate",
		Severity:    "Severe",
		Certainty:   "Observed",
		Area:        "Bogor Selatan",
		Province:    "Jawa Barat",
		Effective:   "2026-08-20T10:00:00+07:00",
		Expires:     "20 Aug 2026 16:00 WIB",
		Headline:    "Official headline",
		Description: "Official description",
		Instruction: "Official instruction",
		Source:      "BMKG",
		IsActive:    true,
	}}}
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/api/warnings", handler.GetWarnings)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/warnings?provinsi=Jawa+Barat", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, fragment := range []string{`"status":true`, `"data":[`, `"alert_id":"alert-1"`, `"headline":"Official headline"`, `"description":"Official description"`, `"instruction":"Official instruction"`, `"is_active":true`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("response missing %s: %s", fragment, body)
		}
	}
	if service.calls != 1 || service.province != "Jawa Barat" {
		t.Fatalf("unexpected service call: calls=%d province=%q", service.calls, service.province)
	}
}

func TestHandlerReturnsServiceUnavailableWhenDependencyMissing(t *testing.T) {
	handler := NewHandler(nil)
	router := gin.New()
	router.GET("/api/warnings", handler.GetWarnings)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/warnings?provinsi=Jawa+Barat", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "not configured") || strings.Contains(body, "warning upstream client") {
		t.Fatalf("response exposed internal dependency error: %s", body)
	}
}

func TestHandlerMapsTypedUpstreamNotFoundToSafeNotFound(t *testing.T) {
	const (
		apiKey       = "SECRET_API_KEY"
		upstreamBody = "private upstream body"
	)
	handler := NewHandler(&serviceFake{err: fmt.Errorf("upstream response %s with key %s: %w", upstreamBody, apiKey, &apiindonesia.UpstreamError{
		StatusCode: http.StatusNotFound,
		Code:       "WARNING_NOT_FOUND",
	})})
	router := gin.New()
	router.GET("/api/warnings", handler.GetWarnings)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/warnings?provinsi=Jawa+Barat", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Warning not found") {
		t.Fatalf("expected safe not-found response: %s", body)
	}
	for _, secret := range []string{apiKey, upstreamBody, "upstream response", "WARNING_NOT_FOUND"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response exposed upstream detail %q: %s", secret, body)
		}
	}
}

func TestHandlerMapsOtherUpstreamErrorsToSafeBadGateway(t *testing.T) {
	const (
		apiKey       = "SECRET_API_KEY"
		upstreamBody = "private upstream body"
	)
	handler := NewHandler(&serviceFake{err: fmt.Errorf("upstream response %s with key %s: %w", upstreamBody, apiKey, &apiindonesia.UpstreamError{
		StatusCode: http.StatusTooManyRequests,
		Code:       "QUOTA_EXCEEDED",
	})})
	router := gin.New()
	router.GET("/api/warnings", handler.GetWarnings)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/warnings?provinsi=Jawa+Barat", nil))

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, secret := range []string{apiKey, upstreamBody, "upstream response", "QUOTA_EXCEEDED"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response exposed upstream detail %q: %s", secret, body)
		}
	}
}

func TestHandlerMapsUntypedFailureToSafeBadGateway(t *testing.T) {
	const secret = "upstream body and API key must stay private"
	handler := NewHandler(&serviceFake{err: errors.New(secret)})
	router := gin.New()
	router.GET("/api/warnings", handler.GetWarnings)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/warnings?provinsi=Jawa+Barat", nil))

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), secret) {
		t.Fatalf("response exposed upstream error: %s", recorder.Body.String())
	}
}

func TestRegisterAddsWarningRouteAndHandlesMissingClient(t *testing.T) {
	router := gin.New()
	Register(router, nil, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/warnings?provinsi=Jawa+Barat", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected registered route to return 503 without client, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

var _ Service = (*serviceFake)(nil)
