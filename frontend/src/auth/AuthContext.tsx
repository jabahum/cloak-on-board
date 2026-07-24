/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { Button, InlineLoading, InlineNotification } from "@carbon/react";
import { api } from "../api/client";
import { initKeycloak, keycloak, logout } from "./keycloak";

export type Role = "admin" | "manager" | "viewer";
export type Permission =
  | "read"
  | "manage_drafts"
  | "submit_approval"
  | "admin_clients"
  | "manage_settings"
  | "review_approval"
  | "view_audit";
export type CurrentUser = {
  subject: string;
  username: string;
  email: string;
  display_name: string;
  realm_roles: string[];
  effective_role: Role;
};

const permissions: Record<Role, Permission[]> = {
  viewer: ["read"],
  manager: ["read", "manage_drafts", "submit_approval"],
  admin: [
    "read",
    "manage_drafts",
    "submit_approval",
    "admin_clients",
    "manage_settings",
    "review_approval",
    "view_audit",
  ],
};

type AuthValue = {
  user: CurrentUser;
  can: (permission: Permission) => boolean;
  logout: () => void;
};
const AuthContext = createContext<AuthValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    initKeycloak()
      .then(() => api.get<{ data: CurrentUser }>("/auth/me"))
      .then((response) => setUser(response.data.data))
      .catch((err) =>
        setError(err instanceof Error ? err.message : "Authentication failed"),
      )
      .finally(() => setLoading(false));
  }, []);

  const value = useMemo<AuthValue | null>(
    () =>
      user
        ? {
            user,
            can: (permission) =>
              permissions[user.effective_role].includes(permission),
            logout: () => void logout(),
          }
        : null,
    [user],
  );

  if (loading) {
    return <InlineLoading description="Signing in..." />;
  }
  if (error || !value) {
    return (
      <div className="auth-error">
        <InlineNotification
          kind="error"
          title="Unable to sign in"
          subtitle={error || "No supported cloak-on-board role was found."}
        />
        <Button onClick={() => void keycloak.login()}>Try again</Button>
      </div>
    );
  }
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("useAuth must be used inside AuthProvider");
  return value;
}

export function RequirePermission({
  permission,
  children,
}: {
  permission: Permission;
  children: ReactNode;
}) {
  const { can } = useAuth();
  if (!can(permission)) {
    return (
      <InlineNotification
        kind="error"
        title="Access denied"
        subtitle="You do not have permission to view this page."
      />
    );
  }
  return children;
}
