import { TextArea, Tile } from "@carbon/react";
import type { WizardForm } from "../wizardTypes";

type Props = {
  form: WizardForm;
  setForm: React.Dispatch<React.SetStateAction<WizardForm>>;
};

function toText(values: string[]) {
  return values.join("\n");
}

function fromText(value: string) {
  return value
    .split("\n")
    .map((item) => item.trim())
    .filter(Boolean);
}

export function RolesStep({ form, setForm }: Props) {
  return (
    <Tile>
      <h3>Roles</h3>
      <p>Define client roles that should be created in Keycloak.</p>

      <TextArea
        id="roles"
        labelText="Roles"
        value={toText(form.roles)}
        onChange={(event) =>
          setForm((current) => ({
            ...current,
            roles: fromText(event.target.value),
          }))
        }
      />
    </Tile>
  );
}
