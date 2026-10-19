import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const teamSchema = z.object({
  id: z.string(),
  tenant_id: z.string(),
  alias: z.string(),
  display_name: z.string().optional(),
  workspace_type: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

const teamMemberSchema = z.object({
  id: z.string().optional(),
  user_id: z.string(),
  user_display_name: z.string().optional(),
  user_email: z.string().optional(),
  user_mobile: z.string().optional(),
  role: z.string(),
  team_member_permissions: z.array(z.string()).optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  updated_at: z.string().optional(),
});

const accessPrincipalSchema = z.object({
  principal_id: z.string(),
  display_name: z.string(),
  secondary: z.string().optional(),
});

const teamListSchema = z.object({ data: z.array(teamSchema).optional(), total: z.number().optional() });
const memberListSchema = z.object({ data: z.array(teamMemberSchema).optional(), total: z.number().optional() });
const principalListSchema = z.object({ data: z.array(accessPrincipalSchema).optional(), total: z.number().optional() });
const entitlementSchema = z.object({ features: z.record(z.string(), z.unknown()).optional() }).nullable();
const usageSummarySchema = z.object({
  data: z.object({ request_count: z.number().optional(), cost_irr: z.number().optional() }).optional(),
});
const costBreakdownEntrySchema = z.object({
  team_id: z.string().optional(),
  cost_irr: z.number().optional(),
});
const costBreakdownSchema = z.object({ data: z.array(costBreakdownEntrySchema).optional() });

export type Team = z.infer<typeof teamSchema>;
export type TeamMember = z.infer<typeof teamMemberSchema>;
export type AccessPrincipalCandidate = z.infer<typeof accessPrincipalSchema>;
export type EffectiveEntitlements = z.infer<typeof entitlementSchema>;
export type TeamPayload = {
  alias: string;
  display_name: string;
  workspace_type: string;
  metadata: Record<string, unknown>;
};
export type TeamMemberPayload = {
  user_id?: string;
  role: string;
  team_member_permissions: readonly string[];
  metadata?: Record<string, unknown>;
};

export function getTeamEntitlements() {
  return requestParsedJSON(entitlementSchema, "/api/entitlements");
}

export function getTeamUsage(params: URLSearchParams, signal: AbortSignal) {
  return requestParsedJSON(usageSummarySchema, `/api/logs/summary?${params.toString()}`, { signal });
}

/**
 * Per-team cost breakdown for the teams table's "cost this month" column.
 * group_by=team_id support on the control-plane analytics endpoint is unconfirmed;
 * callers must treat a failure or an empty/unparseable response as "not available yet", not zero.
 */
export function listTeamCostBreakdown(params: URLSearchParams, signal: AbortSignal) {
  return requestParsedJSON(costBreakdownSchema, `/api/logs/analytics?${params.toString()}`, { signal });
}

export function listTeams(params: URLSearchParams) {
  return requestParsedJSON(teamListSchema, `/api/teams?${params.toString()}`);
}

export function getTeam(id: string) {
  return requestParsedJSON(teamSchema, `/api/teams/${encodeURIComponent(id)}`);
}

export function listTeamMembers(id: string) {
  return requestParsedJSON(memberListSchema, `/api/teams/${encodeURIComponent(id)}/members`);
}

export function listTeamUserCandidates() {
  return requestParsedJSON(principalListSchema, "/api/access-principals?kind=user&limit=100");
}

export function createTeam(body: TeamPayload) {
  return requestParsedJSON(teamSchema, "/api/teams", { method: "POST", body: JSON.stringify(body) });
}

export function updateTeam(id: string, body: TeamPayload) {
  return requestParsedJSON(teamSchema, `/api/teams/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}

export async function addTeamMember(teamID: string, body: TeamMemberPayload): Promise<void> {
  await requestJSON(`/api/teams/${encodeURIComponent(teamID)}/members`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export async function updateTeamMember(teamID: string, userID: string, body: TeamMemberPayload): Promise<void> {
  await requestJSON(`/api/teams/${encodeURIComponent(teamID)}/members/${encodeURIComponent(userID)}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}

export async function removeTeamMember(teamID: string, userID: string): Promise<void> {
  await requestJSON(`/api/teams/${encodeURIComponent(teamID)}/members/${encodeURIComponent(userID)}`, { method: "DELETE" });
}

export async function deleteTeam(id: string): Promise<void> {
  await requestJSON(`/api/teams/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export function parseTeamListPayload(payload: unknown) {
  const parsed = teamListSchema.safeParse(payload);
  return parsed.success ? parsed.data : { data: [], total: 0 };
}
