import { describe, expect, it } from "vitest";

import {
  ACCESS_NAV_ITEM_KEYS,
  ADMIN_NAV_SECTIONS,
  BILLING_TAB_KEYS,
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
    "nav.users": "کاربران",
    "nav.personalAccessTokens": "توکن‌های دسترسی شخصی",
    "nav.teams": "تیم‌ها",
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
      "personalAccessTokens",
      "oidcClients",
      "accessManagement",
      "teams",
      "virtualUsers",
    ]);
  });

  it("shows connected apps only to members who can manage settings", () => {
    const keysFor = (permissions: string[]) =>
      filterVisibleNavItems(buildAdminNavItems("/oidc-clients", t), permissions).map((item) => item.key);

    expect(keysFor(["settings:list"])).not.toContain("oidcClients");
    expect(keysFor(["settings:manage"])).toContain("oidcClients");
    const active = buildAdminNavItems("/oidc-clients", t).filter((item) => item.isActive).map((item) => item.key);
    expect(active).toEqual(expect.arrayContaining(["oidcClients", "accessAndUsers"]));
  });

  it("shows the org structure item for team:read only", () => {
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

  it("keeps support visible without permissions", () => {
    expect(getVisibleTitles([])).toContain("پشتیبانی");
  });
});

describe("admin sidebar sections", () => {
  it("keeps the sidebar at ten first-level entries", () => {
    const itemCount = ADMIN_NAV_SECTIONS.reduce((sum, section) => sum + section.itemKeys.length, 0);
    expect(ROOT_NAV_ITEM_KEYS.length + itemCount).toBe(10);
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
    expect(activeKeys("/guardrails")).toContain("guardrails");
    expect(activeKeys("/guardrails")).not.toContain("gatewayPolicies");
    expect(activeKeys("/gateway-policies/budget-limiting-rules")).toContain("gatewayPolicies");
    expect(activeKeys("/gateway-policies")).toContain("gatewayRules");
    expect(activeKeys("/teams")).toContain("accessAndUsers");
    expect(activeKeys("/teams")).toContain("teams");
    expect(activeKeys("/billing/topup")).toContain("billing");
  });

  it("keeps every retired key reachable as a tab key", () => {
    const tabKeys = [...ACCESS_NAV_ITEM_KEYS, ...OBSERVE_TAB_KEYS, ...BILLING_TAB_KEYS];
    for (const key of ["users", "teams", "logs", "metrics", "topup", "plan"] as const) {
      expect(tabKeys).toContain(key);
    }
  });

  it("opens the matching tab on a legacy deep link and the first visible tab on a group root", () => {
    const visible = filterVisibleNavItems(buildAdminNavItems("/metrics", t), ["logs:read_metadata", "settings:list"]);
    expect(resolveNavTabs("/metrics", visible)?.tabs.find((tab) => tab.isCurrent)?.key).toBe("metrics");
    expect(resolveNavTabs("/observe", visible)?.tabs.find((tab) => tab.isCurrent)?.key).toBe("logs");
    expect(resolveNavTabs("/logs/abc", visible)).toBeNull();
  });

  it("carries the shared observe query on observe tab links and nothing else", () => {
    const visible = filterVisibleNavItems(buildAdminNavItems("/logs", t), ["logs:read_metadata", "settings:list"]);
    const params = new URLSearchParams("range=last7d&model=gpt-5&status=500&sort_by=duration_ms&trace=tr_1");
    const tabs = resolveNavTabs("/logs", visible, params)?.tabs ?? [];

    expect(tabs.map((tab) => tab.href)).toEqual([
      "/logs?range=last7d&model=gpt-5&status=500",
      "/traces?range=last7d&model=gpt-5&status=500",
      "/metrics?range=last7d&model=gpt-5&status=500",
    ]);
    expect(tabs.find((tab) => tab.isCurrent)?.key).toBe("logs");
  });

  it("keeps plain hrefs on observe tabs without a query and on other tab groups", () => {
    const visible = filterVisibleNavItems(buildAdminNavItems("/logs", t), ["logs:read_metadata", "settings:list", "policy:list"]);
    expect(resolveNavTabs("/logs", visible, new URLSearchParams())?.tabs.map((tab) => tab.href)).toEqual(["/logs", "/traces", "/metrics"]);
    expect(resolveNavTabs("/guardrails", visible, new URLSearchParams("model=m1"))).toBeNull();
  });

  it("gives guardrails its own sidebar entry in operations", () => {
    const ops = ADMIN_NAV_SECTIONS.find((section) => section.key === "operations");
    expect(ops?.itemKeys).toEqual(["observe", "gatewayPolicies", "guardrails"]);
  });

  it("keeps the guardrails entry active across the wizard routes", () => {
    for (const path of ["/guardrails", "/guardrails/new", "/guardrails/pii-mask/edit"]) {
      const active = buildAdminNavItems(path, t).filter((item) => item.isActive).map((item) => item.key);
      expect(active).toContain("guardrails");
      expect(active).not.toContain("gatewayPolicies");
    }
  });

  it("shows no tab strip on gateway policies", () => {
    const visible = filterVisibleNavItems(buildAdminNavItems("/gateway-policies", t), ["policy:list"]);
    expect(resolveNavTabs("/gateway-policies", visible)).toBeNull();
  });

  it("hides gateway policies from callers who can manage but not list policies", () => {
    const visible = filterVisibleNavItems(buildAdminNavItems("/", t), ["policy:manage"]).map((item) => item.key);
    expect(visible).not.toContain("gatewayPolicies");
  });

  it("sends users without user:list to their first permitted access tab", () => {
    const all = buildAdminNavItems("/", t);
    const visible = filterVisibleNavItems(all, ["team:read"]);
    const entry = all.find((item) => item.key === "accessAndUsers")!;
    // The keys tab needs no permission, so it is the first fallback for anyone.
    expect(resolveSectionEntryHref(entry, all, visible)).toBe("/personal-access-tokens");
    expect(resolveSectionEntryHref(entry, all, filterVisibleNavItems(all, ["user:list"]))).toBe("/users");
  });
});
