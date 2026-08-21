package bmkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kota-siaga/pkg/config"
)

const maxBMKGResponseBodySize = 1 << 20

type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(clientConfig config.BMKGConfig) (*Client, error) {
	baseURLString := strings.TrimSpace(clientConfig.BaseURL)
	if err := config.ValidateBMKGBaseURL(baseURLString); err != nil {
		return nil, errors.New("invalid BMKG base URL")
	}
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		return nil, errors.New("invalid BMKG base URL")
	}

	timeout := clientConfig.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (c *Client) getJSON(ctx context.Context, requestPath string, out any) error {
	if c == nil || c.BaseURL == nil {
		return errors.New("BMKG client is not configured")
	}

	target := c.BaseURL.JoinPath(requestPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return fmt.Errorf("create BMKG request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("BMKG request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBMKGResponseBodySize+1))
	if err != nil {
		return fmt.Errorf("read BMKG response: %w", err)
	}
	if len(body) > maxBMKGResponseBodySize {
		return errors.New("BMKG response body too large")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("BMKG request failed with status %d", resp.StatusCode)
	}
	if out == nil {
		return errors.New("BMKG output is nil")
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode BMKG response: %w", err)
	}
	return nil
}
