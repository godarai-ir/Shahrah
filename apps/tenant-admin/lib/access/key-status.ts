import type { TenantKey } from "./api";

export type KeyType = "pat" | "vat";
export type KeyState = "active" | "expired" | "revoked" | "blocked";

/**
 * The control plane derives key_type from team_id when the column is unset,
 * so mirror that rule rather than trusting the field to be present.
 */
export function resolveKeyType(key: TenantKey): KeyType {
  const declared = key.key_type?.trim().toLowerCase();
  if (declared === "pat" || declared === "vat") return declared;
  return key.team_id?.trim() || key.virtual_user_id?.trim() ? "vat" : "pat";
}

/** Revocation wins over expiry, which wins over a block; otherwise the key is live. */
export function resolveKeyState(key: TenantKey, now: Date = new Date()): KeyState {
  if (key.revoked_at?.trim()) return "revoked";
  if (isExpired(key.expires_at, now)) return "expired";
  if (key.blocked) return "blocked";
  return "active";
}

export function isExpired(expiresAt: string | undefined, now: Date = new Date()): boolean {
  const raw = expiresAt?.trim();
  if (!raw) return false;
  const expiry = new Date(raw);
  return !Number.isNaN(expiry.getTime()) && expiry.getTime() <= now.getTime();
}

/** Keys a member holds directly: PATs they created. VATs belong to a service, not a person. */
export function listKeysOwnedByMember(keys: readonly TenantKey[], userID: string): TenantKey[] {
  const owner = userID.trim();
  if (!owner) return [];
  return keys.filter((key) => resolveKeyType(key) === "pat" && key.created_by_user_id?.trim() === owner);
}

export function countActiveKeys(keys: readonly TenantKey[], now: Date = new Date()): number {
  return keys.filter((key) => resolveKeyState(key, now) === "active").length;
}

/** Most recent use across a set of keys, or null when none has ever been used. */
export function resolveLastKeyUse(keys: readonly TenantKey[]): string | null {
  const stamps = keys
    .map((key) => key.last_used_at?.trim())
    .filter((value): value is string => Boolean(value))
    .map((value) => ({ value, time: new Date(value).getTime() }))
    .filter((entry) => !Number.isNaN(entry.time));
  if (stamps.length === 0) return null;
  return stamps.reduce((latest, entry) => (entry.time > latest.time ? entry : latest)).value;
}

export function sumKeySpend(keys: readonly TenantKey[]): number {
  return keys.reduce((total, key) => total + (key.spend ?? 0), 0);
}
