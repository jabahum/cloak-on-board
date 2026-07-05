import { Checkbox, Tag, Tile } from "@carbon/react";
import type { WizardForm } from "../wizardTypes";

type Props = {
  form: WizardForm;
  setForm: React.Dispatch<React.SetStateAction<WizardForm>>;
};

export function ReviewStep({ form, setForm }: Props) {
  return (
    <Tile>
      <h3>Review</h3>

      <p>
        <strong>Name:</strong> {form.name}
      </p>

      <p>
        <strong>Client ID:</strong> {form.slug}
      </p>

      <p>
        <strong>Type:</strong> {form.app_type}
      </p>

      <p>
        <strong>Owner:</strong> {form.owner_name || "-"}
      </p>

      <h4>Redirect URIs</h4>
      {form.redirect_uris.map((uri) => (
        <p key={uri}>
          <code>{uri}</code>
        </p>
      ))}

      <h4>Web Origins</h4>
      {form.web_origins.map((origin) => (
        <p key={origin}>
          <code>{origin}</code>
        </p>
      ))}

      <h4>Roles</h4>
      {form.roles.map((role) => (
        <Tag key={role}>{role}</Tag>
      ))}

      <br />

      <Checkbox
        id="auto_provision"
        labelText="Provision to Keycloak immediately after creating"
        checked={form.auto_provision}
        onChange={(_, data) =>
          setForm((current) => ({
            ...current,
            auto_provision: Boolean(data.checked),
          }))
        }
      />
    </Tile>
  );
}
