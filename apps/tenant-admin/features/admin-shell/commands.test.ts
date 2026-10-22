import { describe, expect, it } from "vitest";

import { PERMISSIONS } from "../../lib/permissions";
import { buildPaletteCommands } from "./commands";
import { buildAdminNavItems } from "./navigation";

const ALL_PERMISSIONS = Object.values(PERMISSIONS);

const t = (key: string) => key;

describe("buildPaletteCommands", () => {
  it("returns only permission-free destinations and actions when the user has no permissions", () => {
    const { navigate, action } = buildPaletteCommands("/", t, []);
    const openKeys = buildAdminNavItems("/", t)
      .filter((item) => item.requiredPermissions.length === 0)
      .map((item) => `nav:${item.key}`);

    expect(navigate.map((command) => command.id)).toEqual(openKeys);
    expect(action.map((command) => command.id)).toEqual(["action:createSupportTicket"]);
  });

  it("gives every command a unique id", () => {
    const { navigate, action } = buildPaletteCommands("/", t, ALL_PERMISSIONS);
    const ids = [...navigate, ...action].map((command) => command.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it("labels a tab with its parent's title and a section item with its section", () => {
    const { byId } = buildPaletteCommands("/", t, ALL_PERMISSIONS);
    expect(byId.get("nav:logs")?.sectionLabel).toBe("nav.observe");
    expect(byId.get("nav:observe")?.sectionLabel).toBe("navSections.operations");
    expect(byId.get("nav:dashboard")?.sectionLabel).toBeUndefined();
  });

  it("adds the route key and href segments as English keywords", () => {
    const { byId } = buildPaletteCommands("/", t, ALL_PERMISSIONS);
    expect(byId.get("nav:logs")?.keywords).toContain("logs");
  });

  it("only includes actions the user is permitted to run", () => {
    const { action } = buildPaletteCommands("/", t, ["policy:manage"]);
    expect(action.map((command) => command.id)).toEqual([
      "action:createRateLimitRule",
      "action:createBudgetLimitRule",
      "action:createSupportTicket",
    ]);
  });
});
