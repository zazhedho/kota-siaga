package satusehat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"kota-siaga/pkg/config"
)

const (
	maxSATUSEHATResponseBodySize = 1 << 20
	accessTokenSafetyWindow      = 30 * time.Second
	defaultAccessTokenTTL        = 5 * time.Minute
)

type Client struct {
	BaseURL      *url.URL
	AuthBaseURL  *url.URL
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client

	tokenMu        sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

type UpstreamError struct {
	StatusCode      int
	Code            string
	IsResourceError bool
}

func (e *UpstreamError) Error() string {
	code := e.Code
	if code == "" {
		code = "UNKNOWN"
	}
	return fmt.Sprintf("SATUSEHAT upstream error: status %d code %s", e.StatusCode, code)
}

func NewClient(clientConfig config.SATUSEHATConfig) (*Client, error) {
	if err := config.ValidateSATUSEHATConfig(clientConfig); err != nil {
		return nil, errors.New("invalid SATUSEHAT configuration")
	}

	baseURL, err := url.Parse(strings.TrimSpace(clientConfig.BaseURL))
	if err != nil {
		return nil, errors.New("invalid SATUSEHAT base URL")
	}
	authBaseURL, err := url.Parse(strings.TrimSpace(clientConfig.AuthBaseURL))
	if err != nil {
		return nil, errors.New("invalid SATUSEHAT auth base URL")
	}

	timeout := clientConfig.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		BaseURL:      baseURL,
		AuthBaseURL:  authBaseURL,
		ClientID:     strings.TrimSpace(clientConfig.ClientID),
		ClientSecret: strings.TrimSpace(clientConfig.ClientSecret),
		HTTPClient:   withoutRedirects(&http.Client{Timeout: timeout}),
	}, nil
}

func (c *Client) GetJSON(ctx context.Context, requestPath string, query url.Values, out any) error {
	if c == nil || c.BaseURL == nil || c.AuthBaseURL == nil {
		return errors.New("SATUSEHAT client is not configured")
	}
	if out == nil {
		return errors.New("SATUSEHAT output is nil")
	}

	token, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}
	target, err := c.buildURL(requestPath, query)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return fmt.Errorf("create SATUSEHAT request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("SATUSEHAT request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := readResponseBody(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return newUpstreamError(resp.StatusCode, true)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode SATUSEHAT response: %w", err)
	}
	return nil
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiresAt) {
		return c.accessToken, nil
	}

	target := c.AuthBaseURL.JoinPath("accesstoken")
	target.RawQuery = url.Values{"grant_type": {"client_credentials"}}.Encode()
	form := url.Values{
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create SATUSEHAT token request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("SATUSEHAT token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := readResponseBody(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", newUpstreamError(resp.StatusCode, false)
	}

	var tokenResponse struct {
		AccessToken string          `json:"access_token"`
		ExpiresIn   json.RawMessage `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("decode SATUSEHAT token response: %w", err)
	}
	if strings.TrimSpace(tokenResponse.AccessToken) == "" {
		return "", errors.New("SATUSEHAT token response missing access token")
	}

	ttl := parseAccessTokenTTL(tokenResponse.ExpiresIn)
	if ttl > accessTokenSafetyWindow {
		ttl -= accessTokenSafetyWindow
	}
	c.accessToken = strings.TrimSpace(tokenResponse.AccessToken)
	c.tokenExpiresAt = time.Now().Add(ttl)
	return c.accessToken, nil
}

func parseAccessTokenTTL(raw json.RawMessage) time.Duration {
	value := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return defaultAccessTokenTTL
	}
	return time.Duration(seconds) * time.Second
}

func (c *Client) buildURL(requestPath string, query url.Values) (*url.URL, error) {
	pathURL, err := url.Parse(requestPath)
	if err != nil || pathURL.IsAbs() || pathURL.Host != "" || pathURL.RawQuery != "" || pathURL.Fragment != "" {
		return nil, errors.New("invalid SATUSEHAT request path")
	}
	target := c.BaseURL.JoinPath(strings.TrimPrefix(pathURL.Path, "/"))
	target.RawQuery = query.Encode()
	target.Fragment = ""
	return target, nil
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return withoutRedirects(c.HTTPClient)
	}
	return withoutRedirects(&http.Client{Timeout: 5 * time.Second})
}

func readResponseBody(reader io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxSATUSEHATResponseBodySize+1))
	if err != nil {
		return nil, errors.New("read SATUSEHAT response failed")
	}
	if len(body) > maxSATUSEHATResponseBodySize {
		return nil, errors.New("SATUSEHAT response body too large")
	}
	return body, nil
}

func newUpstreamError(statusCode int, isResourceError bool) *UpstreamError {
	return &UpstreamError{
		StatusCode:      statusCode,
		Code:            "HTTP_" + strconv.Itoa(statusCode),
		IsResourceError: isResourceError,
	}
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
