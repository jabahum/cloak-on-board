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

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/applications" element={<ApplicationsPage />} />
        <Route path="/applications/new" element={<ApplicationWizardPage />} />
        <Route path="/applications/import" element={<ImportApplicationPage />} />
        <Route path="/applications/:id/edit" element={<EditApplicationPage />} />
        <Route path="/applications/:id" element={<ApplicationDetailPage />} />
        <Route path="/templates" element={<TemplatesPage />} />
        <Route path="/jobs" element={<ProvisioningJobsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Route>
    </Routes>
  );
}
