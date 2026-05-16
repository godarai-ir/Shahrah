import { describe, expect, it } from "vitest";

import { buildAdminNavItems, filterVisibleNavItems } from "./admin-shell-nav";

function t(key: string) {
  const dictionary: Record<string, string> = {
    "nav.accessManagement": "دسترسی‌ها",
  };
  return dictionary[key] ?? key;
}

function getVisibleTitles(permissions: string[]) {
  const navItems = buildAdminNavItems("/roles", t);
  return filterVisibleNavItems(navItems, permissions).map((item) => item.title);
}

describe("admin access nav label", () => {
  it("shows Access label for team:read only", () => {
    expect(getVisibleTitles(["team:read"])).toContain("دسترسی‌ها");
  });

  it("shows Access label for user:list only", () => {
    expect(getVisibleTitles(["user:list"])).toContain("دسترسی‌ها");
  });

  it("shows Access label for combined user:list and team:read", () => {
    expect(getVisibleTitles(["user:list", "team:read"])).toContain("دسترسی‌ها");
  });
});
