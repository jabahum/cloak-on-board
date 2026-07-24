import { useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Modal,
  Tag,
  TextArea,
  Tile,
} from "@carbon/react";
import axios from "axios";
import { decideApproval, listApprovals } from "../../api/approvals";
import type { ApprovalRequest } from "../../types/approval";
import { useAuth } from "../../auth/AuthContext";

export function ApprovalsPage() {
  const { user, can } = useAuth();
  const [items, setItems] = useState<ApprovalRequest[]>([]);
  const [selected, setSelected] = useState<ApprovalRequest | null>(null);
  const [decision, setDecision] = useState<
    "approve" | "reject" | "cancel" | "retry" | null
  >(null);
  const [comment, setComment] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    try {
      setItems(await listApprovals());
    } catch (err) {
      setError(message(err));
    } finally {
      setLoading(false);
    }
  }
  useEffect(() => {
    listApprovals()
      .then(setItems)
      .catch((err) => setError(message(err)))
      .finally(() => setLoading(false));
  }, []);

  async function submitDecision() {
    if (!selected || !decision) return;
    setSaving(true);
    try {
      await decideApproval(selected.id, decision, comment);
      setDecision(null);
      setSelected(null);
      setComment("");
      await load();
    } catch (err) {
      setError(message(err));
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <InlineLoading description="Loading approvals..." />;
  return (
    <>
      <div className="page-header">
        <h1 className="page-title">
          {can("review_approval") ? "Approval inbox" : "My approval requests"}
        </h1>
        <p className="page-subtitle">
          Review and track proposed Keycloak changes.
        </p>
      </div>
      {error && (
        <InlineNotification kind="error" title="Approvals" subtitle={error} />
      )}
      <div className="phase-two-stack">
        {items.map((item) => (
          <Tile key={item.id}>
            <h4>{item.application_name || "Deleted application"}</h4>
            <p>
              <strong>Action:</strong> {item.action}
            </p>
            <p>
              <strong>Requested by:</strong> {item.requested_by_username}
            </p>
            <Tag type={statusColor(item.status)}>{item.status}</Tag>
            <div className="detail-actions">
              <Button size="sm" kind="secondary" onClick={() => setSelected(item)}>
                Review
              </Button>
            </div>
          </Tile>
        ))}
        {items.length === 0 && <Tile>No approval requests found.</Tile>}
      </div>

      <Modal
        open={Boolean(selected) && !decision}
        modalHeading="Approval request"
        passiveModal
        onRequestClose={() => setSelected(null)}
      >
        {selected && (
          <div className="phase-two-stack">
            <p><strong>Action:</strong> {selected.action}</p>
            <p><strong>Summary:</strong> {selected.request_summary || "-"}</p>
            <p><strong>Application version:</strong> {selected.application_version}</p>
            <pre className="json-preview">{JSON.stringify(selected.request_payload, null, 2)}</pre>
            {selected.execution_error && (
              <InlineNotification kind="error" title="Execution failed" subtitle={selected.execution_error} />
            )}
            <div className="detail-actions">
              {can("review_approval") && selected.status === "pending" && selected.requested_by_subject !== user.subject && (
                <>
                  <Button size="sm" onClick={() => setDecision("approve")}>Approve</Button>
                  <Button size="sm" kind="danger--tertiary" onClick={() => setDecision("reject")}>Reject</Button>
                </>
              )}
              {selected.status === "pending" &&
                (selected.requested_by_subject === user.subject || can("review_approval")) && (
                  <Button size="sm" kind="secondary" onClick={() => setDecision("cancel")}>Cancel request</Button>
                )}
              {can("review_approval") && selected.status === "failed" && (
                <Button size="sm" onClick={() => setDecision("retry")}>Retry</Button>
              )}
            </div>
          </div>
        )}
      </Modal>
      <Modal
        danger={decision === "reject"}
        open={Boolean(decision)}
        modalHeading={`${decision ?? ""} approval request?`}
        primaryButtonText={saving ? "Working..." : "Confirm"}
        secondaryButtonText="Back"
        primaryButtonDisabled={saving || (decision === "reject" && !comment.trim())}
        onRequestSubmit={submitDecision}
        onRequestClose={() => setDecision(null)}
      >
        <TextArea
          id="approval-comment"
          labelText={decision === "reject" ? "Reason (required)" : "Comment"}
          value={comment}
          onChange={(event) => setComment(event.target.value)}
        />
      </Modal>
    </>
  );
}

function statusColor(status: string) {
  if (status === "succeeded" || status === "approved") return "green";
  if (status === "failed" || status === "rejected") return "red";
  return "blue";
}
function message(err: unknown) {
  return axios.isAxiosError(err)
    ? (err.response?.data?.error ?? err.message)
    : err instanceof Error ? err.message : "Unexpected error";
}
