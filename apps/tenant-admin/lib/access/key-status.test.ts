import { describe, expect, it } from "vitest";

import type { TenantKey } from "./api";
import {
  countActiveKeys,
  listKeysOwnedByMember,
  resolveKeyState,
  resolveKeyType,
  resolveLastKeyUse,
  sumKeySpend,
} from "./key-status";

const NOW = new Date("2026-09-18T00:00:00Z");

function key(overrides: Partial<TenantKey> & { id: string }): TenantKey {
  return { key_alias: "alias", ...overrides };
}

describe("resolveKeyType", () => {
  it("trusts an explicit key_type", () => {
    expect(resolveKeyType(key({ id: "1", key_type: "VAT" }))).toBe("vat");
  });

  it("derives vat from team_id when key_type is missing, as the control plane does", () => {
    expect(resolveKeyType(key({ id: "1", team_id: "team-1" }))).toBe("vat");
  });

  it("derives vat from virtual_user_id", () => {
    expect(resolveKeyType(key({ id: "1", virtual_user_id: "vu-1" }))).toBe("vat");
  });

  it("falls back to pat", () => {
    expect(resolveKeyType(key({ id: "1" }))).toBe("pat");
  });
});

describe("resolveKeyState", () => {
  it("prefers revoked over expired", () => {
    const revoked = key({ id: "1", revoked_at: "2026-01-01T00:00:00Z", expires_at: "2026-01-02T00:00:00Z" });
    expect(resolveKeyState(revoked, NOW)).toBe("revoked");
  });

  it("marks a past expiry expired", () => {
    expect(resolveKeyState(key({ id: "1", expires_at: "2026-09-17T00:00:00Z" }), NOW)).toBe("expired");
  });

  it("keeps a future expiry active", () => {
    expect(resolveKeyState(key({ id: "1", expires_at: "2026-09-19T00:00:00Z" }), NOW)).toBe("active");
  });

  it("treats a missing expiry as no expiry", () => {
    expect(resolveKeyState(key({ id: "1" }), NOW)).toBe("active");
  });

  it("ignores an unparseable expiry rather than calling the key expired", () => {
    expect(resolveKeyState(key({ id: "1", expires_at: "not-a-date" }), NOW)).toBe("active");
  });

  it("reports a blocked key", () => {
    expect(resolveKeyState(key({ id: "1", blocked: true }), NOW)).toBe("blocked");
  });
});

describe("member ownership", () => {
  const keys = [
    key({ id: "1", created_by_user_id: "u1", spend: 2.5, last_used_at: "2026-09-01T00:00:00Z" }),
    key({ id: "2", created_by_user_id: "u1", revoked_at: "2026-02-01T00:00:00Z", spend: 1 }),
    key({ id: "3", created_by_user_id: "u2" }),
    key({ id: "4", key_type: "vat", virtual_user_id: "vu-1", created_by_user_id: "u1" }),
  ];

  it("counts only the member's own PATs, not VATs they created", () => {
    expect(listKeysOwnedByMember(keys, "u1").map((item) => item.id)).toEqual(["1", "2"]);
  });

  it("returns nothing for a blank user id", () => {
    expect(listKeysOwnedByMember(keys, "  ")).toEqual([]);
  });

  it("counts active keys only", () => {
    expect(countActiveKeys(listKeysOwnedByMember(keys, "u1"), NOW)).toBe(1);
  });

  it("sums spend across the member's keys", () => {
    expect(sumKeySpend(listKeysOwnedByMember(keys, "u1"))).toBe(3.5);
  });
});

describe("resolveLastKeyUse", () => {
  it("returns the most recent use", () => {
    const keys = [
      key({ id: "1", last_used_at: "2026-09-01T00:00:00Z" }),
      key({ id: "2", last_used_at: "2026-09-10T00:00:00Z" }),
      key({ id: "3" }),
    ];
    expect(resolveLastKeyUse(keys)).toBe("2026-09-10T00:00:00Z");
  });

  it("returns null when nothing has been used", () => {
    expect(resolveLastKeyUse([key({ id: "1" })])).toBeNull();
  });
});
