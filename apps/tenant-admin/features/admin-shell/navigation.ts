import {
  BadgeCheck,
  BarChart3,
  BookOpen,
  Building2,
  CircleDollarSign,
  CreditCard,
  Eye,
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
import { observeFiltersFromSearchParams, observeHref, type ObserveTab } from "../../lib/observe-filters";
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
  | "gatewayRules"
  | "guardrails"
  | "supportTickets"
  | "settings"
  | "observe"
  | "gatewayPolicies"
  | "accessAndUsers";

export type AdminNavItem = {
  key: AdminNavKey;
  href: string;
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  isActive: boolean;
  requiredPermissions: Permission[];
};

/** Tab order on "Users & access": identity first, then what it grants, then grouping. */
export const ACCESS_NAV_ITEM_KEYS: AdminNavItem["key"][] = [
  "users",
  "personalAccessTokens",
  "accessManagement",
  "teams",
  "orgUnits",
  "virtualUsers",
];

export const OBSERVE_TAB_KEYS: AdminNavKey[] = ["logs", "traces", "metrics"];
export const GATEWAY_POLICY_TAB_KEYS: AdminNavKey[] = ["gatewayRules", "guardrails"];
export const BILLING_TAB_KEYS: AdminNavKey[] = ["billing", "topup", "bankTransfer", "billingReports", "billingDocuments", "billingProfile", "plan"];

export type AdminNavSectionKey = "build" | "operations" | "organization";

export type AdminNavSection = {
  key: AdminNavSectionKey;
  labelKey: string;
  itemKeys: AdminNavKey[];
};

/** The three always-open sidebar sections. Array order is display order. */
export const ADMIN_NAV_SECTIONS: AdminNavSection[] = [
  { key: "build", labelKey: "navSections.build", itemKeys: ["providerAccounts", "virtualModels", "prompts", "playground"] },
  { key: "operations", labelKey: "navSections.operations", itemKeys: ["observe", "gatewayPolicies"] },
  { key: "organization", labelKey: "navSections.organization", itemKeys: ["accessAndUsers", "billing"] },
];

/** Items rendered above the sections, outside any group. */
export const ROOT_NAV_ITEM_KEYS: AdminNavKey[] = ["dashboard"];

export type AdminNavTabGroup = {
  /** Sidebar entry that owns the tabs. */
  parentKey: AdminNavKey;
  /** Landing route of the parent page; it shows the first visible tab. */
  rootHref: string;
  tabKeys: AdminNavKey[];
};

/** Former sidebar entries now reachable as tabs. Each tab keeps its own route as a deep link. */
export const ADMIN_NAV_TAB_GROUPS: AdminNavTabGroup[] = [
  { parentKey: "observe", rootHref: "/observe", tabKeys: OBSERVE_TAB_KEYS },
  { parentKey: "gatewayPolicies", rootHref: "/gateway-policies", tabKeys: GATEWAY_POLICY_TAB_KEYS },
  { parentKey: "accessAndUsers", rootHref: "/users", tabKeys: ACCESS_NAV_ITEM_KEYS },
  { parentKey: "billing", rootHref: "/billing", tabKeys: BILLING_TAB_KEYS },
];

export type ResolvedNavTabs = {
  group: AdminNavTabGroup;
  tabs: Array<AdminNavItem & { isCurrent: boolean }>;
};

/**
 * Finds the tab strip for a list route. Detail and form routes (e.g. /logs/abc, …/new) get no tabs.
 * On a group root that is not itself a tab (/observe), the first visible tab is current.
 * Observe tabs share one filter state, so their links carry the shared part of `searchParams`.
 */
export function resolveNavTabs(
  pathname: string,
  visibleItems: AdminNavItem[],
  searchParams?: Pick<URLSearchParams, "get">,
): ResolvedNavTabs | null {
  const byKey = new Map(visibleItems.map((item) => [item.key, item]));
  for (const group of ADMIN_NAV_TAB_GROUPS) {
    const tabs = group.tabKeys.map((key) => byKey.get(key)).filter((item): item is AdminNavItem => Boolean(item));
    if (tabs.length === 0) continue;
    const exact = tabs.find((tab) => tab.href === pathname);
    if (!exact && pathname !== group.rootHref) continue;
    const currentHref = exact?.href ?? tabs[0].href;
    const observeFilters = group.parentKey === "observe" && searchParams ? observeFiltersFromSearchParams(searchParams) : null;
    return {
      group,
      tabs: tabs.map((tab) => ({
        ...tab,
        href: observeFilters ? observeHref(OBSERVE_TAB_BY_KEY[tab.key] ?? "logs", observeFilters) : tab.href,
        isCurrent: tab.href === currentHref,
      })),
    };
  }
  return null;
}

const OBSERVE_TAB_BY_KEY: Partial<Record<AdminNavKey, ObserveTab>> = { logs: "logs", traces: "traces", metrics: "metrics" };

/**
 * Points a parent entry at its first visible tab when its landing route is a tab the user cannot open
 * (e.g. "Users & access" for someone without user:list). Group roots such as /observe pick a tab themselves.
 */
export function resolveSectionEntryHref(item: AdminNavItem, allItems: AdminNavItem[], visibleItems: AdminNavItem[]): string {
  const group = ADMIN_NAV_TAB_GROUPS.find((candidate) => candidate.parentKey === item.key);
  if (!group) return item.href;
  const landingTab = allItems.find((candidate) => group.tabKeys.includes(candidate.key) && candidate.href === item.href);
  if (!landingTab || visibleItems.some((visible) => visible.key === landingTab.key)) return item.href;
  const firstVisibleTab = group.tabKeys.map((key) => visibleItems.find((visible) => visible.key === key)).find(Boolean);
  return firstVisibleTab?.href ?? item.href;
}

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
      isActive: pathname.startsWith("/billing"),
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
      // Rate and budget rules share one list at the section root; the old per-type routes redirect here.
      key: "gatewayRules",
      href: "/gateway-policies",
      title: t("nav.gatewayRules"),
      icon: Gauge,
      isActive: pathname.startsWith("/gateway-policies"),
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
      key: "observe",
      href: "/observe",
      title: t("nav.observe"),
      icon: Eye,
      isActive: ["/observe", "/logs", "/traces", "/metrics"].some((prefix) => pathname.startsWith(prefix)),
      requiredPermissions: ["logs:read_metadata", "logs:read_content", "settings:list"],
    },
    {
      key: "gatewayPolicies",
      href: "/gateway-policies",
      title: t("nav.gatewayPolicies"),
      icon: ShieldCheck,
      isActive: pathname.startsWith("/gateway-policies") || pathname.startsWith("/guardrails"),
      requiredPermissions: ["policy:list", "policy:manage"],
    },
    {
      key: "accessAndUsers",
      href: "/users",
      title: t("nav.accessAndUsers"),
      icon: Users,
      isActive: ["/users", "/teams", "/org-units", "/roles", "/personal-access-tokens", "/virtual-users"].some((prefix) => pathname.startsWith(prefix)),
      // Personal access tokens are open to everyone; the individual tabs gate themselves.
      requiredPermissions: [],
    },
  ];
}

export function filterVisibleNavItems(navItems: AdminNavItem[], permissions: string[]): AdminNavItem[] {
  return navItems.filter((item) => canAnyPermission(permissions, item.requiredPermissions));
}
