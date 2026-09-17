import { describe, expect, it } from "vitest";

import {
  ACCESS_NAV_ITEM_KEYS,
  ADMIN_NAV_SECTIONS,
  BILLING_TAB_KEYS,
  GATEWAY_POLICY_TAB_KEYS,
  OBSERVE_TAB_KEYS,
  ROOT_NAV_ITEM_KEYS,
  buildAdminNavItems,
  filterVisibleNavItems,
  resolveNavTabs,
  resolveSectionEntryHref,
  shouldHandleActiveNavClick,
} from "./navigation";

function t(key: string) {
  const dictionary: Record<string, string> = {
    "nav.roles": "نقش‌ها",
    "nav.teams": "تیم‌ها",
    "nav.users": "کاربران",
    "nav.personalAccessTokens": "توکن‌های دسترسی شخصی",
    "nav.orgUnits": "واحدهای سازمانی",
    "nav.virtualUsers": "کاربران مجازی",
    "nav.supportTickets": "پشتیبانی",
  };
  return dictionary[key] ?? key;
}

function getVisibleTitles(permissions: string[]) {
  const navItems = buildAdminNavItems("/roles", t);
  return filterVisibleNavItems(navItems, permissions).map((item) => item.title);
}

describe("admin access navigation items", () => {
  it("handles a plain active-item click so stale detail query parameters are cleared", () => {
    expect(shouldHandleActiveNavClick({ button: 0, metaKey: false, ctrlKey: false, shiftKey: false, altKey: false })).toBe(true);
    expect(shouldHandleActiveNavClick({ button: 0, metaKey: false, ctrlKey: true, shiftKey: false, altKey: false })).toBe(false);
  });

  it("orders access items by common administrative workflow", () => {
    expect(ACCESS_NAV_ITEM_KEYS).toEqual([
      "users",
      "teams",
      "orgUnits",
      "accessManagement",
      "personalAccessTokens",
      "virtualUsers",
    ]);
  });

  it("shows the teams item for team:read only", () => {
    expect(getVisibleTitles(["team:read"])).toContain("تیم‌ها");
  });

  it("shows the users item for user:list only", () => {
    expect(getVisibleTitles(["user:list"])).toContain("کاربران");
  });

  it("keeps personal access tokens visible without permissions", () => {
    expect(getVisibleTitles([])).toContain("توکن‌های دسترسی شخصی");
  });

  it("shows roles only with role permissions", () => {
    expect(getVisibleTitles(["role:list"])).toContain("نقش‌ها");
    expect(getVisibleTitles(["team:read"])).not.toContain("نقش‌ها");
  });

  it("shows support only for users who can read tickets", () => {
    expect(getVisibleTitles(["support_ticket:read_own"])).toContain("پشتیبانی");
    expect(getVisibleTitles(["support_ticket:read_all"])).toContain("پشتیبانی");
    expect(getVisibleTitles(["support_ticket:create"])).not.toContain("پشتیبانی");
  });
});

describe("admin sidebar sections", () => {
  it("keeps the sidebar at nine first-level entries", () => {
    const itemCount = ADMIN_NAV_SECTIONS.reduce((sum, section) => sum + section.itemKeys.length, 0);
    expect(ROOT_NAV_ITEM_KEYS.length + itemCount).toBe(9);
  });

  it("only references keys that buildAdminNavItems produces", () => {
    const keys = new Set(buildAdminNavItems("/", t).map((item) => item.key));
    for (const key of [...ROOT_NAV_ITEM_KEYS, ...ADMIN_NAV_SECTIONS.flatMap((section) => section.itemKeys)]) {
      expect(keys).toContain(key);
    }
  });

  it("marks the observe entry active on every legacy monitoring route", () => {
    for (const path of ["/observe", "/logs", "/traces", "/metrics"]) {
      const item = buildAdminNavItems(path, t).find((i) => i.key === "observe");
      expect(item?.isActive).toBe(true);
    }
  });

  it("marks parent entries active on their legacy routes", () => {
    const activeKeys = (path: string) => buildAdminNavItems(path, t).filter((item) => item.isActive).map((item) => item.key);
    expect(activeKeys("/guardrails")).toContain("gatewayPolicies");
    expect(activeKeys("/gateway-policies/budget-limiting-rules")).toContain("gatewayPolicies");
    expect(activeKeys("/teams")).toContain("accessAndUsers");
    expect(activeKeys("/billing/topup")).toContain("billing");
  });

  it("keeps every retired key reachable as a tab key", () => {
    const tabKeys = [...ACCESS_NAV_ITEM_KEYS, ...OBSERVE_TAB_KEYS, ...GATEWAY_POLICY_TAB_KEYS, ...BILLING_TAB_KEYS];
    for (const key of ["users", "teams", "logs", "metrics", "guardrails", "topup", "plan"] as const) {
      expect(tabKeys).toContain(key);
    }
  });

  it("opens the matching tab on a legacy deep link and the first visible tab on a group root", () => {
    const visible = filterVisibleNavItems(buildAdminNavItems("/metrics", t), ["logs:read_metadata", "settings:list"]);
    expect(resolveNavTabs("/metrics", visible)?.tabs.find((tab) => tab.isCurrent)?.key).toBe("metrics");
    expect(resolveNavTabs("/observe", visible)?.tabs.find((tab) => tab.isCurrent)?.key).toBe("logs");
    expect(resolveNavTabs("/logs/abc", visible)).toBeNull();
  });

  it("sends users without user:list to their first permitted access tab", () => {
    const all = buildAdminNavItems("/", t);
    const visible = filterVisibleNavItems(all, ["team:read"]);
    const entry = all.find((item) => item.key === "accessAndUsers")!;
    expect(resolveSectionEntryHref(entry, all, visible)).toBe("/teams");
    expect(resolveSectionEntryHref(entry, all, filterVisibleNavItems(all, ["user:list"]))).toBe("/users");
  });
});
