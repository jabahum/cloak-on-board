import { useState, type FormEvent } from "react";
import {
  Button,
  TextInput,
  Select,
  SelectItem,
  TextArea,
  InlineNotification,
} from "@carbon/react";
import { useNavigate } from "react-router-dom";
import { createApplication } from "../../api/applications";
import axios from "axios";

export function ApplicationCreatePage() {
  const navigate = useNavigate();
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");

    const form = new FormData(event.currentTarget);

    try {
      const app = await createApplication({
        name: String(form.get("name") ?? ""),
        slug: String(form.get("slug") ?? ""),
        description: String(form.get("description") ?? ""),
        app_type: String(form.get("app_type") ?? ""),
        owner_name: String(form.get("owner_name") ?? ""),
        owner_email: String(form.get("owner_email") ?? ""),
        redirect_uris: String(form.get("redirect_uris") ?? "")
          .split("\n")
          .map((v) => v.trim())
          .filter(Boolean),
        web_origins: String(form.get("web_origins") ?? "")
          .split("\n")
          .map((v) => v.trim())
          .filter(Boolean),
        roles: String(form.get("roles") ?? "")
          .split("\n")
          .map((v) => v.trim())
          .filter(Boolean),
      });

      navigate(`/applications/${app.id}`);
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error ?? err.message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred");
      }
    }
  }

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Create Application</h1>
        <p className="page-subtitle">
          Enter the application details and Keycloak defaults.
        </p>
      </div>

      {error && (
        <InlineNotification kind="error" title="Error" subtitle={error} />
      )}

      <form onSubmit={handleSubmit} style={{ maxWidth: 760 }}>
        <TextInput
          id="name"
          name="name"
          labelText="Application name"
          placeholder="Health BI"
          required
        />
        <br />

        <TextInput
          id="slug"
          name="slug"
          labelText="Client ID / slug"
          placeholder="health-bi"
          required
        />
        <br />

        <TextArea id="description" name="description" labelText="Description" />
        <br />

        <Select
          id="app_type"
          name="app_type"
          labelText="Application type"
          defaultValue="frontend"
        >
          <SelectItem value="frontend" text="Frontend" />
          <SelectItem value="backend" text="Backend API" />
          <SelectItem value="mobile" text="Mobile" />
          <SelectItem value="machine_to_machine" text="Machine-to-machine" />
        </Select>
        <br />

        <TextInput id="owner_name" name="owner_name" labelText="Owner name" />
        <br />

        <TextInput
          id="owner_email"
          name="owner_email"
          labelText="Owner email"
        />
        <br />

        <TextArea
          id="redirect_uris"
          name="redirect_uris"
          labelText="Redirect URIs"
          defaultValue="http://localhost:3000/*"
        />
        <br />

        <TextArea
          id="web_origins"
          name="web_origins"
          labelText="Web origins"
          defaultValue="http://localhost:3000"
        />
        <br />

        <TextArea
          id="roles"
          name="roles"
          labelText="Roles"
          defaultValue={"access\nadmin\nmanager\nviewer"}
        />
        <br />

        <Button type="submit">Create application</Button>
      </form>
    </>
  );
}
