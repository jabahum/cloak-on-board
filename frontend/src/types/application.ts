export type Application = {
  id: string;
  name: string;
  slug: string;
  description: string;
  app_type: string;
  owner_name: string;
  owner_email: string;
  status: string;
  redirect_uris?: string[];
  web_origins?: string[];
  roles?: string[];
  keycloak_client_uuid?: string;
  keycloak_client_id?: string;
  keycloak_client_secret?: string;
  provisioned_at?: string;
  created_at: string;
  updated_at: string;
};

export type CreateApplicationPayload = {
  name: string;
  slug: string;
  description: string;
  app_type: string;
  owner_name: string;
  owner_email: string;
  redirect_uris: string[];
  web_origins: string[];
  roles: string[];
};
