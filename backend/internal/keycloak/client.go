package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL      string
	realm        string
	adminClient  string
	adminSecret  string
	httpClient   *http.Client
	accessToken  string
	tokenExpires time.Time
}

func NewClient(baseURL, realm, adminClient, adminSecret string) *Client {
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		realm:       realm,
		adminClient: adminClient,
		adminSecret: adminSecret,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	var reader io.Reader

	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		reader = bytes.NewReader(payload)
	}

	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func decodeError(resp *http.Response) error {
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if len(body) == 0 {
		return fmt.Errorf("keycloak request failed with status %d", resp.StatusCode)
	}

	return fmt.Errorf("keycloak request failed with status %d: %s", resp.StatusCode, string(body))
}
