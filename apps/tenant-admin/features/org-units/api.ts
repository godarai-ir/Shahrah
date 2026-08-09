import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const orgUnitSchema = z.object({
  id: z.string(),
  tenant_id: z.string(),
  alias: z.string(),
  display_name: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

const orgUnitMemberSchema = z.object({
  id: z.string(),
  tenant_id: z.string(),
  org_unit_id: z.string(),
  user_id: z.string(),
  role: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

const accessPrincipalSchema = z.object({
  type: z.enum(["user", "team"]),
  principal_id: z.string(),
  display_name: z.string(),
  secondary: z.string().optional(),
});

const teamSummarySchema = z.object({
  id: z.string(),
  alias: z.string(),
  display_name: z.string().optional(),
  workspace_type: z.string().optional(),
  org_unit_id: z.string().nullable().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});

const orgUnitListSchema = z.object({
  data: z.array(orgUnitSchema).optional(),
  total: z.number().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
  next_offset: z.number().optional(),
});
const memberListSchema = z.object({ data: z.array(orgUnitMemberSchema).optional() });
const userListSchema = z.object({ data: z.array(accessPrincipalSchema).optional() });
const teamListSchema = z.object({ data: z.array(teamSummarySchema).optional(), total: z.number().optional() });
const usageSummarySchema = z.object({
  data: z.object({ request_count: z.number().optional(), cost_irr: z.number().optional() }).optional(),
});
const entitlementsSchema = z.object({ features: z.record(z.string(), z.unknown()).optional() }).nullable();

export type OrgUnit = z.infer<typeof orgUnitSchema>;
export type OrgUnitMember = z.infer<typeof orgUnitMemberSchema>;
export type AccessPrincipalCandidate = z.infer<typeof accessPrincipalSchema>;
export type TeamSummary = z.infer<typeof teamSummarySchema>;
export type EffectiveEntitlements = z.infer<typeof entitlementsSchema>;

export function getOrgUnitEntitlements() {
  return requestParsedJSON(entitlementsSchema, "/api/entitlements");
}

export function getOrgUnitUsage(params: URLSearchParams, signal: AbortSignal) {
  return requestParsedJSON(usageSummarySchema, `/api/logs/summary?${params.toString()}`, { signal });
}

export function listOrgUnits(params: URLSearchParams) {
  return requestParsedJSON(orgUnitListSchema, `/api/org-units?${params.toString()}`);
}

export function getOrgUnit(id: string) {
  return requestParsedJSON(orgUnitSchema, `/api/org-units/${encodeURIComponent(id)}`);
}

export function listOrgUnitMembers(id: string) {
  return requestParsedJSON(memberListSchema, `/api/org-units/${encodeURIComponent(id)}/members`);
}

export function listOrgUnitUsers() {
  return requestParsedJSON(userListSchema, "/api/access-principals?kind=user&limit=100");
}

export function listOrgUnitTeams(limit: number, offset: number) {
  return requestParsedJSON(teamListSchema, `/api/teams?limit=${limit}&offset=${offset}`);
}

export function createOrgUnit(body: unknown) {
  return requestParsedJSON(orgUnitSchema, "/api/org-units", { method: "POST", body: JSON.stringify(body) });
}

export function updateOrgUnit(id: string, body: unknown) {
  return requestParsedJSON(orgUnitSchema, `/api/org-units/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}

export async function addOrgUnitMember(id: string, userID: string, role: string): Promise<void> {
  await requestJSON(`/api/org-units/${encodeURIComponent(id)}/members`, {
    method: "POST",
    body: JSON.stringify({ user_id: userID, role }),
  });
}

export async function updateOrgUnitMember(id: string, userID: string, role: string): Promise<void> {
  await requestJSON(`/api/org-units/${encodeURIComponent(id)}/members/${encodeURIComponent(userID)}`, {
    method: "PATCH",
    body: JSON.stringify({ role }),
  });
}

export async function removeOrgUnitMember(id: string, userID: string): Promise<void> {
  await requestJSON(`/api/org-units/${encodeURIComponent(id)}/members/${encodeURIComponent(userID)}`, {
    method: "DELETE",
  });
}

export async function deleteOrgUnit(id: string): Promise<void> {
  await requestJSON(`/api/org-units/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export function parseOrgUnitListPayload(payload: unknown) {
  const parsed = orgUnitListSchema.safeParse(payload);
  return parsed.success ? parsed.data : { data: [], total: 0 };
}
