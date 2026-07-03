export type OnboardingTemplate = {
  id: string;
  name: string;
  description: string;
  app_type: string;
  default_roles: string[];
  default_redirect_uris: string[];
  default_web_origins: string[];
  default_scopes: string[];
  default_mappers: string[];
  client_config: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};
