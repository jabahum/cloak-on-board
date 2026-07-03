import { useEffect, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Tag,
  Tile,
} from "@carbon/react";
import { listTemplates, seedTemplates } from "../../api/templates";
import type { OnboardingTemplate } from "../../types/template";

export function TemplatesPage() {
  const [templates, setTemplates] = useState<OnboardingTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    listTemplates()
      .then(setTemplates)
      .catch(() => setError("Failed to load templates"))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleSeed() {
    setError("");
    setMessage("");

    try {
      const res = await seedTemplates();
      setMessage(res.message);
      await load();
    } catch (err: any) {
      setError(err?.response?.data?.error ?? "Failed to seed templates");
    }
  }

  if (loading) return <InlineLoading description="Loading templates..." />;

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Templates</h1>
        <p className="page-subtitle">Default onboarding templates.</p>
      </div>

      {message && (
        <InlineNotification kind="success" title="Success" subtitle={message} />
      )}
      {error && (
        <InlineNotification kind="error" title="Error" subtitle={error} />
      )}

      <Button onClick={handleSeed}>Seed default templates</Button>

      <br />
      <br />

      <div className="template-grid">
        {templates.map((template) => (
          <Tile key={template.id}>
            <h4>{template.name}</h4>
            <p>{template.description}</p>
            <Tag>{template.app_type}</Tag>

            <h5>Roles</h5>
            {(template.default_roles ?? []).map((role) => (
              <Tag key={role}>{role}</Tag>
            ))}
          </Tile>
        ))}
      </div>
    </>
  );
}
