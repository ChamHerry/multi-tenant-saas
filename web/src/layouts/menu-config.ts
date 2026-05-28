import {
  BarChart3,
  Building2,
  CreditCard,
  Inbox,
  KeyRound,
  LayoutDashboard,
  MailPlus,
  ScrollText,
  ShieldCheck,
  UserCog,
  Users,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type {
  PlatformPermission,
  TenantPermission,
} from "@/features/access/access-types";

export type MenuItem = {
  to: string;
  label: string;
  icon: LucideIcon;
  requiredTenantPermissions?: TenantPermission[];
  requiredPlatformPermissions?: PlatformPermission[];
};

export const tenantNavItems: MenuItem[] = [
  { to: "/", label: "仪表盘", icon: LayoutDashboard },
  { to: "/tenants", label: "组织", icon: Building2 },
  {
    to: "/members",
    label: "成员",
    icon: Users,
    requiredTenantPermissions: ["member:read"],
  },
  {
    to: "/invitations",
    label: "邀请",
    icon: MailPlus,
    requiredTenantPermissions: ["tenant:invitation:manage"],
  },
  {
    to: "/usage",
    label: "配额",
    icon: BarChart3,
    requiredTenantPermissions: ["tenant:billing:read"],
  },
  {
    to: "/audit-logs",
    label: "审计",
    icon: ScrollText,
    requiredTenantPermissions: ["tenant:audit:read"],
  },
  { to: "/me/invitations", label: "我的邀请", icon: Inbox },
  { to: "/me/security", label: "安全事件", icon: ShieldCheck },
  { to: "/api-keys", label: "个人 API Keys", icon: KeyRound },
];

export const platformNavItems: MenuItem[] = [
  {
    to: "/admin/tenants",
    label: "平台组织",
    icon: ShieldCheck,
    requiredPlatformPermissions: ["platform:tenant:read"],
  },
  {
    to: "/admin/users",
    label: "平台用户",
    icon: UserCog,
    requiredPlatformPermissions: ["platform:user:read"],
  },
  {
    to: "/admin/audit-logs",
    label: "平台审计",
    icon: ScrollText,
    requiredPlatformPermissions: ["platform:audit:read"],
  },
  {
    to: "/admin/plans",
    label: "套餐配额",
    icon: CreditCard,
    requiredPlatformPermissions: ["platform:billing:manage"],
  },
];
