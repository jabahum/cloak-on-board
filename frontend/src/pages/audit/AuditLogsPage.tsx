import { useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Modal,
  TextInput,
  Tile,
} from "@carbon/react";
import { listAuditLogs } from "../../api/audit";
import type { AuditLog } from "../../types/audit";

export function AuditLogsPage() {
  const [items, setItems] = useState<AuditLog[]>([]);
  const [actor, setActor] = useState("");
  const [action, setAction] = useState("");
  const [selected, setSelected] = useState<AuditLog | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  async function load() {
    setLoading(true);
    try {
      setItems(await listAuditLogs({ actor, action }));
    } catch {
      setError("Failed to load audit logs");
    } finally {
      setLoading(false);
    }
  }
  useEffect(() => {
    listAuditLogs()
      .then(setItems)
      .catch(() => setError("Failed to load audit logs"))
      .finally(() => setLoading(false));
  }, []);
  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Audit log</h1>
        <p className="page-subtitle">Append-only security and mutation history.</p>
      </div>
      {error && <InlineNotification kind="error" title="Audit log" subtitle={error} />}
      <div className="inline-form">
        <TextInput id="audit-actor" labelText="Actor" value={actor} onChange={(e) => setActor(e.target.value)} />
        <TextInput id="audit-action" labelText="Action" value={action} onChange={(e) => setAction(e.target.value)} />
        <Button onClick={load}>Filter</Button>
      </div>
      <br />
      {loading ? <InlineLoading description="Loading audit logs..." /> : (
        <div className="phase-two-stack">
          {items.map((item) => (
            <Tile key={item.id}>
              <p><strong>{item.action}</strong> — {item.result}</p>
              <p>{item.actor_username || "system"} ({item.actor_role || "-"})</p>
              <p>{new Date(item.occurred_at).toLocaleString()} · Request {item.request_id}</p>
              <Button size="sm" kind="ghost" onClick={() => setSelected(item)}>Details</Button>
            </Tile>
          ))}
          {items.length === 0 && <Tile>No audit entries found.</Tile>}
        </div>
      )}
      <Modal open={Boolean(selected)} passiveModal modalHeading="Audit details" onRequestClose={() => setSelected(null)}>
        <pre className="json-preview">{JSON.stringify(selected, null, 2)}</pre>
      </Modal>
    </>
  );
}
