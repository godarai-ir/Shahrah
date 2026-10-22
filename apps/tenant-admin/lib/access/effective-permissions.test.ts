import { describe, expect, it } from "vitest";

import { collectEffectivePermissions } from "./effective-permissions";

const catalog = [
  { key: "user:list", group: "users", description: "" },
  { key: "user:manage", group: "users", description: "" },
  { key: "billing:read", group: "billing", description: "" },
];

const roles = [
  { name: "member", permissions: ["user:list"] },
  { name: "billing_manager", permissions: ["billing:read", "user:list"] },
  { name: "empty_role", permissions: [] },
];

describe("collectEffectivePermissions", () => {
  it("unions permissions across every role the member holds", () => {
    const result = collectEffectivePermissions(["member", "billing_manager"], roles, catalog);
    expect(result.total).toBe(2);
    expect(result.groups.map((group) => group.group)).toEqual(["billing", "users"]);
  });

  it("records every role that grants a shared permission", () => {
    const result = collectEffectivePermissions(["member", "billing_manager"], roles, catalog);
    const users = result.groups.find((group) => group.group === "users");
    expect(users?.grants[0]).toEqual({ key: "user:list", grantedBy: ["member", "billing_manager"] });
  });

  it("returns nothing for a member with no roles", () => {
    expect(collectEffectivePermissions([], roles, catalog)).toEqual({ groups: [], total: 0 });
  });

  it("returns nothing for a role that grants nothing", () => {
    expect(collectEffectivePermissions(["empty_role"], roles, catalog).total).toBe(0);
  });

  it("ignores a role name that is not in the catalog", () => {
    expect(collectEffectivePermissions(["ghost"], roles, catalog).total).toBe(0);
  });

  it("matches role names case-insensitively", () => {
    expect(collectEffectivePermissions(["MEMBER"], roles, catalog).total).toBe(1);
  });

  it("still shows a permission the catalog does not describe, grouped by its key", () => {
    const result = collectEffectivePermissions(
      ["agent"],
      [{ name: "agent", permissions: ["export_job:run"] }],
      catalog,
    );
    expect(result.groups).toEqual([
      { group: "export_job", grants: [{ key: "export_job:run", grantedBy: ["agent"] }] },
    ]);
  });

  it("sorts grants inside a group", () => {
    const result = collectEffectivePermissions(["admin"], [{ name: "admin", permissions: ["user:manage", "user:list"] }], catalog);
    expect(result.groups[0].grants.map((grant) => grant.key)).toEqual(["user:list", "user:manage"]);
  });
});
