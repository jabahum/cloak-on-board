package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type ClientScope struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ClientScopeAssignments struct {
	Default   []ClientScope `json:"default"`
	Optional  []ClientScope `json:"optional"`
	Available []ClientScope `json:"available"`
}

func (c *Client) ListClientScopes(ctx context.Context) ([]ClientScope, error) {
	return c.listScopes(ctx, "/admin/realms/"+url.PathEscape(c.realm)+"/client-scopes")
}

func (c *Client) ListDefaultClientScopes(ctx context.Context, clientUUID string) ([]ClientScope, error) {
	return c.listScopes(ctx, c.clientPath(clientUUID)+"/default-client-scopes")
}

func (c *Client) ListOptionalClientScopes(ctx context.Context, clientUUID string) ([]ClientScope, error) {
	return c.listScopes(ctx, c.clientPath(clientUUID)+"/optional-client-scopes")
}

func (c *Client) listScopes(ctx context.Context, path string) ([]ClientScope, error) {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeError(resp)
	}
	defer resp.Body.Close()

	scopes := []ClientScope{}
	if err := json.NewDecoder(resp.Body).Decode(&scopes); err != nil {
		return nil, err
	}
	return scopes, nil
}

func (c *Client) GetClientScopeAssignments(ctx context.Context, clientUUID string) (ClientScopeAssignments, error) {
	all, err := c.ListClientScopes(ctx)
	if err != nil {
		return ClientScopeAssignments{}, err
	}
	defaults, err := c.ListDefaultClientScopes(ctx, clientUUID)
	if err != nil {
		return ClientScopeAssignments{}, err
	}
	optional, err := c.ListOptionalClientScopes(ctx, clientUUID)
	if err != nil {
		return ClientScopeAssignments{}, err
	}

	assigned := map[string]bool{}
	for _, scope := range append(defaults, optional...) {
		assigned[scope.ID] = true
	}
	available := []ClientScope{}
	for _, scope := range all {
		if !assigned[scope.ID] {
			available = append(available, scope)
		}
	}
	return ClientScopeAssignments{Default: defaults, Optional: optional, Available: available}, nil
}

func (c *Client) AssignDefaultClientScope(ctx context.Context, clientUUID string, scopeID string) error {
	resp, err := c.do(
		ctx,
		http.MethodPut,
		c.clientPath(clientUUID)+"/default-client-scopes/"+url.PathEscape(scopeID),
		nil,
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

func (c *Client) AssignOptionalClientScope(ctx context.Context, clientUUID string, scopeID string) error {
	resp, err := c.do(
		ctx,
		http.MethodPut,
		c.clientPath(clientUUID)+"/optional-client-scopes/"+url.PathEscape(scopeID),
		nil,
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

func (c *Client) RemoveDefaultClientScope(ctx context.Context, clientUUID string, scopeID string) error {
	return c.removeScope(ctx, c.clientPath(clientUUID)+"/default-client-scopes/"+url.PathEscape(scopeID))
}

func (c *Client) RemoveOptionalClientScope(ctx context.Context, clientUUID string, scopeID string) error {
	return c.removeScope(ctx, c.clientPath(clientUUID)+"/optional-client-scopes/"+url.PathEscape(scopeID))
}

func (c *Client) removeScope(ctx context.Context, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
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
