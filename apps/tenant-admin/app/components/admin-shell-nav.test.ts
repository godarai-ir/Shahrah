import { describe, expect, it } from "vitest";

import { ACCESS_NAV_ITEM_KEYS, buildAdminNavItems, filterVisibleNavItems, shouldHandleActiveNavClick } from "./admin-shell-nav";

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
