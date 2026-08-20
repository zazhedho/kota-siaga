package apiindonesia

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"kota-siaga/pkg/config"
)

const (
	maxAPIIndonesiaResponseBodySize = 1 << 20
	maxAPIIndonesiaErrorCodeLength  = 64
)

type APIIndonesiaConfig = config.APIIndonesiaConfig

type Client struct {
	BaseURL    *url.URL
	APIKey     string
	HTTPClient *http.Client
}

type UpstreamError struct {
	StatusCode int
	Code       string
}

func (e *UpstreamError) Error() string {
	code := e.Code
	if code == "" {
		code = "UNKNOWN"
	}
	return fmt.Sprintf("API Indonesia upstream error: status %d code %s", e.StatusCode, code)
}

func NewClient(clientConfig APIIndonesiaConfig) (*Client, error) {
	baseURLString := strings.TrimSpace(clientConfig.BaseURL)
	if err := config.ValidateAPIIndonesiaBaseURL(baseURLString); err != nil {
		return nil, errors.New("invalid API Indonesia base URL")
	}
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, errors.New("invalid API Indonesia base URL")
	}

	timeout := clientConfig.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		BaseURL:    baseURL,
		APIKey:     strings.TrimSpace(clientConfig.APIKey),
		HTTPClient: withoutRedirects(&http.Client{Timeout: timeout}),
	}, nil
}

func (c *Client) GetJSON(ctx context.Context, requestPath string, query url.Values, out any) error {
	if c == nil || c.BaseURL == nil {
		return errors.New("API Indonesia client is not configured")
	}

	target, err := c.buildURL(requestPath, query)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return fmt.Errorf("create API Indonesia request: %w", err)
	}
	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := withoutRedirects(c.HTTPClient).Do(req)
	if err != nil {
		return fmt.Errorf("API Indonesia request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIIndonesiaResponseBodySize+1))
	if err != nil {
		return fmt.Errorf("read API Indonesia response: %w", err)
	}
	if len(body) > maxAPIIndonesiaResponseBodySize {
		return errors.New("API Indonesia response body too large")
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return newUpstreamError(resp.StatusCode, body, c.APIKey)
	}

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode API Indonesia response: %w", err)
	}
	data := bytes.TrimSpace(envelope.Data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return errors.New("API Indonesia response missing data")
	}
	if out == nil {
		return errors.New("API Indonesia output is nil")
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode API Indonesia response: %w", err)
	}
	return nil
}

func (c *Client) buildURL(requestPath string, query url.Values) (*url.URL, error) {
	pathURL, err := url.Parse(requestPath)
	if err != nil || pathURL.IsAbs() || pathURL.Host != "" || pathURL.RawQuery != "" || pathURL.Fragment != "" {
		return nil, errors.New("invalid API Indonesia request path")
	}

	target := c.BaseURL.JoinPath(strings.TrimPrefix(pathURL.Path, "/"))
	target.RawQuery = query.Encode()
	target.Fragment = ""
	return target, nil
}

func withoutRedirects(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	copyClient := *client
	copyClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &copyClient
}

func newUpstreamError(statusCode int, body []byte, apiKey string) *UpstreamError {
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &envelope)
	code := envelope.Error.Code
	if !validUpstreamErrorCode(code) || containsAPIKey(code, apiKey) {
		code = safeUpstreamErrorCode(statusCode, apiKey)
	}
	return &UpstreamError{StatusCode: statusCode, Code: code}
}

func safeUpstreamErrorCode(statusCode int, apiKey string) string {
	for _, code := range []string{
		"HTTP_" + strconv.Itoa(statusCode),
		"UPSTREAM_ERROR",
		"UNKNOWN",
	} {
		if !containsAPIKey(code, apiKey) {
			return code
		}
	}
	if !containsAPIKey("X", apiKey) {
		return "X"
	}
	return "Y"
}

func containsAPIKey(value, apiKey string) bool {
	return apiKey != "" && strings.Contains(strings.ToLower(value), strings.ToLower(apiKey))
}

func validUpstreamErrorCode(code string) bool {
	if code == "" || len(code) > maxAPIIndonesiaErrorCodeLength {
		return false
	}
	for i := 0; i < len(code); i++ {
		char := code[i]
		if (char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
}
