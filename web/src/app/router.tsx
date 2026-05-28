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
      { index: true, element: <DashboardPage />, handle: { title: "仪表盘" } },
      { path: "tenants", element: <TenantsPage />, handle: { title: "组织管理" } },
      {
        path: "tenants/:tenantId",
        handle: { title: "组织详情" },
        element: (
          <RequirePermission tenant={["tenant:read"]}>
            <TenantDetailPage />
          </RequirePermission>
        ),
      },
      {
        path: "members",
        handle: { title: "成员管理" },
        element: (
          <RequirePermission tenant={["member:read"]}>
            <MembersPage />
          </RequirePermission>
        ),
      },
      {
        path: "invitations",
        handle: { title: "组织邀请" },
        element: (
          <RequirePermission tenant={["tenant:invitation:manage"]}>
            <TenantInvitationsPage />
          </RequirePermission>
        ),
      },
      { path: "me/invitations", element: <MyInvitationsPage />, handle: { title: "我的邀请" } },
      { path: "me/security", element: <SecurityEventsPage />, handle: { title: "我的安全事件" } },
      { path: "api-keys", element: <PersonalApiKeysPage />, handle: { title: "个人 API Key" } },
      {
        path: "usage",
        handle: { title: "配额与用量" },
        element: (
          <RequirePermission tenant={["tenant:billing:read"]}>
            <TenantUsagePage />
          </RequirePermission>
        ),
      },
      {
        path: "audit-logs",
        handle: { title: "组织审计日志" },
        element: (
          <RequirePermission tenant={["tenant:audit:read"]}>
            <TenantAuditPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/tenants",
        handle: { title: "平台组织管理" },
        element: (
          <RequirePermission platform={["platform:tenant:read"]}>
            <AdminTenantsPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/users",
        handle: { title: "平台用户管理" },
        element: (
          <RequirePermission platform={["platform:user:read"]}>
            <AdminUsersPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/audit-logs",
        handle: { title: "平台审计日志" },
        element: (
          <RequirePermission platform={["platform:audit:read"]}>
            <AdminAuditPage />
          </RequirePermission>
        ),
      },
      {
        path: "admin/plans",
        handle: { title: "套餐与配额管理" },
        element: (
          <RequirePermission platform={["platform:billing:manage"]}>
            <AdminPlansPage />
          </RequirePermission>
        ),
      },
      { path: "*", element: <NotFoundPage />, handle: { title: "页面不存在" } },
    ],
  },
]);
