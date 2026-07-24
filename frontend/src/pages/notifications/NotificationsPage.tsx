import { useEffect, useState } from "react";
import { Button, InlineLoading, InlineNotification, Tile } from "@carbon/react";
import {
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from "../../api/notifications";
import type { Notification } from "../../types/notification";
import { Link } from "react-router-dom";

export function NotificationsPage() {
  const [items, setItems] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  async function load() {
    try { setItems(await listNotifications()) }
    catch { setError("Failed to load notifications") }
    finally { setLoading(false) }
  }
  useEffect(() => { listNotifications().then(setItems).catch(() => setError("Failed to load notifications")).finally(() => setLoading(false)) }, [])
  async function read(item: Notification) { await markNotificationRead(item.id); await load() }
  async function readAll() { await markAllNotificationsRead(); await load() }
  if (loading) return <InlineLoading description="Loading notifications..." />
  return <>
    <div className="page-header"><h1 className="page-title">Notifications</h1><p className="page-subtitle">Approval and execution updates.</p></div>
    {error && <InlineNotification kind="error" title="Notifications" subtitle={error} />}
    <Button kind="secondary" onClick={readAll}>Mark all read</Button><br /><br />
    <div className="phase-two-stack">
      {items.map(item => <Tile key={item.id} className={item.read_at ? "" : "notification-unread"}>
        <h4>{item.title}</h4><p>{item.message}</p><small>{new Date(item.created_at).toLocaleString()}</small>
        <div className="detail-actions">
          {!item.read_at && <Button size="sm" kind="ghost" onClick={() => read(item)}>Mark read</Button>}
          {item.application_id && <Button as={Link} size="sm" kind="ghost" to={`/applications/${item.application_id}`}>Open application</Button>}
        </div>
      </Tile>)}
      {items.length===0 && <Tile>No notifications.</Tile>}
    </div>
  </>
}
