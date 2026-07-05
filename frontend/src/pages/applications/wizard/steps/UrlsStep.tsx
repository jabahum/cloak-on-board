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

export function UrlsStep({ form, setForm }: Props) {
  return (
    <Tile>
      <h3>URLs</h3>
      <p>Configure redirect URIs and web origins used by Keycloak.</p>

      <TextArea
        id="redirect_uris"
        labelText="Redirect URIs"
        value={toText(form.redirect_uris)}
        onChange={(event) =>
          setForm((current) => ({
            ...current,
            redirect_uris: fromText(event.target.value),
          }))
        }
      />

      <br />

      <TextArea
        id="web_origins"
        labelText="Web origins"
        value={toText(form.web_origins)}
        onChange={(event) =>
          setForm((current) => ({
            ...current,
            web_origins: fromText(event.target.value),
          }))
        }
      />
    </Tile>
  );
}
