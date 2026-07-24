import { useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Tag,
  Tile,
} from "@carbon/react";
import axios from "axios";
import {
  assignClientScope,
  getClientScopes,
  removeClientScope,
} from "../../api/applications";
import type {
  ClientScope,
  ClientScopeAssignments,
} from "../../types/application";

export function ClientScopesTab({ applicationId, canManage }: { applicationId: string; canManage: boolean }) {
  const [scopes, setScopes] = useState<ClientScopeAssignments | null>(null);
  const [selected, setSelected] = useState("");
  const [scopeType, setScopeType] = useState<"default" | "optional">("default");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      setScopes(await getClientScopes(applicationId));
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    getClientScopes(applicationId)
      .then(setScopes)
      .catch((err) => setError(errorMessage(err)))
      .finally(() => setLoading(false));
  }, [applicationId]);

  async function assign() {
    if (!selected) return;
    setSaving(true);
    setError("");
    try {
      await assignClientScope(applicationId, selected, scopeType);
      setSelected("");
      await load();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSaving(false);
    }
  }

  async function remove(scope: ClientScope, type: "default" | "optional") {
    setSaving(true);
    setError("");
    try {
      await removeClientScope(applicationId, scope.id, type);
      await load();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <InlineLoading description="Loading client scopes..." />;

  return (
    <div className="phase-two-stack">
      {error && (
        <InlineNotification kind="error" title="Client scopes" subtitle={error} />
      )}

      {canManage && <Tile>
        <h4>Assign client scope</h4>
        <div className="inline-form">
          <Select
            id="available-client-scope"
            labelText="Available scope"
            value={selected}
            onChange={(event) => setSelected(event.target.value)}
          >
            <SelectItem value="" text="Select a scope" />
            {(scopes?.available ?? []).map((scope) => (
              <SelectItem key={scope.id} value={scope.id} text={scope.name} />
            ))}
          </Select>
          <Select
            id="client-scope-type"
            labelText="Assignment type"
            value={scopeType}
            onChange={(event) =>
              setScopeType(event.target.value as "default" | "optional")
            }
          >
            <SelectItem value="default" text="Default" />
            <SelectItem value="optional" text="Optional" />
          </Select>
          <Button disabled={!selected || saving} onClick={assign}>
            Assign
          </Button>
        </div>
        {(scopes?.available ?? []).length === 0 && (
          <p>All realm client scopes are assigned.</p>
        )}
      </Tile>}

      <ScopeList
        title="Default scopes"
        scopes={scopes?.default ?? []}
        onRemove={(scope) => remove(scope, "default")}
        disabled={saving}
        canManage={canManage}
      />
      <ScopeList
        title="Optional scopes"
        scopes={scopes?.optional ?? []}
        onRemove={(scope) => remove(scope, "optional")}
        disabled={saving}
        canManage={canManage}
      />
    </div>
  );
}

function ScopeList({
  title,
  scopes,
  onRemove,
  disabled,
  canManage,
}: {
  title: string;
  scopes: ClientScope[];
  onRemove: (scope: ClientScope) => void;
  disabled: boolean;
  canManage: boolean;
}) {
  return (
    <Tile>
      <h4>{title}</h4>
      {scopes.length === 0 ? (
        <p>No scopes assigned.</p>
      ) : (
        <div className="tag-list">
          {scopes.map((scope) => (
            <Tag
              key={scope.id}
              filter={canManage}
              disabled={disabled}
              onClose={() => onRemove(scope)}
            >
              {scope.name}
            </Tag>
          ))}
        </div>
      )}
    </Tile>
  );
}

function errorMessage(err: unknown) {
  if (axios.isAxiosError(err)) {
    return err.response?.data?.error ?? err.message;
  }
  return err instanceof Error ? err.message : "An unexpected error occurred";
}
