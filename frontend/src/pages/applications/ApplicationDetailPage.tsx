import { useEffect, useState } from "react";
import {
  Button,
  Checkbox,
  InlineLoading,
  InlineNotification,
  Modal,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Tag,
  TextInput,
  Tile,
} from "@carbon/react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  deleteApplication,
  getApplication,
  provisionApplication,
} from "../../api/applications";
import type { Application } from "../../types/application";
import type { ProvisioningJob } from "../../types/provisioning";
import axios from "axios";
import { ClientScopesTab } from "./ClientScopesTab";
import { ProtocolMappersTab } from "./ProtocolMappersTab";

export function ApplicationDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();

  const [app, setApp] = useState<Application | null>(null);
  const [job, setJob] = useState<ProvisioningJob | null>(null);
  const [loading, setLoading] = useState(true);
  const [provisioning, setProvisioning] = useState(false);
  const [error, setError] = useState("");
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteFromKeycloak, setDeleteFromKeycloak] = useState(false);
  const [deleteConfirmation, setDeleteConfirmation] = useState("");
  const [deleting, setDeleting] = useState(false);

  async function load() {
    if (!id) return;

    setLoading(true);
    try {
      const result = await getApplication(id);
      setApp(result);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (!id) return;
    getApplication(id)
      .then(setApp)
      .catch((err) => {
        if (axios.isAxiosError(err)) {
          setError(err.response?.data?.error ?? err.message);
        } else {
          setError(err instanceof Error ? err.message : "Failed to load application");
        }
      })
      .finally(() => setLoading(false));
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

  async function handleDelete() {
    if (!id || !app) return;
    setDeleting(true);
    setError("");
    try {
      await deleteApplication(id, deleteFromKeycloak);
      navigate("/applications");
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error ?? err.message);
      } else {
        setError(err instanceof Error ? err.message : "Delete failed");
      }
      setDeleteOpen(false);
    } finally {
      setDeleting(false);
    }
  }

  if (loading) {
    return <InlineLoading description="Loading application..." />;
  }

  if (!app) {
    return (
      <InlineNotification
        kind="error"
        title="Not found"
        subtitle="Application not found"
      />
    );
  }

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">{app.name}</h1>
        <p className="page-subtitle">
          {app.description || "Application details"}
        </p>
        <div className="detail-actions">
          <Button as={Link} kind="secondary" to={`/applications/${app.id}/edit`}>
            Edit
          </Button>
          <Button kind="danger--tertiary" onClick={() => setDeleteOpen(true)}>
            Delete
          </Button>
        </div>
      </div>

      {error && (
        <InlineNotification kind="error" title="Error" subtitle={error} />
      )}

      <Tabs>
        <TabList aria-label="Application detail tabs">
          <Tab>Overview</Tab>
          <Tab>URLs</Tab>
          <Tab>Roles</Tab>
          <Tab>Client scopes</Tab>
          <Tab>Protocol mappers</Tab>
          <Tab>Secrets</Tab>
          <Tab>Provisioning</Tab>
        </TabList>

        <TabPanels>
          <TabPanel>
            <OverviewTab
              app={app}
              onProvision={handleProvision}
              provisioning={provisioning}
            />
          </TabPanel>

          <TabPanel>
            <UrlsTab app={app} />
          </TabPanel>

          <TabPanel>
            <RolesTab app={app} />
          </TabPanel>

          <TabPanel>
            <ClientScopesTab applicationId={app.id} />
          </TabPanel>

          <TabPanel>
            <ProtocolMappersTab applicationId={app.id} />
          </TabPanel>

          <TabPanel>
            <SecretsTab app={app} />
          </TabPanel>

          <TabPanel>
            <ProvisioningTab job={job} />
          </TabPanel>
        </TabPanels>
      </Tabs>

      <Modal
        danger
        open={deleteOpen}
        modalHeading="Delete application?"
        primaryButtonText={deleting ? "Deleting..." : "Delete"}
        secondaryButtonText="Cancel"
        primaryButtonDisabled={
          deleting ||
          (deleteFromKeycloak && deleteConfirmation !== app.slug)
        }
        onRequestSubmit={handleDelete}
        onRequestClose={() => setDeleteOpen(false)}
      >
        <div className="phase-two-stack">
          <p>
            Removing the application locally leaves the existing Keycloak
            client untouched.
          </p>
          <Checkbox
            id="delete-from-keycloak"
            labelText="Also permanently delete the linked Keycloak client"
            checked={deleteFromKeycloak}
            disabled={!app.keycloak_client_uuid}
            onChange={(_, data) =>
              setDeleteFromKeycloak(Boolean(data.checked))
            }
          />
          {deleteFromKeycloak && (
            <TextInput
              id="delete-client-confirmation"
              labelText={`Type ${app.slug} to confirm`}
              value={deleteConfirmation}
              onChange={(event) => setDeleteConfirmation(event.target.value)}
            />
          )}
        </div>
      </Modal>
    </>
  );
}

