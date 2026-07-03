import { useEffect, useState } from "react";
import {
  DataTable,
  InlineLoading,
  Tag,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";
import { listProvisioningJobs } from "../../api/provisioning";
import type { ProvisioningJob } from "../../types/provisioning";

const headers = [
  { key: "action", header: "Action" },
  { key: "application_id", header: "Application ID" },
  { key: "status", header: "Status" },
  { key: "error_message", header: "Error" },
  { key: "created_at", header: "Created" },
];

export function ProvisioningJobsPage() {
  const [jobs, setJobs] = useState<ProvisioningJob[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listProvisioningJobs()
      .then(setJobs)
      .finally(() => setLoading(false));
  }, []);

  const rows = jobs.map((job) => ({
    id: job.id,
    action: job.action,
    application_id: job.application_id,
    status: job.status,
    error_message: job.error_message || "-",
    created_at: new Date(job.created_at).toLocaleString(),
  }));

  if (loading) return <InlineLoading description="Loading jobs..." />;

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Provisioning Jobs</h1>
        <p className="page-subtitle">Track Keycloak provisioning progress.</p>
      </div>

      <DataTable rows={rows} headers={headers}>
        {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
          <TableContainer>
            <Table {...getTableProps()}>
              <TableHead>
                <TableRow>
                  {headers.map((header) => (
                    <TableHeader
                      {...getHeaderProps({ header })}
                      key={header.key}
                    >
                      {header.header}
                    </TableHeader>
                  ))}
                </TableRow>
              </TableHead>

              <TableBody>
                {rows.map((row) => (
                  <TableRow {...getRowProps({ row })} key={row.id}>
                    {row.cells.map((cell) => (
                      <TableCell key={cell.id}>
                        {cell.info.header === "status" ? (
                          <Tag
                            type={
                              cell.value === "succeeded"
                                ? "green"
                                : cell.value === "failed"
                                  ? "red"
                                  : "blue"
                            }
                          >
                            {cell.value}
                          </Tag>
                        ) : (
                          cell.value
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </DataTable>
    </>
  );
}
