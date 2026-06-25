import {
  Activity,
  BadgeCheck,
  BarChart3,
  Building2,
  CircleDollarSign,
  CreditCard,
  Gauge,
  LayoutDashboard,
  Plug,
  Route,
  Settings,
  ShieldCheck,
  UserCog,
  Users,
} from "lucide-react";

import { canAnyPermission } from "../../lib/authz";

export type AdminNavKey =
  | "dashboard"
  | "topup"
  | "plan"
  | "providerAccounts"
  | "virtualModels"
  | "virtualUsers"
  | "accessManagement"
  | "traces"
  | "metrics"
  | "orgUnits"
  | "rateLimitPolicies"
  | "budgetLimitPolicies"
  | "guardrails"
  | "settings";

export type AdminNavItem = {
  key: AdminNavKey;
  href: string;
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  isActive: boolean;
  requiredPermissions: string[];
};

const ACCESS_SECTION_PERMISSIONS = [
  "user:list",
  "user:manage",
  "team:read",
  "team:manage",
  "virtual_account:read",
  "virtual_account:manage",
  "virtual_account:create",
  "role:list",
  "role:manage",
];

export function buildAdminNavItems(pathname: string, t: (key: string) => string): AdminNavItem[] {
  return [
    {
      key: "dashboard",
      href: "/",
      title: t("nav.dashboard"),
      icon: LayoutDashboard,
      isActive: pathname === "/",
      requiredPermissions: [],
    },
    {
      key: "topup",
      href: "/billing/topup",
      title: t("nav.topup"),
      icon: CreditCard,
      isActive: pathname.startsWith("/billing/topup"),
      requiredPermissions: ["settings:list"],
    },
    {
      key: "plan",
      href: "/billing/plan",
      title: t("nav.plan"),
      icon: BadgeCheck,
      isActive: pathname.startsWith("/billing/plan"),
      requiredPermissions: ["settings:list"],
    },
    {
      key: "providerAccounts",
      href: "/provider-accounts",
      title: t("nav.providerAccounts"),
      icon: Plug,
      isActive: pathname.startsWith("/provider-accounts"),
      requiredPermissions: ["provider_account:read", "provider_account:manage", "provider_account:create"],
    },
    {
      key: "virtualModels",
      href: "/virtual-models",
      title: t("nav.virtualModels"),
      icon: Route,
      isActive: pathname.startsWith("/virtual-models"),
      requiredPermissions: ["virtual_account:read", "virtual_account:manage", "virtual_account:create", "provider_account:read"],
    },
    {
      key: "virtualUsers",
      href: "/virtual-users",
      title: t("nav.virtualUsers"),
      icon: Users,
      isActive: pathname.startsWith("/virtual-users"),
      requiredPermissions: ["virtual_account:read", "virtual_account:manage", "virtual_account:create"],
    },
    {
      key: "accessManagement",
      href: "/roles",
      title: t("nav.accessManagement"),
      icon: UserCog,
      isActive:
        pathname.startsWith("/users")
        || pathname.startsWith("/roles")
        || pathname.startsWith("/teams")
        || pathname.startsWith("/personal-access-tokens"),
      requiredPermissions: ACCESS_SECTION_PERMISSIONS,
    },
    {
      key: "traces",
      href: "/traces",
      title: t("nav.traces"),
      icon: Activity,
      isActive: pathname.startsWith("/traces"),
      requiredPermissions: ["settings:list"],
    },
    {
      key: "metrics",
      href: "/metrics",
      title: t("nav.metrics"),
      icon: BarChart3,
      isActive: pathname.startsWith("/metrics"),
      requiredPermissions: ["settings:list"],
    },
    {
      key: "orgUnits",
      href: "/org-units",
      title: t("nav.orgUnits"),
      icon: Building2,
      isActive: pathname.startsWith("/org-units"),
      requiredPermissions: ["org_unit:list", "org_unit:read"],
    },
    {
      key: "rateLimitPolicies",
      href: "/gateway-policies/rate-limiting-rules",
      title: t("nav.rateLimitPolicies"),
      icon: Gauge,
      isActive: pathname.startsWith("/gateway-policies/rate-limiting-rules"),
      requiredPermissions: ["policy:list"],
    },
    {
      key: "budgetLimitPolicies",
      href: "/gateway-policies/budget-limiting-rules",
      title: t("nav.budgetLimitPolicies"),
      icon: CircleDollarSign,
      isActive: pathname.startsWith("/gateway-policies/budget-limiting-rules"),
      requiredPermissions: ["policy:list"],
    },
    {
      key: "guardrails",
      href: "/guardrails",
      title: t("nav.guardrails"),
      icon: ShieldCheck,
      isActive: pathname.startsWith("/guardrails"),
      requiredPermissions: ["policy:list"],
    },
    {
      key: "settings",
      href: "/settings",
      title: t("nav.settings"),
      icon: Settings,
      isActive: pathname.startsWith("/settings"),
      requiredPermissions: ["settings:list"],
    },
  ];
}

export function filterVisibleNavItems(navItems: AdminNavItem[], permissions: string[]): AdminNavItem[] {
  return navItems.filter((item) => canAnyPermission(permissions, item.requiredPermissions));
}
