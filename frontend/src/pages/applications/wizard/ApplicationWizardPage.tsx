import { useEffect, useMemo, useState } from "react";
import {
  Button,
  InlineNotification,
  ProgressIndicator,
  ProgressStep,
} from "@carbon/react";
import { useNavigate } from "react-router-dom";
import {
  createApplication,
  provisionApplication,
} from "../../../api/applications";
import { listTemplates } from "../../../api/templates";
import type { OnboardingTemplate } from "../../../types/template";
import { GeneralStep } from "./steps/GeneralStep";
import { ReviewStep } from "./steps/ReviewStep";
import { RolesStep } from "./steps/RolesStep";
import { TemplateStep } from "./steps/TemplateStep";
import { UrlsStep } from "./steps/UrlsStep";
import { initialWizardForm, type WizardForm } from "./wizardTypes";
import axios from "axios";
import { useAuth } from "../../../auth/AuthContext";
import { submitApproval } from "../../../api/approvals";

const steps = ["Template", "General", "URLs", "Roles", "Review"];

export function ApplicationWizardPage() {
  const navigate = useNavigate();
  const { can } = useAuth();

  const [currentStep, setCurrentStep] = useState(0);
  const [templates, setTemplates] = useState<OnboardingTemplate[]>([]);
  const [loadingTemplates, setLoadingTemplates] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const [form, setForm] = useState<WizardForm>(initialWizardForm);
  const [error, setError] = useState("");

  useEffect(() => {
    listTemplates()
      .then((items) => {
        setTemplates(items);

        if (items.length > 0) {
          applyTemplate(items[0]);
        }
      })
      .catch(() => setError("Failed to load templates"))
      .finally(() => setLoadingTemplates(false));
  }, []);

  const isLastStep = currentStep === steps.length - 1;
  const isFirstStep = currentStep === 0;

  const canContinue = useMemo(() => {
    if (currentStep === 0) return Boolean(form.template);

    if (currentStep === 1) {
      return Boolean(
        form.name.trim() && form.slug.trim() && form.app_type.trim(),
      );
    }

    if (currentStep === 2) {
      if (form.app_type === "frontend" || form.app_type === "mobile") {
        return form.redirect_uris.length > 0;
      }

      return true;
    }

    if (currentStep === 3) {
      return form.roles.length > 0;
    }

    return true;
  }, [currentStep, form]);

  function applyTemplate(template: OnboardingTemplate) {
    setForm((current) => ({
      ...current,
      template,
      app_type: template.app_type,
      roles: template.default_roles?.length
        ? template.default_roles
        : current.roles,
      redirect_uris: template.default_redirect_uris?.length
        ? template.default_redirect_uris
        : current.redirect_uris,
      web_origins: template.default_web_origins?.length
        ? template.default_web_origins
        : current.web_origins,
    }));
  }

  function next() {
    if (!canContinue) return;
    setCurrentStep((value) => Math.min(value + 1, steps.length - 1));
  }

  function previous() {
    setCurrentStep((value) => Math.max(value - 1, 0));
  }

  async function submit() {
    setError("");
    setSubmitting(true);

    try {
      const app = await createApplication({
        name: form.name,
        slug: form.slug,
        description: form.description,
        app_type: form.app_type,
        owner_name: form.owner_name,
        owner_email: form.owner_email,
        redirect_uris: form.redirect_uris,
        web_origins: form.web_origins,
        roles: form.roles,
      });

      if (form.auto_provision && can("admin_clients")) {
        await provisionApplication(app.id);
      } else if (form.auto_provision) {
        await submitApproval(app.id, "provision_application", {}, "Provision new application");
      }

      navigate(`/applications/${app.id}`);
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error ?? err.message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to create application");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <div className="page-header">
        <h1 className="page-title">Onboard Application</h1>
        <p className="page-subtitle">
          Use the wizard to create an application configuration for Keycloak.
        </p>
      </div>

      {error && (
        <InlineNotification kind="error" title="Error" subtitle={error} />
      )}

      <ProgressIndicator currentIndex={currentStep} spaceEqually>
        {steps.map((step) => (
          <ProgressStep key={step} label={step} />
        ))}
      </ProgressIndicator>

      <br />

      {currentStep === 0 && (
        <TemplateStep
          templates={templates}
          loading={loadingTemplates}
          selectedTemplateId={form.template?.id}
          onSelect={applyTemplate}
        />
      )}

      {currentStep === 1 && <GeneralStep form={form} setForm={setForm} />}
      {currentStep === 2 && <UrlsStep form={form} setForm={setForm} />}
      {currentStep === 3 && <RolesStep form={form} setForm={setForm} />}
      {currentStep === 4 && <ReviewStep form={form} setForm={setForm} />}
      <br />

      <div className="wizard-actions">
        <Button kind="secondary" onClick={previous} disabled={isFirstStep}>
          Back
        </Button>

        {!isLastStep ? (
          <Button onClick={next} disabled={!canContinue}>
            Next
          </Button>
        ) : (
          <Button onClick={submit} disabled={submitting}>
            {submitting
              ? form.auto_provision
                ? "Creating and provisioning..."
                : "Creating..."
              : form.auto_provision
                ? "Create and provision"
                : "Create application"}
          </Button>
        )}
      </div>
    </>
  );
}
