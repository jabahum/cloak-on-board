import { useEffect, useState } from "react";
import {
  Button,
  Checkbox,
  InlineLoading,
  InlineNotification,
} from "@carbon/react";
import { useNavigate, useParams } from "react-router-dom";
import axios from "axios";
import {
  getApplication,
  updateApplication,
} from "../../api/applications";
import { GeneralStep } from "./wizard/steps/GeneralStep";
import { RolesStep } from "./wizard/steps/RolesStep";
import { UrlsStep } from "./wizard/steps/UrlsStep";
import {
  initialWizardForm,
  type WizardForm,
} from "./wizard/wizardTypes";

export function EditApplicationPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [form, setForm] = useState<WizardForm>(initialWizardForm);
  const [enabled, setEnabled] = useState(true);
  const [linked, setLinked] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) return;
    getApplication(id)
      .then((app) => {
        setForm({
          ...initialWizardForm,
          name: app.name,
          slug: app.slug,
          description: app.description,
          app_type: app.app_type,
          owner_name: app.owner_name,
          owner_email: app.owner_email,
          redirect_uris: app.redirect_uris ?? [],
          web_origins: app.web_origins ?? [],
          roles: app.roles ?? [],
          auto_provision: false,
        });
        setEnabled(app.enabled);
        setLinked(Boolean(app.keycloak_client_uuid));
      })
      .catch((err) => setError(errorMessage(err)))
      .finally(() => setLoading(false));
  }, [id]);

  async function submit() {
    if (!id) return;
    setSaving(true);
    setError("");
    try {
      await updateApplication(id, {
        name: form.name,
        slug: form.slug,
        description: form.description,
        app_type: form.app_type,
        owner_name: form.owner_name,
        owner_email: form.owner_email,
        redirect_uris: form.redirect_uris,
        web_origins: form.web_origins,
        roles: form.roles,
        enabled,
      });
      navigate(`/applications/${id}`);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <InlineLoading description="Loading application..." />;

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Edit application</h1>
        <p className="page-subtitle">
          {linked
            ? "Saving synchronizes these settings to the linked Keycloak client."
            : "Update the local application configuration."}
        </p>
      </div>
      {error && <InlineNotification kind="error" title="Update failed" subtitle={error} />}
      <div className="phase-two-stack">
        <GeneralStep form={form} setForm={setForm} />
        <UrlsStep form={form} setForm={setForm} />
        <RolesStep form={form} setForm={setForm} />
        <Checkbox
          id="application-enabled"
          labelText="Client enabled"
          checked={enabled}
          onChange={(_, data) => setEnabled(Boolean(data.checked))}
        />
      </div>
      <br />
      <div className="wizard-actions">
        <Button kind="secondary" onClick={() => navigate(`/applications/${id}`)}>Cancel</Button>
        <Button disabled={saving || !form.name.trim() || !form.slug.trim()} onClick={submit}>
          {saving ? "Saving..." : linked ? "Save and sync" : "Save"}
        </Button>
      </div>
    </>
  );
}

function errorMessage(err: unknown) {
  if (axios.isAxiosError(err)) {
    return err.response?.data?.error ?? err.message;
  }
  return err instanceof Error ? err.message : "An unexpected error occurred";
}
