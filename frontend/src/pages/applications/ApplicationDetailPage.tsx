import { useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Tag,
  Tile,
} from "@carbon/react";
import { useParams } from "react-router-dom";
import { getApplication, provisionApplication } from "../../api/applications";
import type { Application } from "../../types/application";
import type { ProvisioningJob } from "../../types/provisioning";
import axios from "axios";

export function ApplicationDetailPage() {
  const { id } = useParams();
  const [app, setApp] = useState<Application | null>(null);
  const [job, setJob] = useState<ProvisioningJob | null>(null);
  const [loading, setLoading] = useState(true);
  const [provisioning, setProvisioning] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    if (!id) return;

    setLoading(true);
    getApplication(id)
      .then(setApp)
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    load();
  }, [id]);

  async function handleProvision() {
    if (!id) return;

    setError("");
    setProvisioning(true);

    try {
      const result = await provisionApplication(id);
      setJob(result);
      await load();
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error ?? err.message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setProvisioning(false);
    }
  }

  if (loading) return <InlineLoading description="Loading application..." />;

  if (!app)
    return (
      <InlineNotification
        kind="error"
        title="Not found"
        subtitle="Application not found"
      />
    );

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">{app.name}</h1>
        <p className="page-subtitle">{app.description}</p>
      </div>

      {error && (
        <InlineNotification kind="error" title="Error" subtitle={error} />
      )}

      <Tile>
        <p>
          <strong>Client ID:</strong> {app.slug}
        </p>
        <p>
          <strong>Type:</strong> {app.app_type}
        </p>
        <p>
          <strong>Status:</strong>{" "}
          <Tag
            type={
              app.status === "provisioned"
                ? "green"
                : app.status === "failed"
                  ? "red"
                  : "gray"
            }
          >
            {app.status}
          </Tag>
        </p>
        <p>
          <strong>Owner:</strong> {app.owner_name || "-"}
        </p>

        {app.keycloak_client_uuid && (
          <p>
            <strong>Keycloak UUID:</strong> {app.keycloak_client_uuid}
          </p>
        )}

        {app.keycloak_client_secret && (
          <p>
            <strong>Client Secret:</strong>{" "}
            <code>{app.keycloak_client_secret}</code>
          </p>
        )}
      </Tile>

      <br />

      <Tile>
        <h4>Redirect URIs</h4>
        {(app.redirect_uris ?? []).map((uri) => (
          <p key={uri}>
            <code>{uri}</code>
          </p>
        ))}

        <h4>Web Origins</h4>
        {(app.web_origins ?? []).map((origin) => (
          <p key={origin}>
            <code>{origin}</code>
          </p>
        ))}

        <h4>Roles</h4>
        {(app.roles ?? []).map((role) => (
          <Tag key={role}>{role}</Tag>
        ))}
      </Tile>

      <br />

      <Button onClick={handleProvision} disabled={provisioning}>
        {provisioning ? "Provisioning..." : "Provision to Keycloak"}
      </Button>

      {job && (
        <>
          <br />
          <br />
          <Tile>
            <h4>Latest Provisioning Job</h4>
            <p>
              <strong>Status:</strong> {job.status}
            </p>
            {(job.steps ?? []).map((step) => (
              <p key={step.id}>
                {step.step_name}: <strong>{step.status}</strong>
                {step.error_message ? ` - ${step.error_message}` : ""}
              </p>
            ))}
          </Tile>
        </>
      )}
    </>
  );
}
