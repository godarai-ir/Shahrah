import {
  BadgeCheck,
  BarChart3,
  BookOpen,
  Building2,
  CircleDollarSign,
  CreditCard,
  FileSearch2,
  FileText,
  FlaskConical,
  Gauge,
  Landmark,
  LayoutDashboard,
  LineChart,
  KeyRound,
  Plug,
  Route,
  Settings,
  ShieldCheck,
  UserCog,
  UserCircle2,
  Users,
  LifeBuoy,
} from "lucide-react";

import { canAnyPermission } from "../../lib/authz";
import type { Permission } from "../../lib/permissions";

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
  | "playground"
  | "virtualUsers"
  | "accountSettings"
  | "tracePrivacySettings"
  | "accessManagement"
  | "personalAccessTokens"
  | "teams"
  | "users"
  | "logs"
  | "traces"
  | "metrics"
  | "orgUnits"
  | "rateLimitPolicies"
  | "budgetLimitPolicies"
  | "guardrails"
  | "supportTickets"
  | "settings";

export type AdminNavItem = {
  key: AdminNavKey;
  href: string;
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  isActive: boolean;
  requiredPermissions: Permission[];
};

export const ACCESS_NAV_ITEM_KEYS: AdminNavItem["key"][] = [
  "users",
  "teams",
  "orgUnits",
  "accessManagement",
  "personalAccessTokens",
  "virtualUsers",
];

export function shouldHandleActiveNavClick(event: Pick<MouseEvent, "button" | "metaKey" | "ctrlKey" | "shiftKey" | "altKey">): boolean {
  return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
}

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
      key: "supportTickets",
      href: "/support/tickets",
      title: t("nav.supportTickets"),
      icon: LifeBuoy,
      isActive: pathname.startsWith("/support/tickets"),
      requiredPermissions: ["support_ticket:read_own", "support_ticket:read_all"],
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
      key: "playground",
      href: "/playground",
      title: t("nav.playground"),
      icon: FlaskConical,
      isActive: pathname.startsWith("/playground"),
      requiredPermissions: ["playground:read", "playground:run", "policy:list", "policy:manage"],
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
      title: t("nav.roles"),
      icon: UserCog,
      isActive: pathname.startsWith("/roles"),
      requiredPermissions: ["role:list", "role:manage"],
    },
    {
      key: "personalAccessTokens",
      href: "/personal-access-tokens",
      title: t("nav.personalAccessTokens"),
      icon: KeyRound,
      isActive: pathname.startsWith("/personal-access-tokens"),
      requiredPermissions: [],
    },
    {
      key: "teams",
      href: "/teams",
      title: t("nav.teams"),
      icon: Users,
      isActive: pathname.startsWith("/teams"),
      requiredPermissions: ["team:read", "team:manage"],
    },
    {
      key: "users",
      href: "/users",
      title: t("nav.users"),
      icon: UserCircle2,
      isActive: pathname.startsWith("/users"),
      requiredPermissions: ["user:list", "user:manage"],
    },
    {
      key: "logs",
      href: "/logs",
      title: t("nav.logs"),
      icon: FileText,
      isActive: pathname.startsWith("/logs"),
      requiredPermissions: ["logs:read_metadata", "logs:read_content"],
    },
    {
      key: "traces",
      href: "/traces",
      title: t("nav.traces"),
      icon: FileSearch2,
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
