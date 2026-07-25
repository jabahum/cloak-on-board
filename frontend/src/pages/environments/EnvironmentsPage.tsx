import { type FormEvent, useEffect, useState } from "react";
import {
  Button,
  Checkbox,
  InlineLoading,
  InlineNotification,
  PasswordInput,
  Select,
  SelectItem,
  Tag,
  TextInput,
  Tile,
} from "@carbon/react";
import axios from "axios";
import {
  createRealmConnection,
  listEnvironments,
  listRealmConnections,
  testRealmConnection,
} from "../../api/delivery";
import type { Environment, RealmConnection } from "../../types/delivery";

export function EnvironmentsPage() {
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [connections, setConnections] = useState<RealmConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    const [environmentData, connectionData] = await Promise.all([
      listEnvironments(),
      listRealmConnections(),
    ]);
    setEnvironments(environmentData);
    setConnections(connectionData);
  }

  useEffect(() => {
    Promise.all([listEnvironments(), listRealmConnections()])
      .then(([environmentData, connectionData]) => {
        setEnvironments(environmentData);
        setConnections(connectionData);
      })
      .catch(() => setError("Failed to load managed environments"))
      .finally(() => setLoading(false));
  }, []);

  async function addConnection(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      await createRealmConnection({
        environment_id: String(form.get("environment_id")),
        name: String(form.get("name")),
        base_url: String(form.get("base_url")),
        realm: String(form.get("realm")),
        admin_client_id: String(form.get("admin_client_id")),
        admin_client_secret: String(form.get("admin_client_secret")),
      });
      event.currentTarget.reset();
      await load();
    } catch (err) {
      setError(axios.isAxiosError(err) ? (err.response?.data?.error ?? err.message) : "Connection creation failed");
    } finally {
      setSaving(false);
    }
  }

  async function testConnection(id: string) {
    setError("");
    try {
      await testRealmConnection(id);
      await load();
    } catch (err) {
      setError(axios.isAxiosError(err) ? (err.response?.data?.error ?? err.message) : "Connection test failed");
      await load();
    }
  }

  if (loading) return <InlineLoading description="Loading environments..." />;

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Environments & realms</h1>
        <p className="page-subtitle">
          Managed target realms are separate from the onboarder authentication realm.
        </p>
      </div>
      {error && <InlineNotification kind="error" title="Environment operation failed" subtitle={error} />}

      <div className="delivery-grid">
        {environments.map((environment) => (
          <Tile key={environment.id}>
            <h3>{environment.name}</h3>
            <p>Promotion order: {environment.promotion_order}</p>
            <Tag type={environment.protected ? "purple" : "gray"}>
              {environment.protected ? "Approval protected" : "Unprotected"}
            </Tag>
            <div className="phase-two-stack connection-list">
              {connections.filter((item) => item.environment_id === environment.id).map((connection) => (
                <div key={connection.id}>
                  <strong>{connection.name}</strong>
                  <p>{connection.base_url}/realms/{connection.realm}</p>
                  <Tag type={connection.last_test_status === "succeeded" ? "green" : connection.last_test_status === "failed" ? "red" : "gray"}>
                    {connection.last_test_status || "Not tested"}
                  </Tag>{" "}
                  <Button size="sm" kind="ghost" onClick={() => void testConnection(connection.id)}>Test</Button>
                </div>
              ))}
            </div>
          </Tile>
        ))}
      </div>

      <Tile className="delivery-form">
        <h3>Add realm connection</h3>
        <form onSubmit={addConnection} className="phase-two-stack">
          <Select id="environment_id" name="environment_id" labelText="Environment" required>
            <SelectItem value="" text="Choose an environment" />
            {environments.map((item) => <SelectItem key={item.id} value={item.id} text={item.name} />)}
          </Select>
          <TextInput id="connection_name" name="name" labelText="Connection name" required />
          <TextInput id="connection_url" name="base_url" labelText="Keycloak base URL" placeholder="http://keycloak:8080" required />
          <TextInput id="connection_realm" name="realm" labelText="Managed realm" required />
          <TextInput id="connection_client" name="admin_client_id" labelText="Admin service-account client ID" required />
          <PasswordInput id="connection_secret" name="admin_client_secret" labelText="Admin client secret" required />
          <Checkbox id="connection_notice" labelText="I understand this credential is encrypted and never returned by the API." checked readOnly />
          <Button type="submit" disabled={saving}>{saving ? "Saving..." : "Add connection"}</Button>
        </form>
      </Tile>
    </>
  );
}
