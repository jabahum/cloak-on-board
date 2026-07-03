package keycloak

import (
	"context"
	"net/http"
)

type ClientScope struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) AssignDefaultClientScope(ctx context.Context, clientUUID string, scopeID string) error {
	resp, err := c.do(
		ctx,
		http.MethodPut,
		"/admin/realms/"+c.realm+"/clients/"+clientUUID+"/default-client-scopes/"+scopeID,
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
		"/admin/realms/"+c.realm+"/clients/"+clientUUID+"/optional-client-scopes/"+scopeID,
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
