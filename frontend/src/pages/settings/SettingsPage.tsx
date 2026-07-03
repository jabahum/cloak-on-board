import { type FormEvent, useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  PasswordInput,
  TextInput,
} from "@carbon/react";
import { getSettings, saveSettings } from "../../api/settings";
import type { Settings } from "../../types/settings";
import axios from "axios";

export function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    getSettings()
      .then(setSettings)
      .catch(() => setError("Failed to load settings"))
      .finally(() => setLoading(false));
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSuccess("");
    setSaving(true);

    const form = new FormData(event.currentTarget);

    try {
      const saved = await saveSettings({
        keycloak_base_url: String(form.get("keycloak_base_url") ?? ""),
        keycloak_realm: String(form.get("keycloak_realm") ?? ""),
        keycloak_admin_client_id: String(
          form.get("keycloak_admin_client_id") ?? "",
        ),
        keycloak_admin_client_secret: String(
          form.get("keycloak_admin_client_secret") ?? "",
        ),
      });

      setSettings(saved);
      setSuccess("Settings saved successfully");
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error ?? err.message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <InlineLoading description="Loading settings..." />;

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Settings</h1>
        <p className="page-subtitle">Configure Keycloak connection details.</p>
      </div>

      {error && (
        <InlineNotification kind="error" title="Error" subtitle={error} />
      )}
      {success && (
        <InlineNotification kind="success" title="Saved" subtitle={success} />
      )}

      <form onSubmit={handleSubmit} style={{ maxWidth: 720 }}>
        <TextInput
          id="keycloak_base_url"
          name="keycloak_base_url"
          labelText="Keycloak URL"
          defaultValue={settings?.keycloak_base_url ?? ""}
          placeholder="http://keycloak:8080"
          required
        />
        <br />

        <TextInput
          id="keycloak_realm"
          name="keycloak_realm"
          labelText="Realm"
          defaultValue={settings?.keycloak_realm ?? ""}
          placeholder="master"
          required
        />
        <br />

        <TextInput
          id="keycloak_admin_client_id"
          name="keycloak_admin_client_id"
          labelText="Admin client ID"
          defaultValue={settings?.keycloak_admin_client_id ?? ""}
          placeholder="onboarder-admin"
          required
        />
        <br />

        <PasswordInput
          id="keycloak_admin_client_secret"
          name="keycloak_admin_client_secret"
          labelText="Admin client secret"
          placeholder="Paste secret"
        />
        <br />

        <Button type="submit" disabled={saving}>
          {saving ? "Saving..." : "Save settings"}
        </Button>
      </form>
    </>
  );
}
