package keycloak

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func testClient(t *testing.T, admin http.HandlerFunc) *Client {
	t.Helper()
	kc := NewClient("http://keycloak.test", "test", "admin", "secret")
	kc.httpClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/realms/test/protocol/openid-connect/token" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"access_token":"token","expires_in":300}`)),
				Request:    r,
			}, nil
		}
		recorder := httptest.NewRecorder()
		w := recorder
		admin(w, r)
		response := recorder.Result()
		response.Request = r
		return response, nil
	})
	return kc
}

func TestListClientsUsesSearchAndDoesNotRequestSecrets(t *testing.T) {
	kc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/admin/realms/test/clients" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("search"); got != "portal app" {
			t.Fatalf("search = %q", got)
		}
		_, _ = w.Write([]byte(`[{"id":"uuid","clientId":"portal","name":"Portal"}]`))
	})

	clients, err := kc.ListClients(context.Background(), "portal app")
	if err != nil {
		t.Fatal(err)
	}
	if len(clients) != 1 || clients[0].ClientID != "portal" {
		t.Fatalf("unexpected clients: %#v", clients)
	}
}

func TestUpdateClientPreservesUnmanagedRepresentationFields(t *testing.T) {
	var update map[string]any
	kc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/realms/test/clients/uuid" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"id":"uuid","clientId":"old","name":"Old","enabled":true,
				"protocol":"openid-connect","rootUrl":"https://preserve.example"
			}`))
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})

	client, err := kc.GetClient(context.Background(), "uuid")
	if err != nil {
		t.Fatal(err)
	}
	client.Name = "Updated"
	if err := kc.UpdateClient(context.Background(), "uuid", client); err != nil {
		t.Fatal(err)
	}
	if update["name"] != "Updated" || update["rootUrl"] != "https://preserve.example" {
		t.Fatalf("update did not preserve representation: %#v", update)
	}
}

func TestClientScopeAssignmentAndRemovalPaths(t *testing.T) {
	requests := []string{}
	kc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	if err := kc.AssignDefaultClientScope(ctx, "client", "scope"); err != nil {
		t.Fatal(err)
	}
	if err := kc.RemoveOptionalClientScope(ctx, "client", "scope"); err != nil {
		t.Fatal(err)
	}
	got := strings.Join(requests, "\n")
	want := "PUT /admin/realms/test/clients/client/default-client-scopes/scope\n" +
		"DELETE /admin/realms/test/clients/client/optional-client-scopes/scope"
	if got != want {
		t.Fatalf("requests:\n%s\nwant:\n%s", got, want)
	}
}

func TestProtocolMapperCRUDMethods(t *testing.T) {
	methods := []string{}
	kc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":"mapper","name":"department","protocol":"openid-connect","protocolMapper":"oidc-usermodel-attribute-mapper","config":{"claim.name":"department"}}]`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	if _, err := kc.ListClientProtocolMappers(ctx, "client"); err != nil {
		t.Fatal(err)
	}
	req := UserAttributeMapper("department", "department", "department")
	if err := kc.CreateClientProtocolMapper(ctx, "client", req); err != nil {
		t.Fatal(err)
	}
	if err := kc.UpdateClientProtocolMapper(ctx, "client", "mapper", req); err != nil {
		t.Fatal(err)
	}
	if err := kc.DeleteClientProtocolMapper(ctx, "client", "mapper"); err != nil {
		t.Fatal(err)
	}
	if len(methods) != 4 {
		t.Fatalf("unexpected requests: %#v", methods)
	}
}

func TestDeleteClientTreatsNotFoundAsSuccess(t *testing.T) {
	kc := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	if err := kc.DeleteClient(context.Background(), "missing"); err != nil {
		t.Fatal(err)
	}
}
