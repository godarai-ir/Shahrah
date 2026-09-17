import { CreditCard, Gauge, LifeBuoy, Plug, Wallet } from "lucide-react";

import { canAnyPermission } from "../../lib/authz";
import type { Permission } from "../../lib/permissions";
import {
  ADMIN_NAV_SECTIONS,
  ADMIN_NAV_TAB_GROUPS,
  buildAdminNavItems,
  filterVisibleNavItems,
  resolveSectionEntryHref,
  type AdminNavItem,
  type AdminNavKey,
} from "./navigation";

export type PaletteCommand = {
  /** Stable across sessions; recent commands are stored by this id. */
  id: string;
  kind: "navigate" | "action";
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  href: string;
  sectionLabel?: string;
  keywords?: string[];
  requiredPermissions?: Permission[];
};

type Translate = (key: string) => string;

type ActionDefinition = {
  id: string;
  titleKey: string;
  icon: React.ComponentType<{ className?: string }>;
  href: string;
  requiredPermissions: Permission[];
};

/** Keep this small: each entry is an existing create route, never a side effect. */
const ACTION_DEFINITIONS: ActionDefinition[] = [
  {
    id: "action:createProviderAccount",
    titleKey: "palette.actions.createProviderAccount",
    icon: Plug,
    href: "/provider-accounts/new",
    requiredPermissions: ["provider_account:create"],
  },
  {
    id: "action:createRateLimitRule",
    titleKey: "palette.actions.createRateLimitRule",
    icon: Gauge,
    href: "/gateway-policies/rate-limiting-rules/new",
    requiredPermissions: ["policy:manage"],
  },
  {
    id: "action:createBudgetLimitRule",
    titleKey: "palette.actions.createBudgetLimitRule",
    icon: Wallet,
    href: "/gateway-policies/budget-limiting-rules/new",
    requiredPermissions: ["policy:manage"],
  },
  {
    id: "action:createSupportTicket",
    titleKey: "palette.actions.createSupportTicket",
    icon: LifeBuoy,
    href: "/support/tickets/new",
    requiredPermissions: ["support_ticket:create"],
  },
  {
    id: "action:topup",
    titleKey: "palette.actions.topup",
    icon: CreditCard,
    href: "/billing/topup",
    requiredPermissions: ["billing:manage", "settings:manage", "settings:list"],
  },
];

function hrefKeywords(href: string): string[] {
  return href.split("/").filter(Boolean);
}

/** Where an item sits in the sidebar: its section label, or its parent's title when it is a tab. */
export function sectionLabelFor(key: AdminNavKey, allItems: AdminNavItem[], t: Translate): string | undefined {
  const section = ADMIN_NAV_SECTIONS.find((candidate) => candidate.itemKeys.includes(key));
  if (section) return t(section.labelKey);
  const group = ADMIN_NAV_TAB_GROUPS.find((candidate) => candidate.tabKeys.includes(key));
  if (group) return allItems.find((item) => item.key === group.parentKey)?.title;
  return undefined;
}

export type PaletteCommandSet = {
  navigate: PaletteCommand[];
  action: PaletteCommand[];
  byId: Map<string, PaletteCommand>;
};

/** Every palette entry derives from the sidebar's nav model and the same permission check. */
export function buildPaletteCommands(pathname: string, t: Translate, permissions: readonly string[]): PaletteCommandSet {
  const allItems = buildAdminNavItems(pathname, t);
  const visibleItems = filterVisibleNavItems(allItems, [...permissions]);

  const navigate: PaletteCommand[] = visibleItems.map((item) => ({
    id: `nav:${item.key}`,
    kind: "navigate",
    title: item.title,
    icon: item.icon,
    href: resolveSectionEntryHref(item, allItems, visibleItems),
    sectionLabel: sectionLabelFor(item.key, allItems, t),
    keywords: [item.key, ...hrefKeywords(item.href)],
    requiredPermissions: item.requiredPermissions,
  }));

  const action: PaletteCommand[] = ACTION_DEFINITIONS
    .filter((definition) => canAnyPermission(permissions, definition.requiredPermissions))
    .map((definition) => ({
      id: definition.id,
      kind: "action",
      title: t(definition.titleKey),
      icon: definition.icon,
      href: definition.href,
      keywords: [definition.id.slice("action:".length), ...hrefKeywords(definition.href)],
      requiredPermissions: definition.requiredPermissions,
    }));

  const byId = new Map([...navigate, ...action].map((command) => [command.id, command]));
  return { navigate, action, byId };
}
