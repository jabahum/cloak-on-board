package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type CreateClientRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ClientRole  bool   `json:"clientRole"`
}

type ClientRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) ListClientRoles(ctx context.Context, clientUUID string) ([]ClientRole, error) {
	resp, err := c.do(ctx, http.MethodGet, c.clientPath(clientUUID)+"/roles", nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeError(resp)
	}
	defer resp.Body.Close()

	roles := []ClientRole{}
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (c *Client) CreateClientRole(ctx context.Context, clientUUID string, roleName string) error {
	req := CreateClientRoleRequest{
		Name:       roleName,
		ClientRole: true,
	}

	resp, err := c.do(
		ctx,
		http.MethodPost,
		c.clientPath(clientUUID)+"/roles",
		req,
	)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusConflict {
		resp.Body.Close()
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}

	resp.Body.Close()
	return nil
}

func (c *Client) CreateClientRoles(ctx context.Context, clientUUID string, roles []string) error {
	for _, role := range roles {
		if role == "" {
			continue
		}

		if err := c.CreateClientRole(ctx, clientUUID, role); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteClientRole(ctx context.Context, clientUUID string, roleName string) error {
	resp, err := c.do(ctx, http.MethodDelete, c.clientPath(clientUUID)+"/roles/"+url.PathEscape(roleName), nil)
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

func (c *Client) SyncClientRoles(ctx context.Context, clientUUID string, desired []string) error {
	current, err := c.ListClientRoles(ctx, clientUUID)
	if err != nil {
		return err
	}
	want := map[string]bool{}
	for _, role := range desired {
		if role != "" {
			want[role] = true
		}
	}
	have := map[string]bool{}
	for _, role := range current {
		have[role.Name] = true
		if !want[role.Name] {
			if err := c.DeleteClientRole(ctx, clientUUID, role.Name); err != nil {
				return err
			}
		}
	}
	for role := range want {
		if !have[role] {
			if err := c.CreateClientRole(ctx, clientUUID, role); err != nil {
				return err
			}
		}
	}
	return nil
}
