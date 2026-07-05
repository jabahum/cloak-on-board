import { Select, SelectItem, TextArea, TextInput, Tile } from "@carbon/react";
import type { WizardForm } from "../wizardTypes";
import { toSlug } from "../../../../utils/slug";

type Props = {
  form: WizardForm;
  setForm: React.Dispatch<React.SetStateAction<WizardForm>>;
};

export function GeneralStep({ form, setForm }: Props) {
  function update<K extends keyof WizardForm>(key: K, value: WizardForm[K]) {
    setForm((current) => ({
      ...current,
      [key]: value,
    }));
  }

  function updateName(value: string) {
    setForm((current) => ({
      ...current,
      name: value,
      slug: current.slug ? current.slug : toSlug(value),
    }));
  }

  return (
    <Tile>
      <h3>General details</h3>

      <TextInput
        id="name"
        labelText="Application name"
        value={form.name}
        placeholder="Health BI"
        onChange={(event) => updateName(event.target.value)}
        required
      />

      <br />

      <TextInput
        id="slug"
        labelText="Client ID / slug"
        value={form.slug}
        placeholder="health-bi"
        onChange={(event) => update("slug", toSlug(event.target.value))}
        required
      />

      <br />

      <TextArea
        id="description"
        labelText="Description"
        value={form.description}
        placeholder="Brief description"
        onChange={(event) => update("description", event.target.value)}
      />

      <br />

      <Select
        id="app_type"
        labelText="Application type"
        value={form.app_type}
        onChange={(event) => update("app_type", event.target.value)}
      >
        <SelectItem value="frontend" text="Frontend" />
        <SelectItem value="backend" text="Backend API" />
        <SelectItem value="mobile" text="Mobile" />
        <SelectItem value="machine_to_machine" text="Machine-to-machine" />
      </Select>

      <br />

      <TextInput
        id="owner_name"
        labelText="Owner name"
        value={form.owner_name}
        onChange={(event) => update("owner_name", event.target.value)}
      />

      <br />

      <TextInput
        id="owner_email"
        labelText="Owner email"
        value={form.owner_email}
        onChange={(event) => update("owner_email", event.target.value)}
      />
    </Tile>
  );
}
