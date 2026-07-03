import { useEffect, useState } from "react";
import {
  Button,
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
} from "@carbon/react";
import { Add } from "@carbon/icons-react";
import { Link } from "react-router-dom";
import { listApplications } from "../../api/applications";
import type { Application } from "../../types/application";

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

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Applications</h1>
        <p className="page-subtitle">
          Register and provision applications into Keycloak.
        </p>
      </div>

      {loading ? (
        <InlineLoading description="Loading applications..." />
      ) : (
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
            <TableContainer>
              <TableToolbar>
                <TableToolbarContent>
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
                            <Button
                              as={Link}
                              size="sm"
                              kind="ghost"
                              to={`/applications/${cell.value}`}
                            >
                              View
                            </Button>
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
    </>
  );
}
