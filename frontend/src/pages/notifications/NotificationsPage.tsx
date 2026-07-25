import { useEffect, useState } from "react";
import { Button, CodeSnippet, InlineLoading, InlineNotification, Tile } from "@carbon/react";
import {
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from "../../api/notifications";
import type { Notification } from "../../types/notification";
import { Link } from "react-router-dom";
import { consumeSecretDelivery } from "../../api/delivery";

export function NotificationsPage() {
  const [items, setItems] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [revealedSecret, setRevealedSecret] = useState("");
  async function load() {
    try { setItems(await listNotifications()) }
    catch { setError("Failed to load notifications") }
    finally { setLoading(false) }
  }
  useEffect(() => { listNotifications().then(setItems).catch(() => setError("Failed to load notifications")).finally(() => setLoading(false)) }, [])
  async function read(item: Notification) { await markNotificationRead(item.id); await load() }
  async function readAll() { await markAllNotificationsRead(); await load() }
  async function reveal(item: Notification) {
    setError("");
    try {
      const delivery = await consumeSecretDelivery(item.resource_id);
      setRevealedSecret(delivery.secret);
      await markNotificationRead(item.id);
      await load();
    } catch {
      setError("This secret has expired, was already consumed, or is not assigned to you.");
    }
  }
  if (loading) return <InlineLoading description="Loading notifications..." />
  return <>
    <div className="page-header"><h1 className="page-title">Notifications</h1><p className="page-subtitle">Approval and execution updates.</p></div>
    {error && <InlineNotification kind="error" title="Notifications" subtitle={error} />}
    {revealedSecret && <>
      <InlineNotification
        kind="warning"
        title="Copy this secret now"
        subtitle="It has been consumed and cannot be retrieved again."
        onCloseButtonClick={() => setRevealedSecret("")}
      />
      <CodeSnippet type="multi">{revealedSecret}</CodeSnippet><br /><br />
    </>}
    <Button kind="secondary" onClick={readAll}>Mark all read</Button><br /><br />
    <div className="phase-two-stack">
      {items.map(item => <Tile key={item.id} className={item.read_at ? "" : "notification-unread"}>
        <h4>{item.title}</h4><p>{item.message}</p><small>{new Date(item.created_at).toLocaleString()}</small>
        <div className="detail-actions">
          {!item.read_at && <Button size="sm" kind="ghost" onClick={() => read(item)}>Mark read</Button>}
          {item.application_id && <Button as={Link} size="sm" kind="ghost" to={`/applications/${item.application_id}`}>Open application</Button>}
          {item.type === "secret_rotation_ready" && !item.read_at && (
            <Button size="sm" kind="danger--tertiary" onClick={() => void reveal(item)}>Reveal once</Button>
          )}
        </div>
      </Tile>)}
      {items.length===0 && <Tile>No notifications.</Tile>}
    </div>
  </>
}
