package apiindonesia

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type clientTestData struct {
	ID int `json:"id"`
}

type clientTestEnvelope struct {
	Data clientTestData `json:"data"`
	Meta struct {
		Page int `json:"page"`
	} `json:"meta"`
}

func TestClientGetJSONBuildsRequestAndDecodesDataEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/root/v1/items" {
			t.Errorf("expected joined path /root/v1/items, got %q", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "test-api-key" {
			t.Errorf("expected API key header, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("expected JSON Accept header, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"id":7},"meta":{"page":2}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{
		BaseURL: server.URL + "/root",
		APIKey:  "test-api-key",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var got clientTestEnvelope
	if err := client.GetJSON(context.Background(), "v1/items", nil, &got); err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
	if got.Data.ID != 7 || got.Meta.Page != 2 {
		t.Fatalf("unexpected decoded envelope: %#v", got)
	}
}

func TestClientGetJSONEncodesQueryValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RawQuery; got != "filter=hello+world&tag=a%26b" {
			t.Errorf("expected encoded query, got %q", got)
		}
		_, _ = io.WriteString(w, `{"data":{}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var got clientTestEnvelope
	query := url.Values{"filter": {"hello world"}, "tag": {"a&b"}}
	if err := client.GetJSON(context.Background(), "/items", query, &got); err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
}

func TestClientGetJSONReturnsSafeTypedUpstreamError(t *testing.T) {
	const apiKey = "test-api-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":{"code":"QUOTA_EXCEEDED","message":"test-api-key"}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{
		BaseURL: server.URL,
		APIKey:  apiKey,
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var got clientTestEnvelope
	err = client.GetJSON(context.Background(), "items", nil, &got)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("expected UpstreamError, got %T: %v", err, err)
	}
	if upstreamErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, upstreamErr.StatusCode)
	}
	if upstreamErr.Code != "QUOTA_EXCEEDED" {
		t.Fatalf("expected QUOTA_EXCEEDED, got %q", upstreamErr.Code)
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("upstream error exposed API key: %v", err)
	}
}

func TestClientGetJSONReturnsMalformedSuccessError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var got clientTestEnvelope
	if err := client.GetJSON(context.Background(), "items", nil, &got); err == nil {
		t.Fatal("expected malformed success payload error")
	}
}

func TestClientGetJSONHonorsContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	var got clientTestEnvelope
	err = client.GetJSON(ctx, "slow", nil, &got)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}

func TestNewClientRejectsUnsafeBaseURL(t *testing.T) {
	for _, baseURL := range []string{"ftp://api.example.test", "/relative"} {
		if _, err := NewClient(APIIndonesiaConfig{BaseURL: baseURL, Timeout: time.Second}); err == nil {
			t.Fatalf("expected invalid base URL %q to fail", baseURL)
		}
	}
}

func TestClientGetJSONDoesNotFollowRedirectsWithCustomHTTPClient(t *testing.T) {
	const apiKey = "SECRET_API_KEY"
	secondHit := make(chan struct{}, 1)
	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondHit <- struct{}{}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer second.Close()

	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, second.URL+"/redirected", http.StatusFound)
	}))
	defer first.Close()

	client, err := NewClient(APIIndonesiaConfig{
		BaseURL: first.URL,
		APIKey:  apiKey,
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	var got clientTestEnvelope
	err = client.GetJSON(context.Background(), "items", nil, &got)
	if err == nil {
		t.Fatal("expected redirect response error")
	}
	select {
	case <-secondHit:
		t.Fatal("redirect target was hit")
	default:
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("redirect error exposed API key: %v", err)
	}
}

func TestClientGetJSONRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("x", 1<<20+1))
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var got clientTestEnvelope
	err = client.GetJSON(context.Background(), "items", nil, &got)
	if err == nil || !strings.Contains(err.Error(), "response body too large") {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}

func TestClientGetJSONFallsBackWhenUpstreamErrorCodeContainsAPIKey(t *testing.T) {
	const apiKey = "SECRET_API_KEY"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"error":{"code":"UPSTREAM_SECRET_API_KEY","message":"do not expose"}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, APIKey: apiKey, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var got clientTestEnvelope
	err = client.GetJSON(context.Background(), "items", nil, &got)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("expected UpstreamError, got %T: %v", err, err)
	}
	if upstreamErr.Code != "HTTP_502" {
		t.Fatalf("expected fallback code HTTP_502, got %q", upstreamErr.Code)
	}
	if strings.Contains(err.Error(), apiKey) || strings.Contains(err.Error(), "do not expose") {
		t.Fatalf("upstream error exposed secret details: %v", err)
	}
}

func TestClientGetJSONRedactsMixedCaseAPIKeyFromFallbackCode(t *testing.T) {
	const apiKey = "hTtP_502"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"error":{"code":"UPSTREAM_HTTP_502","message":"do not expose"}}`)
	}))
	defer server.Close()

	client, err := NewClient(APIIndonesiaConfig{BaseURL: server.URL, APIKey: apiKey, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var got clientTestEnvelope
	err = client.GetJSON(context.Background(), "items", nil, &got)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("expected UpstreamError, got %T: %v", err, err)
	}
	if upstreamErr.Code != "UPSTREAM_ERROR" {
		t.Fatalf("expected safe generic fallback code, got %q", upstreamErr.Code)
	}
	if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(apiKey)) ||
		strings.Contains(err.Error(), "do not expose") {
		t.Fatalf("upstream error exposed secret details: %v", err)
	}
}

func TestWithoutRedirectsOnlyDefaultsTimeoutForNilClient(t *testing.T) {
	custom := withoutRedirects(&http.Client{Timeout: 0})
	if custom.Timeout != 0 {
		t.Fatalf("expected custom zero timeout to remain zero, got %v", custom.Timeout)
	}
	if got := withoutRedirects(nil).Timeout; got != 5*time.Second {
		t.Fatalf("expected nil client default timeout 5s, got %v", got)
	}
}
