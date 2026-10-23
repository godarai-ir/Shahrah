import { describe, expect, it } from "vitest";

import type { SeatType } from "@/lib/seats";

import { filterMembers, listUserRoleNames } from "./member-filters";
import { parseUsersURLState, patchUsersURL } from "./url-state";
import type { TenantAdminUser } from "@/lib/contracts/tenant-users";

function member(overrides: Partial<TenantAdminUser> & { id: string }): TenantAdminUser {
  return {
    tenant_slug: "acme",
    role: "member",
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

const roster = [
  member({ id: "1", email: "ada@example.com", roles: ["tenant_admin", "billing_manager"] }),
  member({ id: "2", email: "grace@example.com", roles: ["member"] }),
  member({ id: "3", email: "alan@example.com", roles: ["member"], active: false }),
];

describe("listUserRoleNames", () => {
  it("returns every role, not just the first", () => {
    expect(listUserRoleNames(roster[0])).toEqual(["tenant_admin", "billing_manager"]);
  });

  it("falls back to the legacy single role field when the array is absent", () => {
    expect(listUserRoleNames(member({ id: "x", role: "billing_manager" }))).toEqual(["billing_manager"]);
  });

  it("drops roles hidden from the UI", () => {
    expect(listUserRoleNames(member({ id: "x", roles: ["proxy_admin", "member"] }))).toEqual(["member"]);
  });

  it("dedupes and lowercases", () => {
    expect(listUserRoleNames(member({ id: "x", roles: ["Member", "member"] }))).toEqual(["member"]);
  });
});

describe("filterMembers", () => {
  const base = { q: "", seat: "", status: "all" } as const;
  const seats = new Map<string, SeatType>([["1", "pro"], ["2", "free"], ["3", "free"]]);

  it("keeps everyone with no filters", () => {
    expect(filterMembers(roster, base)).toHaveLength(3);
  });

  it("filters by seat and drops members without seat data", () => {
    expect(filterMembers(roster, { ...base, seat: "free" }, seats).map((m) => m.id)).toEqual(["2", "3"]);
    expect(filterMembers(roster, { ...base, seat: "free" })).toEqual([]);
  });

  it("filters by status", () => {
    expect(filterMembers(roster, { ...base, status: "deactivated" }).map((m) => m.id)).toEqual(["3"]);
    expect(filterMembers(roster, { ...base, status: "active" }).map((m) => m.id)).toEqual(["1", "2"]);
  });

  it("searches identity and display name", () => {
    expect(filterMembers(roster, { ...base, q: "GRACE" }).map((m) => m.id)).toEqual(["2"]);
  });

  it("combines filters", () => {
    expect(filterMembers(roster, { q: "a", seat: "free", status: "active" }, seats).map((m) => m.id)).toEqual(["2"]);
  });
});

describe("users URL state", () => {
  it("defaults cleanly", () => {
    expect(parseUsersURLState(new URLSearchParams())).toEqual({ q: "", seat: "", status: "all", member: "" });
  });

  it("rejects an unknown status", () => {
    expect(parseUsersURLState(new URLSearchParams("status=banned")).status).toBe("all");
  });

  it("reads every filter", () => {
    expect(parseUsersURLState(new URLSearchParams("q=ada&seat=pro&status=active&member=u1"))).toEqual({
      q: "ada", seat: "pro", status: "active", member: "u1",
    });
  });

  it("drops defaults and empties when patching", () => {
    expect(patchUsersURL(new URLSearchParams("q=ada&status=active"), { status: "all", q: "" })).toBe("");
  });

  it("preserves unrelated params", () => {
    const next = patchUsersURL(new URLSearchParams("tab=members"), { seat: "pro" });
    expect(new URLSearchParams(next).get("tab")).toBe("members");
    expect(new URLSearchParams(next).get("seat")).toBe("pro");
  });
});
