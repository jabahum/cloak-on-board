import { useEffect, useState } from "react";
import {
  Button,
  Checkbox,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  TableToolbar,
  TableToolbarContent,
  Tag,
  InlineLoading,
  InlineNotification,
  Modal,
  TextInput,
} from "@carbon/react";
import { Add, ImportExport } from "@carbon/icons-react";
import { Link } from "react-router-dom";
import {
  deleteApplication,
  listApplications,
} from "../../api/applications";
import type { Application } from "../../types/application";
import axios from "axios";

const headers = [
  { key: "name", header: "Name" },
  { key: "slug", header: "Client ID" },
  { key: "app_type", header: "Type" },
  { key: "status", header: "Status" },
  { key: "owner_name", header: "Owner" },
  { key: "actions", header: "Actions" },
];

export function ApplicationsPage() {
  const [apps, setApps] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState<Application | null>(null);
  const [deleteFromKeycloak, setDeleteFromKeycloak] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    listApplications()
      .then(setApps)
      .finally(() => setLoading(false));
  }, []);

  const rows = apps.map((app) => ({
    id: app.id,
    name: app.name,
    slug: app.slug,
    app_type: app.app_type,
    status: app.status,
    owner_name: app.owner_name || "-",
    actions: app.id,
  }));

  async function confirmDelete() {
    if (!deleteTarget) return;
    setDeleting(true);
    setError("");
    try {
      await deleteApplication(deleteTarget.id, deleteFromKeycloak);
      setApps((current) =>
        current.filter((app) => app.id !== deleteTarget.id),
      );
      setDeleteTarget(null);
      setDeleteFromKeycloak(false);
      setConfirmation("");
    } catch (err) {
      setError(
        axios.isAxiosError(err)
          ? (err.response?.data?.error ?? err.message)
          : err instanceof Error
            ? err.message
            : "Delete failed",
      );
      setDeleteTarget(null);
    } finally {
      setDeleting(false);
    }
  }

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Applications</h1>
        <p className="page-subtitle">
          Register and provision applications into Keycloak.
        </p>
      </div>

      {error && (
        <InlineNotification kind="error" title="Delete failed" subtitle={error} />
      )}

      {loading ? (
        <InlineLoading description="Loading applications..." />
      ) : (
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
            <TableContainer>
              <TableToolbar>
                <TableToolbarContent>
                  <Button
                    as={Link}
                    kind="secondary"
                    to="/applications/import"
                    renderIcon={ImportExport}
                  >
                    Import from Keycloak
                  </Button>
                  <Button as={Link} to="/applications/new" renderIcon={Add}>
                    New application
                  </Button>
                </TableToolbarContent>
              </TableToolbar>

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
                                cell.value === "provisioned" ? "green" : "gray"
                              }
                            >
                              {cell.value}
                            </Tag>
                          ) : cell.info.header === "actions" ? (
                            <div className="table-actions">
                              <Button
                                as={Link}
                                size="sm"
                                kind="ghost"
                                to={`/applications/${cell.value}`}
                              >
                                View
                              </Button>
                              <Button
                                as={Link}
                                size="sm"
                                kind="ghost"
                                to={`/applications/${cell.value}/edit`}
                              >
                                Edit
                              </Button>
                              <Button
                                size="sm"
                                kind="danger--ghost"
                                onClick={() =>
                                  setDeleteTarget(
                                    apps.find(
                                      (app) => app.id === cell.value,
                                    ) ?? null,
                                  )
                                }
                              >
                                Delete
                              </Button>
                            </div>
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
      )}

      <Modal
        danger
        open={Boolean(deleteTarget)}
        modalHeading="Delete application?"
        primaryButtonText={deleting ? "Deleting..." : "Delete"}
        secondaryButtonText="Cancel"
        primaryButtonDisabled={
          deleting ||
          (deleteFromKeycloak && confirmation !== deleteTarget?.slug)
        }
        onRequestSubmit={confirmDelete}
        onRequestClose={() => setDeleteTarget(null)}
      >
        <div className="phase-two-stack">
          <p>Local-only deletion leaves the Keycloak client untouched.</p>
          <Checkbox
            id="list-delete-from-keycloak"
            labelText="Also permanently delete the linked Keycloak client"
            checked={deleteFromKeycloak}
            disabled={!deleteTarget?.keycloak_client_uuid}
            onChange={(_, data) =>
              setDeleteFromKeycloak(Boolean(data.checked))
            }
          />
          {deleteFromKeycloak && (
            <TextInput
              id="list-delete-confirmation"
              labelText={`Type ${deleteTarget?.slug} to confirm`}
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
            />
          )}
        </div>
      </Modal>
    </>
  );
}
