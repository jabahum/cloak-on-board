import { useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  TextInput,
  Tile,
} from "@carbon/react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import {
  importApplication,
  listKeycloakClients,
} from "../../api/applications";
import type { KeycloakClient } from "../../types/application";

export function ImportApplicationPage() {
  const navigate = useNavigate();
  const [clients, setClients] = useState<KeycloakClient[]>([]);
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<KeycloakClient | null>(null);
  const [form, setForm] = useState({
    name: "",
    description: "",
    app_type: "backend",
    owner_name: "",
    owner_email: "",
  });
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      setClients(await listKeycloakClients(search));
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    listKeycloakClients()
      .then(setClients)
      .catch((err) => setError(errorMessage(err)))
      .finally(() => setLoading(false));
  }, []);

  function selectClient(client: KeycloakClient) {
    setSelected(client);
    setForm({
      name: client.name || client.clientId,
      description: client.description || "",
      app_type: inferType(client),
      owner_name: "",
      owner_email: "",
    });
  }

  async function submit() {
    if (!selected) return;
    setSubmitting(true);
    setError("");
    try {
      const app = await importApplication({
        keycloak_client_uuid: selected.id,
        ...form,
      });
      navigate(`/applications/${app.id}`);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Import from Keycloak</h1>
        <p className="page-subtitle">
          Link an existing realm client without creating a duplicate.
        </p>
      </div>
      {error && <InlineNotification kind="error" title="Import" subtitle={error} />}

      {!selected ? (
        <>
          <div className="inline-form">
            <TextInput
              id="keycloak-client-search"
              labelText="Search clients"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") void load();
              }}
            />
            <Button onClick={load}>Search</Button>
          </div>
          <br />
          {loading ? (
            <InlineLoading description="Loading Keycloak clients..." />
          ) : (
            <div className="template-grid">
              {clients.map((client) => (
                <Tile key={client.id}>
                  <h4>{client.name || client.clientId}</h4>
                  <p><strong>Client ID:</strong> {client.clientId}</p>
                  <p>{client.description || "No description"}</p>
                  <Button
                    size="sm"
                    disabled={client.imported}
                    onClick={() => selectClient(client)}
                  >
                    {client.imported ? "Already imported" : "Review import"}
                  </Button>
                </Tile>
              ))}
              {clients.length === 0 && <Tile>No clients found.</Tile>}
            </div>
          )}
        </>
      ) : (
        <Tile>
          <h3>Review import</h3>
          <p><strong>Keycloak client ID:</strong> {selected.clientId}</p>
          <div className="phase-two-stack">
            <TextInput id="import-name" labelText="Application name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            <TextInput id="import-description" labelText="Description" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
            <Select id="import-type" labelText="Application type" value={form.app_type} onChange={(e) => setForm({ ...form, app_type: e.target.value })}>
              <SelectItem value="frontend" text="Frontend" />
              <SelectItem value="backend" text="Backend" />
              <SelectItem value="mobile" text="Mobile" />
              <SelectItem value="machine_to_machine" text="Machine to machine" />
            </Select>
            <TextInput id="import-owner" labelText="Owner name" value={form.owner_name} onChange={(e) => setForm({ ...form, owner_name: e.target.value })} />
            <TextInput id="import-owner-email" type="email" labelText="Owner email" value={form.owner_email} onChange={(e) => setForm({ ...form, owner_email: e.target.value })} />
          </div>
          <br />
          <div className="wizard-actions">
            <Button kind="secondary" onClick={() => setSelected(null)}>Back</Button>
            <Button disabled={submitting || !form.name.trim()} onClick={submit}>
              {submitting ? "Importing..." : "Import client"}
            </Button>
          </div>
        </Tile>
      )}
    </>
  );
}

function inferType(client: KeycloakClient) {
  if (client.serviceAccountsEnabled) return "machine_to_machine";
  if (client.publicClient) return "frontend";
  return "backend";
}

function errorMessage(err: unknown) {
  if (axios.isAxiosError(err)) {
    return err.response?.data?.error ?? err.message;
  }
  return err instanceof Error ? err.message : "An unexpected error occurred";
}
