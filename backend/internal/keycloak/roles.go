package keycloak

import (
	"context"
	"net/http"
)

type CreateClientRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ClientRole  bool   `json:"clientRole"`
}

func (c *Client) CreateClientRole(ctx context.Context, clientUUID string, roleName string) error {
	req := CreateClientRoleRequest{
		Name:       roleName,
		ClientRole: true,
	}

	resp, err := c.do(
		ctx,
		http.MethodPost,
		"/admin/realms/"+c.realm+"/clients/"+clientUUID+"/roles",
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
