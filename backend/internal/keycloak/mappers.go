package keycloak

import (
	"context"
	"net/http"
)

type ProtocolMapperRequest struct {
	Name           string            `json:"name"`
	Protocol       string            `json:"protocol"`
	ProtocolMapper string            `json:"protocolMapper"`
	Config         map[string]string `json:"config"`
}

func (c *Client) CreateClientProtocolMapper(ctx context.Context, clientUUID string, mapper ProtocolMapperRequest) error {
	resp, err := c.do(
		ctx,
		http.MethodPost,
		"/admin/realms/"+c.realm+"/clients/"+clientUUID+"/protocol-mappers/models",
		mapper,
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
func (c *Client) CreateDefaultProtocolMappers(ctx context.Context, clientUUID string) error {
	for _, mapper := range DefaultProtocolMappers() {
		if err := c.CreateClientProtocolMapper(ctx, clientUUID, mapper); err != nil {
			return err
		}
	}

	return nil
}

func DefaultProtocolMappers() []ProtocolMapperRequest {
	return []ProtocolMapperRequest{
		UserAttributeMapper("user_id", "user_id", "user_id"),
		UserAttributeMapper("username", "username", "preferred_username"),
		UserAttributeMapper("full_name", "full_name", "name"),
		UserAttributeMapper("district", "district", "district"),
		UserAttributeMapper("facility", "facility", "facility"),
		ClientRolesMapper(),
		RealmRolesMapper(),
	}
}

func UserAttributeMapper(name string, userAttribute string, claimName string) ProtocolMapperRequest {
	return ProtocolMapperRequest{
		Name:           name,
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usermodel-attribute-mapper",
		Config: map[string]string{
			"user.attribute":       userAttribute,
			"claim.name":           claimName,
			"jsonType.label":       "String",
			"id.token.claim":       "true",
			"access.token.claim":   "true",
			"userinfo.token.claim": "true",
			"multivalued":          "false",
			"aggregate.attrs":      "false",
		},
	}
}

func ClientRolesMapper() ProtocolMapperRequest {
	return ProtocolMapperRequest{
		Name:           "client roles",
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usermodel-client-role-mapper",
		Config: map[string]string{
			"claim.name":           "resource_access.${client_id}.roles",
			"jsonType.label":       "String",
			"multivalued":          "true",
			"id.token.claim":       "true",
			"access.token.claim":   "true",
			"userinfo.token.claim": "true",
		},
	}
}

func RealmRolesMapper() ProtocolMapperRequest {
	return ProtocolMapperRequest{
		Name:           "realm roles",
		Protocol:       "openid-connect",
		ProtocolMapper: "oidc-usermodel-realm-role-mapper",
		Config: map[string]string{
			"claim.name":           "realm_access.roles",
			"jsonType.label":       "String",
			"multivalued":          "true",
			"id.token.claim":       "true",
			"access.token.claim":   "true",
			"userinfo.token.claim": "true",
		},
	}
}
