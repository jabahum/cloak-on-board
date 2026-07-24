import { useEffect, useState } from "react";
import {
  Button,
  Checkbox,
  InlineLoading,
  InlineNotification,
  Modal,
  Select,
  SelectItem,
  TextInput,
  Tile,
} from "@carbon/react";
import axios from "axios";
import {
  createProtocolMapper,
  deleteProtocolMapper,
  listProtocolMappers,
  updateProtocolMapper,
} from "../../api/applications";
import type {
  ProtocolMapper,
  ProtocolMapperPayload,
} from "../../types/application";

type MapperForm = {
  id?: string;
  name: string;
  type: string;
  userAttribute: string;
  claimName: string;
  jsonType: string;
  idToken: boolean;
  accessToken: boolean;
  userinfo: boolean;
};

const emptyForm: MapperForm = {
  name: "",
  type: "oidc-usermodel-attribute-mapper",
  userAttribute: "",
  claimName: "",
  jsonType: "String",
  idToken: true,
  accessToken: true,
  userinfo: true,
};

export function ProtocolMappersTab({
  applicationId,
}: {
  applicationId: string;
}) {
  const [mappers, setMappers] = useState<ProtocolMapper[]>([]);
  const [form, setForm] = useState<MapperForm>(emptyForm);
  const [editorOpen, setEditorOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<ProtocolMapper | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      setMappers(await listProtocolMappers(applicationId));
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    listProtocolMappers(applicationId)
      .then(setMappers)
      .catch((err) => setError(errorMessage(err)))
      .finally(() => setLoading(false));
  }, [applicationId]);

  function edit(mapper: ProtocolMapper) {
    setForm({
      id: mapper.id,
      name: mapper.name,
      type: mapper.protocolMapper,
      userAttribute: mapper.config["user.attribute"] ?? "",
      claimName: mapper.config["claim.name"] ?? "",
      jsonType: mapper.config["jsonType.label"] ?? "String",
      idToken: mapper.config["id.token.claim"] === "true",
      accessToken: mapper.config["access.token.claim"] === "true",
      userinfo: mapper.config["userinfo.token.claim"] === "true",
    });
    setEditorOpen(true);
  }

  async function save() {
    setSaving(true);
    setError("");
    const payload = mapperPayload(form);
    try {
      if (form.id) {
        await updateProtocolMapper(applicationId, form.id, payload);
      } else {
        await createProtocolMapper(applicationId, payload);
      }
      setEditorOpen(false);
      setForm(emptyForm);
      await load();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSaving(false);
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return;
    setSaving(true);
    try {
      await deleteProtocolMapper(applicationId, deleteTarget.id);
      setDeleteTarget(null);
      await load();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSaving(false);
    }
  }

  if (loading)
    return <InlineLoading description="Loading protocol mappers..." />;

  return (
    <>
      {error && (
        <InlineNotification
          kind="error"
          title="Protocol mappers"
          subtitle={error}
        />
      )}
      <div className="detail-actions">
        <Button
          onClick={() => {
            setForm(emptyForm);
            setEditorOpen(true);
          }}
        >
          Add mapper
        </Button>
      </div>
      <div className="phase-two-stack">
        {mappers.length === 0 ? (
          <Tile>No protocol mappers configured.</Tile>
        ) : (
          mappers.map((mapper) => (
            <Tile key={mapper.id}>
              <h4>{mapper.name}</h4>
              <p>
                <strong>Type:</strong> {mapper.protocolMapper}
              </p>
              <p>
                <strong>Claim:</strong> {mapper.config["claim.name"] || "-"}
              </p>
              <p>
                <strong>JSON type:</strong>{" "}
                {mapper.config["jsonType.label"] || "-"}
              </p>
              <p>
                <strong>Tokens:</strong>{" "}
                {tokenDestinations(mapper.config).join(", ") || "none"}
              </p>
              <div className="detail-actions">
                <Button size="sm" kind="secondary" onClick={() => edit(mapper)}>
                  Edit
                </Button>
                <Button
                  size="sm"
                  kind="danger--tertiary"
                  onClick={() => setDeleteTarget(mapper)}
                >
                  Delete
                </Button>
              </div>
            </Tile>
          ))
        )}
      </div>

      <Modal
        open={editorOpen}
        modalHeading={form.id ? "Edit protocol mapper" : "Add protocol mapper"}
        primaryButtonText={saving ? "Saving..." : "Save"}
        secondaryButtonText="Cancel"
        primaryButtonDisabled={
          saving ||
          !form.name.trim() ||
          !form.claimName.trim() ||
          (form.type === "oidc-usermodel-attribute-mapper" &&
            !form.userAttribute.trim())
        }
        onRequestSubmit={save}
        onRequestClose={() => setEditorOpen(false)}
      >
        <div className="phase-two-stack">
          <TextInput
            id="mapper-name"
            labelText="Name"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
          />
          <Select
            id="mapper-type"
            labelText="Mapper type"
            value={form.type}
            onChange={(e) => setForm({ ...form, type: e.target.value })}
          >
            <SelectItem
              value="oidc-usermodel-attribute-mapper"
              text="User attribute"
            />
            <SelectItem
              value="oidc-usermodel-client-role-mapper"
              text="Client roles"
            />
            <SelectItem
              value="oidc-usermodel-realm-role-mapper"
              text="Realm roles"
            />
          </Select>
          {form.type === "oidc-usermodel-attribute-mapper" && (
            <TextInput
              id="mapper-user-attribute"
              labelText="User attribute"
              value={form.userAttribute}
              onChange={(e) =>
                setForm({ ...form, userAttribute: e.target.value })
              }
            />
          )}
          <TextInput
            id="mapper-claim-name"
            labelText="Claim name"
            value={form.claimName}
            onChange={(e) => setForm({ ...form, claimName: e.target.value })}
          />
          <Select
            id="mapper-json-type"
            labelText="JSON type"
            value={form.jsonType}
            onChange={(e) => setForm({ ...form, jsonType: e.target.value })}
          >
            {["String", "long", "int", "boolean", "JSON"].map((value) => (
              <SelectItem key={value} value={value} text={value} />
            ))}
          </Select>
          <Checkbox
            id="mapper-id-token"
            labelText="Include in ID token"
            checked={form.idToken}
            onChange={(_, data) =>
              setForm({ ...form, idToken: Boolean(data.checked) })
            }
          />
          <Checkbox
            id="mapper-access-token"
            labelText="Include in access token"
            checked={form.accessToken}
            onChange={(_, data) =>
              setForm({ ...form, accessToken: Boolean(data.checked) })
            }
          />
          <Checkbox
            id="mapper-userinfo"
            labelText="Include in userinfo"
            checked={form.userinfo}
            onChange={(_, data) =>
              setForm({ ...form, userinfo: Boolean(data.checked) })
            }
          />
        </div>
      </Modal>

      <Modal
        danger
        open={Boolean(deleteTarget)}
        modalHeading="Delete protocol mapper?"
        primaryButtonText={saving ? "Deleting..." : "Delete"}
        secondaryButtonText="Cancel"
        primaryButtonDisabled={saving}
        onRequestSubmit={confirmDelete}
        onRequestClose={() => setDeleteTarget(null)}
      >
        This permanently deletes “{deleteTarget?.name}” from Keycloak.
      </Modal>
    </>
  );
}

function mapperPayload(form: MapperForm): ProtocolMapperPayload {
  return {
    name: form.name.trim(),
    protocol: "openid-connect",
    protocolMapper: form.type,
    config: {
      "claim.name": form.claimName.trim(),
      "jsonType.label": form.jsonType,
      "id.token.claim": String(form.idToken),
      "access.token.claim": String(form.accessToken),
      "userinfo.token.claim": String(form.userinfo),
      multivalued: String(form.type !== "oidc-usermodel-attribute-mapper"),
      ...(form.type === "oidc-usermodel-attribute-mapper"
        ? { "user.attribute": form.userAttribute.trim() }
        : {}),
    },
  };
}

function tokenDestinations(config: Record<string, string>) {
  return [
    config["id.token.claim"] === "true" ? "ID token" : "",
    config["access.token.claim"] === "true" ? "access token" : "",
    config["userinfo.token.claim"] === "true" ? "userinfo" : "",
  ].filter(Boolean);
}

function errorMessage(err: unknown) {
  if (axios.isAxiosError(err)) {
    return err.response?.data?.error ?? err.message;
  }
  return err instanceof Error ? err.message : "An unexpected error occurred";
}
