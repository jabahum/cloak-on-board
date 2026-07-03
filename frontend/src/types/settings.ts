export type Settings = {
  id?: string;
  keycloak_base_url: string;
  keycloak_realm: string;
  keycloak_admin_client_id: string;
  created_at?: string;
  updated_at?: string;
};

export type SaveSettingsPayload = {
  keycloak_base_url: string;
  keycloak_realm: string;
  keycloak_admin_client_id: string;
  keycloak_admin_client_secret: string;
};