function OverviewTab({
  app,
  onProvision,
  provisioning,
}: {
  app: Application;
  onProvision: () => void;
  provisioning: boolean;
}) {
  return (
    <>
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
          <strong>Source:</strong> {app.source}
        </p>

        <p>
          <strong>Enabled:</strong> {app.enabled ? "Yes" : "No"}
        </p>

        <p>
          <strong>Owner:</strong> {app.owner_name || "-"}
        </p>

        <p>
          <strong>Owner Email:</strong> {app.owner_email || "-"}
        </p>

        {app.keycloak_client_uuid && (
          <p>
            <strong>Keycloak UUID:</strong> {app.keycloak_client_uuid}
          </p>
        )}

        {app.keycloak_client_id && (
          <p>
            <strong>Provisioned Client ID:</strong> {app.keycloak_client_id}
          </p>
        )}

        {app.provisioned_at && (
          <p>
            <strong>Provisioned At:</strong>{" "}
            {new Date(app.provisioned_at).toLocaleString()}
          </p>
        )}
      </Tile>

      <br />

      <Button onClick={onProvision} disabled={provisioning}>
        {provisioning ? "Provisioning..." : "Provision to Keycloak"}
      </Button>
    </>
  );
}

function UrlsTab({ app }: { app: Application }) {
  return (
    <Tile>
      <h4>Redirect URIs</h4>

      {(app.redirect_uris ?? []).length === 0 ? (
        <p>No redirect URIs configured.</p>
      ) : (
        app.redirect_uris?.map((uri) => (
          <p key={uri}>
            <code>{uri}</code>
          </p>
        ))
      )}

      <br />

      <h4>Web Origins</h4>

      {(app.web_origins ?? []).length === 0 ? (
        <p>No web origins configured.</p>
      ) : (
        app.web_origins?.map((origin) => (
          <p key={origin}>
            <code>{origin}</code>
          </p>
        ))
      )}
    </Tile>
  );
}

function RolesTab({ app }: { app: Application }) {
  return (
    <Tile>
      <h4>Client Roles</h4>

      {(app.roles ?? []).length === 0 ? (
        <p>No roles configured.</p>
      ) : (
        app.roles?.map((role) => <Tag key={role}>{role}</Tag>)
      )}
    </Tile>
  );
}

function SecretsTab({ app }: { app: Application }) {
  return (
    <Tile>
      <h4>Client Secret</h4>

      {app.keycloak_client_secret ? (
        <p>
          <code>{app.keycloak_client_secret}</code>
        </p>
      ) : (
        <p>This application does not have a stored client secret.</p>
      )}
    </Tile>
  );
}

function ProvisioningTab({ job }: { job: ProvisioningJob | null }) {
  if (!job) {
    return (
      <Tile>
        <p>No provisioning job loaded in this session.</p>
      </Tile>
    );
  }

  return (
    <Tile>
      <h4>Latest Provisioning Job</h4>

      <p>
        <strong>Status:</strong>{" "}
        <Tag
          type={
            job.status === "succeeded"
              ? "green"
              : job.status === "failed"
                ? "red"
                : "blue"
          }
        >
          {job.status}
        </Tag>
      </p>

      {(job.steps ?? []).map((step) => (
        <p key={step.id}>
          {step.step_name}: <strong>{step.status}</strong>
          {step.error_message ? ` - ${step.error_message}` : ""}
        </p>
      ))}
    </Tile>
  );
}
