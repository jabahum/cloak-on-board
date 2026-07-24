package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	ID                        string            `json:"id"`
	ClientID                  string            `json:"clientId"`
	Name                      string            `json:"name"`
	Description               string            `json:"description"`
	Protocol                  string            `json:"protocol"`
	PublicClient              bool              `json:"publicClient"`
	BearerOnly                bool              `json:"bearerOnly"`
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`
	RedirectURIs              []string          `json:"redirectUris"`
	WebOrigins                []string          `json:"webOrigins"`
	Enabled                   bool              `json:"enabled"`
	Attributes                map[string]string `json:"attributes,omitempty"`
	Imported                  bool              `json:"imported,omitempty"`
	Representation            map[string]any    `json:"-"`
}

var ErrNotFound = errors.New("keycloak resource not found")

func (c *Client) CreateClient(ctx context.Context, req CreateClientRequest) error {
	resp, err := c.do(
		ctx,
		http.MethodPost,
		"/admin/realms/"+url.PathEscape(c.realm)+"/clients",
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
		return KeycloakClient{}, fmt.Errorf("%w: keycloak client %q", ErrNotFound, clientID)
	}

	return c.GetClient(ctx, clients[0].ID)
}

func (c *Client) ListClients(ctx context.Context, search string) ([]KeycloakClient, error) {
	values := url.Values{}
	if search != "" {
		values.Set("search", search)
	}
	path := "/admin/realms/" + url.PathEscape(c.realm) + "/clients"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeError(resp)
	}
	defer resp.Body.Close()

	clients := []KeycloakClient{}
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, err
	}
	return clients, nil
}

func (c *Client) GetClient(ctx context.Context, clientUUID string) (KeycloakClient, error) {
	resp, err := c.do(ctx, http.MethodGet, c.clientPath(clientUUID), nil)
	if err != nil {
		return KeycloakClient{}, err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return KeycloakClient{}, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return KeycloakClient{}, decodeError(resp)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return KeycloakClient{}, err
	}
	var client KeycloakClient
	if err := json.Unmarshal(body, &client); err != nil {
		return KeycloakClient{}, err
	}
	if err := json.Unmarshal(body, &client.Representation); err != nil {
		return KeycloakClient{}, err
	}
	return client, nil
}

func (c *Client) UpdateClient(ctx context.Context, clientUUID string, client KeycloakClient) error {
	client.ID = clientUUID
	payload := any(client)
	if client.Representation != nil {
		client.Representation["id"] = clientUUID
		client.Representation["clientId"] = client.ClientID
		client.Representation["name"] = client.Name
		client.Representation["description"] = client.Description
		client.Representation["protocol"] = client.Protocol
		client.Representation["publicClient"] = client.PublicClient
		client.Representation["bearerOnly"] = client.BearerOnly
		client.Representation["standardFlowEnabled"] = client.StandardFlowEnabled
		client.Representation["directAccessGrantsEnabled"] = client.DirectAccessGrantsEnabled
		client.Representation["serviceAccountsEnabled"] = client.ServiceAccountsEnabled
		client.Representation["redirectUris"] = client.RedirectURIs
		client.Representation["webOrigins"] = client.WebOrigins
		client.Representation["enabled"] = client.Enabled
		client.Representation["attributes"] = client.Attributes
		payload = client.Representation
	}
	resp, err := c.do(ctx, http.MethodPut, c.clientPath(clientUUID), payload)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}
	resp.Body.Close()
	return nil
}

func (c *Client) DeleteClient(ctx context.Context, clientUUID string) error {
	resp, err := c.do(ctx, http.MethodDelete, c.clientPath(clientUUID), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}
	resp.Body.Close()
	return nil
}

func (c *Client) clientPath(clientUUID string) string {
	return "/admin/realms/" + url.PathEscape(c.realm) + "/clients/" + url.PathEscape(clientUUID)
}

func (c *Client) CreateOrGetClient(ctx context.Context, req CreateClientRequest) (KeycloakClient, error) {
	existing, err := c.GetClientByClientID(ctx, req.ClientID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return KeycloakClient{}, err
	}

	if err := c.CreateClient(ctx, req); err != nil {
		return KeycloakClient{}, err
	}

	return c.GetClientByClientID(ctx, req.ClientID)
}
