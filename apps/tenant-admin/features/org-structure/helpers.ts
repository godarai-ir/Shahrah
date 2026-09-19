import type { OrgUnit } from "@/lib/api/org-units";
import type { Team } from "@/lib/api/teams";

export const WORKSPACE_TYPES = ["production", "staging", "development", "research", "internal", "general"] as const;
const WORKSPACE_TYPE_SET = new Set<string>(WORKSPACE_TYPES);

export function getTeamName(team: Team): string {
  return team.display_name || toText(team.metadata?.display_name) || team.alias;
}

export function getOrgUnitName(orgUnit: OrgUnit): string {
  return orgUnit.display_name || toText(orgUnit.metadata?.display_name) || orgUnit.alias;
}

export function getWorkspaceType(team: Team): string {
  const value = (team.workspace_type || toText(team.metadata?.workspace_type) || "production").trim().toLowerCase();
  return WORKSPACE_TYPE_SET.has(value) ? value : "general";
}

export function getActiveKeys(team: Team): number {
  return toNumber(team.metadata?.active_keys_count ?? team.metadata?.active_key_count);
}

export function getVirtualUsers(team: Team): number {
  return toNumber(team.metadata?.virtual_users_count);
}

/** No first-class field exists yet for these; only positive, explicitly-named metadata counts are trusted. Never a guess. */
export function getMemberCount(team: Team): number | null {
  const metadata = team.metadata ?? {};
  return firstPositiveCountOrNull(metadata.member_count, metadata.members_count, metadata.memberCount, metadata.membersCount);
}

export function getAdminCount(team: Team): number | null {
  const metadata = team.metadata ?? {};
  return firstPositiveCountOrNull(metadata.admin_count, metadata.admins_count, metadata.adminCount, metadata.adminsCount);
}

function firstPositiveCountOrNull(...values: unknown[]): number | null {
  for (const value of values) {
    const count = toCount(value);
    if (count > 0) return count;
  }
  return null;
}

export function toCount(value: unknown): number {
  if (Array.isArray(value)) return value.length;
  return toNumber(value);
}

export function toNumber(value: unknown): number {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value.replace(/,/g, ""));
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export function toText(value: unknown): string {
  if (typeof value === "string") return value;
  if (typeof value === "number") return String(value);
  return "";
}

export function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}

export function normalizeError(error: unknown, fallback: string): string {
  if (error instanceof Error) return error.message;
  return fallback;
}

export type UsagePeriod = "current_month" | "last_30_days" | "last_7_days";
