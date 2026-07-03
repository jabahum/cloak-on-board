package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type CreateClientRequest struct {
	ClientID                  string            `json:"clientId"`
	Name                      string            `json:"name,omitempty"`
	Description               string            `json:"description,omitempty"`
	Protocol                  string            `json:"protocol"`
	PublicClient              bool              `json:"publicClient"`
	BearerOnly                bool              `json:"bearerOnly"`
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`
	RedirectURIs              []string          `json:"redirectUris,omitempty"`
	WebOrigins                []string          `json:"webOrigins,omitempty"`
	Enabled                   bool              `json:"enabled"`
	Attributes                map[string]string `json:"attributes,omitempty"`
}

type KeycloakClient struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Name     string `json:"name"`
}

func (c *Client) CreateClient(ctx context.Context, req CreateClientRequest) error {
	resp, err := c.do(
		ctx,
		http.MethodPost,
		"/admin/realms/"+c.realm+"/clients",
		req,
	)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}

	resp.Body.Close()
	return nil
}

func (c *Client) GetClientByClientID(ctx context.Context, clientID string) (KeycloakClient, error) {
	path := "/admin/realms/" + c.realm + "/clients?clientId=" + url.QueryEscape(clientID)

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return KeycloakClient{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return KeycloakClient{}, decodeError(resp)
	}
	defer resp.Body.Close()

	var clients []KeycloakClient

	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return KeycloakClient{}, err
	}

	if len(clients) == 0 {
		return KeycloakClient{}, fmt.Errorf("keycloak client %q not found", clientID)
	}

	return clients[0], nil
}

func (c *Client) CreateOrGetClient(ctx context.Context, req CreateClientRequest) (KeycloakClient, error) {
	existing, err := c.GetClientByClientID(ctx, req.ClientID)
	if err == nil {
		return existing, nil
	}

	if err := c.CreateClient(ctx, req); err != nil {
		return KeycloakClient{}, err
	}

	return c.GetClientByClientID(ctx, req.ClientID)
}
