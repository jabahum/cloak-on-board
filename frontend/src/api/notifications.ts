import { api, type ApiResponse } from "./client";
import type { Notification } from "../types/notification";
export async function listNotifications(){const r=await api.get<ApiResponse<Notification[]>>("/notifications");return r.data.data}
export async function unreadCount(){const r=await api.get<ApiResponse<{count:number}>>("/notifications/unread-count");return r.data.data.count}
export async function markNotificationRead(id:string){await api.put(`/notifications/${id}/read`)}
export async function markAllNotificationsRead(){await api.put("/notifications/read-all")}
