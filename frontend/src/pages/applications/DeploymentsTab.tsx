import { useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Tag,
  Tile,
} from "@carbon/react";
import axios from "axios";
import {
  createSnapshot,
  deployApplication,
  listDeployments,
  listEnvironments,
  listRealmConnections,
  listSnapshots,
  promoteApplication,
} from "../../api/delivery";
import { submitApproval } from "../../api/approvals";
import type { Deployment, Environment, RealmConnection, Snapshot } from "../../types/delivery";
import { useAuth } from "../../auth/AuthContext";

export function DeploymentsTab({ applicationId }: { applicationId: string }) {
  const { can } = useAuth();
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [connections, setConnections] = useState<RealmConnection[]>([]);
  const [snapshots, setSnapshots] = useState<Snapshot[]>([]);
  const [environmentId, setEnvironmentId] = useState("");
  const [connectionId, setConnectionId] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function load() {
    const [deploymentData, environmentData, connectionData, snapshotData] = await Promise.all([
      listDeployments(applicationId),
      listEnvironments(),
      listRealmConnections(),
      listSnapshots(applicationId),
    ]);
    setDeployments(deploymentData);
    setEnvironments(environmentData);
    setConnections(connectionData);
    setSnapshots(snapshotData);
  }

  useEffect(() => {
    Promise.all([
      listDeployments(applicationId),
      listEnvironments(),
      listRealmConnections(),
      listSnapshots(applicationId),
    ]).then(([deploymentData, environmentData, connectionData, snapshotData]) => {
      setDeployments(deploymentData);
      setEnvironments(environmentData);
      setConnections(connectionData);
      setSnapshots(snapshotData);
    }).catch(() => setError("Failed to load deployments")).finally(() => setLoading(false));
  }, [applicationId]);

  async function snapshot() {
    setBusy(true); setError(""); setSuccess("");
    try {
      await createSnapshot(applicationId);
      setSuccess("Immutable snapshot created.");
      await load();
    } catch (err) { setError(message(err)); }
    finally { setBusy(false); }
  }

  async function deployOrPromote() {
    const destination = environments.find((item) => item.id === environmentId);
    if (!destination || !connectionId) return;
    setBusy(true); setError(""); setSuccess("");
    try {
      const source = deployments
        .filter((item) => item.promotion_order < destination.promotion_order && item.status === "deployed")
        .sort((a, b) => b.promotion_order - a.promotion_order)[0];
      const payload = source ? {
        source_deployment_id: source.id,
        destination_environment_id: destination.id,
        realm_connection_id: connectionId,
      } : null;
      if (destination.protected) {
        if (!payload) throw new Error("Deploy through the preceding environment first.");
        await submitApproval(applicationId, "promote_application", payload, `Promote snapshot to ${destination.name}`);
        setSuccess("Protected promotion submitted for approval.");
      } else if (payload) {
        await promoteApplication(applicationId, payload);
        setSuccess(`Promoted to ${destination.name}.`);
      } else {
        await deployApplication(applicationId, {
          environment_id: destination.id,
          realm_connection_id: connectionId,
          snapshot_id: snapshots[0]?.id,
        });
        setSuccess(`Deployed to ${destination.name}.`);
      }
      await load();
    } catch (err) { setError(message(err)); }
    finally { setBusy(false); }
  }

  async function requestAction(action: "reconcile_drift" | "rotate_client_secret", deployment: Deployment) {
    setBusy(true); setError(""); setSuccess("");
    try {
      await submitApproval(applicationId, action, { deployment_id: deployment.id },
        action === "reconcile_drift" ? `Reconcile drift in ${deployment.environment}` : `Rotate secret in ${deployment.environment}`);
      setSuccess("Request submitted for independent approval.");
    } catch (err) { setError(message(err)); }
    finally { setBusy(false); }
  }

  if (loading) return <InlineLoading description="Loading deployments..." />;
  const availableConnections = connections.filter((item) => item.environment_id === environmentId && item.enabled);
  return <div className="phase-two-stack">
    {error && <InlineNotification kind="error" title="Deployment operation failed" subtitle={error} />}
    {success && <InlineNotification kind="success" title="Deployment workflow" subtitle={success} />}
    <Tile>
      <h4>Immutable snapshots</h4>
      <p>{snapshots.length ? `Latest version: ${snapshots[0].version} (${snapshots[0].configuration_hash.slice(0, 12)})` : "No snapshot yet."}</p>
      {can("manage_drafts") && <Button size="sm" kind="secondary" disabled={busy} onClick={() => void snapshot()}>Create snapshot</Button>}
    </Tile>
    {deployments.map((deployment) => <Tile key={deployment.id}>
      <h4>{deployment.environment} · {deployment.realm}</h4>
      <p><Tag type={deployment.status === "deployed" ? "green" : "red"}>{deployment.status}</Tag>{" "}
        <Tag type={deployment.drift_status === "in_sync" ? "green" : deployment.drift_status === "drifted" ? "red" : "gray"}>{deployment.drift_status}</Tag></p>
      <p>Client UUID: <code>{deployment.keycloak_client_uuid || "Not deployed"}</code></p>
      {can("submit_approval") && <div className="detail-actions">
        {deployment.drift_status === "drifted" && <Button size="sm" kind="secondary" disabled={busy} onClick={() => void requestAction("reconcile_drift", deployment)}>Request reconciliation</Button>}
        <Button size="sm" kind="danger--tertiary" disabled={busy} onClick={() => void requestAction("rotate_client_secret", deployment)}>Request secret rotation</Button>
      </div>}
    </Tile>)}
    {can("promote_applications") && <Tile>
      <h4>Deploy or promote</h4>
      <div className="phase-two-stack">
        <Select id="deployment-environment" labelText="Destination environment" value={environmentId}
          onChange={(event) => { setEnvironmentId(event.target.value); setConnectionId(""); }}>
          <SelectItem value="" text="Choose an environment" />
          {environments.map((item) => <SelectItem key={item.id} value={item.id} text={`${item.name}${item.protected ? " (approval required)" : ""}`} />)}
        </Select>
        <Select id="deployment-connection" labelText="Realm connection" value={connectionId}
          onChange={(event) => setConnectionId(event.target.value)}>
          <SelectItem value="" text="Choose a realm" />
          {availableConnections.map((item) => <SelectItem key={item.id} value={item.id} text={`${item.name} · ${item.realm}`} />)}
        </Select>
        <Button disabled={busy || !environmentId || !connectionId} onClick={() => void deployOrPromote()}>
          {busy ? "Working..." : "Continue"}
        </Button>
      </div>
    </Tile>}
  </div>;
}

function message(error: unknown) {
  return axios.isAxiosError(error)
    ? (error.response?.data?.error ?? error.message)
    : error instanceof Error ? error.message : "Operation failed";
}
