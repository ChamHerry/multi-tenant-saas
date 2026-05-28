import { createBrowserRouter } from "react-router-dom";
import { RequireAuth } from "./RequireAuth";
import { RequirePermission } from "@/features/access/RequirePermission";
import { DashboardPage } from "@/pages/DashboardPage";
import { LoginPage } from "@/pages/LoginPage";
import { RegisterPage } from "@/pages/RegisterPage";
import { TenantsPage } from "@/pages/TenantsPage";
import { TenantDetailPage } from "@/pages/TenantDetailPage";
import { MembersPage } from "@/pages/MembersPage";
import { PersonalApiKeysPage } from "@/pages/PersonalApiKeysPage";
import { TenantInvitationsPage } from "@/pages/TenantInvitationsPage";
import { MyInvitationsPage } from "@/pages/MyInvitationsPage";
import { TenantUsagePage } from "@/pages/TenantUsagePage";
import { TenantAuditPage } from "@/pages/TenantAuditPage";
import { SecurityEventsPage } from "@/pages/SecurityEventsPage";
import { AdminTenantsPage } from "@/pages/AdminTenantsPage";
import { AdminUsersPage } from "@/pages/AdminUsersPage";
import { AdminAuditPage } from "@/pages/AdminAuditPage";
import { AdminPlansPage } from "@/pages/AdminPlansPage";
import { NotFoundPage } from "@/pages/NotFoundPage";

export const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/register",
    element: <RegisterPage />,
  },
  {
    path: "/",
    element: <RequireAuth />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: "tenants", element: <TenantsPage /> },
      {
        path: "tenants/:tenantId",
        element: (
          <RequirePermission tenant={["tenant:read"]}>
            <TenantDetailPage />
          </RequirePermission>
        ),
      },
      {
        path: "members",
        element: (
          <RequirePermission tenant={["member:read"]}>
            <MembersPage />
          </RequirePermission>
        ),
      },
      {
        path: "invitations",
        element: (
          <RequirePermission tenant={["tenant:invitation:manage"]}>
            <TenantInvitationsPage />
          </RequirePermission>
        ),
      },
      { path: "me/invitations", element: <MyInvitationsPage /> },
      { path: "me/security", element: <SecurityEventsPage /> },
      { path: "api-keys", element: <PersonalApiKeysPage /> },
      {
        path: "usage",
        element: (
          <RequirePermission tenant={["tenant:billing:read"]}>
            <TenantUsagePage />
          </RequirePermission>
        ),
      },
      {
        path: "audit-logs",
        element: (
          <RequirePermission tenant={["tenant:audit:read"]}>
            <TenantAuditPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/tenants",
        element: (
          <RequirePermission platform={["platform:tenant:read"]}>
            <AdminTenantsPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/users",
        element: (
          <RequirePermission platform={["platform:user:read"]}>
            <AdminUsersPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/audit-logs",
        element: (
          <RequirePermission platform={["platform:audit:read"]}>
            <AdminAuditPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/plans",
        element: (
          <RequirePermission platform={["platform:billing:manage"]}>
            <AdminPlansPage />
          </RequirePermission>
        ),
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
