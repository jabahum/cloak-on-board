package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKSProvider struct {
	baseURL    string
	realm      string
	httpClient *http.Client

	mu        sync.RWMutex
	cached    JWKS
	expiresAt time.Time
}

func NewJWKSProvider(baseURL, realm string) *JWKSProvider {
	return &JWKSProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		realm:   realm,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *JWKSProvider) Get(ctx context.Context) (JWKS, error) {
	p.mu.RLock()
	if time.Now().Before(p.expiresAt) && len(p.cached.Keys) > 0 {
		defer p.mu.RUnlock()
		return p.cached, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	url := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/certs",
		p.baseURL,
		p.realm,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return JWKS{}, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return JWKS{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return JWKS{}, fmt.Errorf("fetch jwks failed with status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return JWKS{}, err
	}

	p.cached = jwks
	p.expiresAt = time.Now().Add(10 * time.Minute)

	return jwks, nil
}
