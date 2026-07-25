import { useEffect, useState } from "react";
import {
  Button,
  DataTable,
  InlineLoading,
  InlineNotification,
  Tag,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";
import axios from "axios";
import { checkDrift, listDeployments } from "../../api/delivery";
import type { Deployment } from "../../types/delivery";
import { useAuth } from "../../auth/AuthContext";

const headers = [
  { key: "application", header: "Application" },
  { key: "environment", header: "Environment" },
  { key: "realm", header: "Realm" },
  { key: "snapshot", header: "Snapshot" },
  { key: "status", header: "Deployment" },
  { key: "drift", header: "Drift" },
  { key: "actions", header: "Actions" },
];

export function DeploymentsPage() {
  const { can } = useAuth();
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    setDeployments(await listDeployments());
  }
  useEffect(() => {
    listDeployments()
      .then(setDeployments)
      .catch(() => setError("Failed to load deployment matrix"))
      .finally(() => setLoading(false));
  }, []);

  async function runDrift(id: string) {
    setError("");
    try {
      await checkDrift(id);
      await load();
    } catch (err) {
      setError(axios.isAxiosError(err) ? (err.response?.data?.error ?? err.message) : "Drift check failed");
    }
  }

  if (loading) return <InlineLoading description="Loading deployments..." />;
  const rows = deployments.map((deployment) => ({
    id: deployment.id,
    application: deployment.application_name,
    environment: deployment.environment,
    realm: deployment.realm,
    snapshot: deployment.snapshot_id.slice(0, 8),
    status: deployment.status,
    drift: deployment.drift_status,
    actions: deployment.id,
  }));
  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Deployment matrix</h1>
        <p className="page-subtitle">Immutable snapshots promoted across managed realms.</p>
      </div>
      {error && <InlineNotification kind="error" title="Deployment operation failed" subtitle={error} />}
      <DataTable rows={rows} headers={headers}>
        {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
          <TableContainer>
            <Table {...getTableProps()}>
              <TableHead><TableRow>{headers.map((header) => (
                <TableHeader {...getHeaderProps({ header })} key={header.key}>{header.header}</TableHeader>
              ))}</TableRow></TableHead>
              <TableBody>{rows.map((row) => (
                <TableRow {...getRowProps({ row })} key={row.id}>
                  {row.cells.map((cell) => (
                    <TableCell key={cell.id}>
                      {cell.info.header === "status" || cell.info.header === "drift" ? (
                        <Tag type={cell.value === "in_sync" || cell.value === "deployed" ? "green" : cell.value === "drifted" || cell.value === "failed" ? "red" : "gray"}>{cell.value}</Tag>
                      ) : cell.info.header === "actions" ? (
                        can("check_drift") && <Button size="sm" kind="ghost" onClick={() => void runDrift(String(cell.value))}>Check drift</Button>
                      ) : cell.value}
                    </TableCell>
                  ))}
                </TableRow>
              ))}</TableBody>
            </Table>
          </TableContainer>
        )}
      </DataTable>
    </>
  );
}
