package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *Client) ensureToken(ctx context.Context) error {
	if c.accessToken != "" && time.Now().Before(c.tokenExpires.Add(-30*time.Second)) {
		return nil
	}

	return c.authenticate(ctx)
}

func (c *Client) authenticate(ctx context.Context) error {
	if c.baseURL == "" {
		return errors.New("keycloak base url is required")
	}

	if c.realm == "" {
		return errors.New("keycloak realm is required")
	}

	if c.adminClient == "" {
		return errors.New("keycloak admin client id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.adminClient)

	if c.adminSecret != "" {
		form.Set("client_secret", c.adminSecret)
	}

	tokenURL := c.baseURL + "/realms/" + c.realm + "/protocol/openid-connect/token"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}
	defer resp.Body.Close()

	var token tokenResponse

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return err
	}

	if token.AccessToken == "" {
		return errors.New("keycloak returned empty access token")
	}

	c.accessToken = token.AccessToken
	c.tokenExpires = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return nil
}
