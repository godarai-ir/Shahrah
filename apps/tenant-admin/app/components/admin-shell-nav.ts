import {
  Activity,
  BadgeCheck,
  BarChart3,
  BookOpen,
  Building2,
  CircleDollarSign,
  CreditCard,
  FileText,
  Gauge,
  Landmark,
  LayoutDashboard,
  LineChart,
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
  | "quickStart"
  | "billing"
  | "topup"
  | "bankTransfer"
  | "billingReports"
  | "billingProfile"
  | "billingDocuments"
  | "plan"
  | "providerAccounts"
  | "virtualModels"
  | "prompts"
  | "virtualUsers"
  | "accountSettings"
  | "tracePrivacySettings"
  | "accessManagement"
  | "logs"
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
      key: "quickStart",
      href: "/quick-start",
      title: t("nav.quickStart"),
      icon: BookOpen,
      isActive: pathname.startsWith("/quick-start"),
      requiredPermissions: [],
    },
    {
      key: "billing",
      href: "/billing",
      title: t("nav.billing"),
      icon: CircleDollarSign,
      isActive: pathname === "/billing",
      requiredPermissions: ["billing:read", "billing:manage", "billing:documents_read", "settings:list"],
    },
    {
      key: "topup",
      href: "/billing/topup",
      title: t("nav.topup"),
      icon: CreditCard,
      isActive: pathname.startsWith("/billing/topup"),
      requiredPermissions: ["billing:manage", "settings:manage", "settings:list"],
    },
    {
      key: "bankTransfer",
      href: "/billing/bank-transfer",
      title: t("nav.bankTransfer"),
      icon: Landmark,
      isActive: pathname.startsWith("/billing/bank-transfer"),
      requiredPermissions: ["billing:manage", "settings:manage"],
    },
    {
      key: "billingReports",
      href: "/billing/reports",
      title: t("nav.billingReports"),
      icon: LineChart,
      isActive: pathname.startsWith("/billing/reports"),
      requiredPermissions: ["billing:read", "settings:list"],
    },
    {
      key: "billingProfile",
      href: "/billing/profile",
      title: t("nav.billingProfile"),
      icon: UserCog,
      isActive: pathname.startsWith("/billing/profile"),
      requiredPermissions: ["billing:profile_manage", "billing:manage", "settings:manage"],
    },
    {
      key: "billingDocuments",
      href: "/billing/documents",
      title: t("nav.billingDocuments"),
      icon: FileText,
      isActive: pathname.startsWith("/billing/documents"),
      requiredPermissions: ["billing:documents_read", "billing:read", "settings:list"],
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
      key: "prompts",
      href: "/prompts",
      title: t("nav.prompts"),
      icon: FileText,
      isActive: pathname.startsWith("/prompts"),
      requiredPermissions: ["policy:list", "policy:manage"],
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
      key: "accountSettings",
      href: "/settings/account",
      title: t("nav.accountSettings"),
      icon: Settings,
      isActive: pathname.startsWith("/settings/account"),
      requiredPermissions: ["settings:list"],
    },
    {
      key: "tracePrivacySettings",
      href: "/settings/trace-privacy",
      title: t("nav.tracePrivacySettings"),
      icon: Settings,
      isActive: pathname.startsWith("/settings/trace-privacy"),
      requiredPermissions: ["settings:list"],
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
      key: "logs",
      href: "/logs",
      title: t("nav.logs"),
      icon: Activity,
      isActive: pathname.startsWith("/logs"),
      requiredPermissions: ["logs:read_metadata", "logs:read_content"],
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
  ];
}

export function filterVisibleNavItems(navItems: AdminNavItem[], permissions: string[]): AdminNavItem[] {
  return navItems.filter((item) => canAnyPermission(permissions, item.requiredPermissions));
}
