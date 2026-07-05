import type { OnboardingTemplate } from "../../../types/template";

export type WizardForm = {
  template?: OnboardingTemplate;

  name: string;
  slug: string;
  description: string;
  app_type: string;
  owner_name: string;
  owner_email: string;
  redirect_uris: string[];
  web_origins: string[];
  roles: string[];

  auto_provision: boolean;
};

export const initialWizardForm: WizardForm = {
  name: "",
  slug: "",
  description: "",
  app_type: "frontend",
  owner_name: "",
  owner_email: "",
  redirect_uris: ["http://localhost:3000/*"],
  web_origins: ["http://localhost:3000"],
  roles: ["access", "admin", "manager", "viewer"],
  auto_provision: true,
};
