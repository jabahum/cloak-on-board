import { Navigate, Route, Routes } from "react-router-dom";
import { AppLayout } from "./layouts/AppLayout";
import { ApplicationWizardPage } from "./pages/applications/wizard/ApplicationWizardPage";
import { ApplicationDetailPage } from "./pages/applications/ApplicationDetailPage";
import { ApplicationsPage } from "./pages/applications/ApplicationsPage";
import { EditApplicationPage } from "./pages/applications/EditApplicationPage";
import { ImportApplicationPage } from "./pages/applications/ImportApplicationPage";
import { DashboardPage } from "./pages/dashboard/DashboardPage";
import { ProvisioningJobsPage } from "./pages/jobs/ProvisioningJobsPage";
import { SettingsPage } from "./pages/settings/SettingsPage";
import { TemplatesPage } from "./pages/templates/TemplatesPage";
import { ApprovalsPage } from "./pages/approvals/ApprovalsPage";
import { AuditLogsPage } from "./pages/audit/AuditLogsPage";
import { NotificationsPage } from "./pages/notifications/NotificationsPage";
import { RequirePermission } from "./auth/AuthContext";

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/applications" element={<ApplicationsPage />} />
        <Route path="/applications/new" element={<RequirePermission permission="manage_drafts"><ApplicationWizardPage /></RequirePermission>} />
        <Route path="/applications/import" element={<RequirePermission permission="manage_drafts"><ImportApplicationPage /></RequirePermission>} />
        <Route path="/applications/:id/edit" element={<RequirePermission permission="manage_drafts"><EditApplicationPage /></RequirePermission>} />
        <Route path="/applications/:id" element={<ApplicationDetailPage />} />
        <Route path="/templates" element={<TemplatesPage />} />
        <Route path="/jobs" element={<ProvisioningJobsPage />} />
        <Route path="/approvals" element={<ApprovalsPage />} />
        <Route path="/notifications" element={<NotificationsPage />} />
        <Route path="/audit-logs" element={<RequirePermission permission="view_audit"><AuditLogsPage /></RequirePermission>} />
        <Route path="/settings" element={<RequirePermission permission="manage_settings"><SettingsPage /></RequirePermission>} />
      </Route>
    </Routes>
  );
}
