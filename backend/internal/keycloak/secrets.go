package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
)

type CredentialRepresentation struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (c *Client) GetClientSecret(ctx context.Context, clientUUID string) (string, error) {
	resp, err := c.do(
		ctx,
		http.MethodGet,
		"/admin/realms/"+c.realm+"/clients/"+clientUUID+"/client-secret",
		nil,
	)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", decodeError(resp)
	}
	defer resp.Body.Close()

	var credential CredentialRepresentation

	if err := json.NewDecoder(resp.Body).Decode(&credential); err != nil {
		return "", err
	}

	return credential.Value, nil
}
