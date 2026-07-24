export type Application = {
  id: string;
  name: string;
  slug: string;
  description: string;
  app_type: string;
  owner_name: string;
  owner_email: string;
  status: string;
  source: "created" | "imported";
  enabled: boolean;
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

export type ApplicationPayload = {
  name: string;
  slug: string;
  description: string;
  app_type: string;
  owner_name: string;
  owner_email: string;
  redirect_uris: string[];
  web_origins: string[];
  roles: string[];
  enabled?: boolean;
};

export type CreateApplicationPayload = ApplicationPayload;
export type UpdateApplicationPayload = ApplicationPayload;

export type KeycloakClient = {
  id: string;
  clientId: string;
  name: string;
  description: string;
  publicClient: boolean;
  serviceAccountsEnabled: boolean;
  standardFlowEnabled: boolean;
  redirectUris?: string[];
  webOrigins?: string[];
  enabled: boolean;
  imported?: boolean;
};

export type ImportApplicationPayload = {
  keycloak_client_uuid: string;
  name?: string;
  description?: string;
  app_type?: string;
  owner_name?: string;
  owner_email?: string;
};

export type ClientScope = {
  id: string;
  name: string;
};

export type ClientScopeAssignments = {
  default: ClientScope[];
  optional: ClientScope[];
  available: ClientScope[];
};

export type ProtocolMapper = {
  id: string;
  name: string;
  protocol: string;
  protocolMapper: string;
  config: Record<string, string>;
};

export type ProtocolMapperPayload = Omit<ProtocolMapper, "id">;
