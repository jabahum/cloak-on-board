import { api, type ApiResponse } from "./client";
import type {
  ApprovalAction,
  ApprovalRequest,
} from "../types/approval";

export async function listApprovals() {
  const response =
    await api.get<ApiResponse<ApprovalRequest[]>>("/approval-requests");
  return response.data.data;
}
export async function getApproval(id: string) {
  const response = await api.get<ApiResponse<ApprovalRequest>>(
    `/approval-requests/${id}`,
  );
  return response.data.data;
}
export async function submitApproval(
  applicationId: string,
  action: ApprovalAction,
  payload: unknown = {},
  summary = "",
) {
  const response = await api.post<ApiResponse<ApprovalRequest>>(
    `/applications/${applicationId}/approval-requests`,
    { action, payload, summary },
  );
  return response.data.data;
}
export async function decideApproval(
  id: string,
  decision: "approve" | "reject" | "cancel" | "retry",
  comment = "",
) {
  const response = await api.post<ApiResponse<ApprovalRequest>>(
    `/approval-requests/${id}/${decision}`,
    { comment },
  );
  return response.data.data;
}
