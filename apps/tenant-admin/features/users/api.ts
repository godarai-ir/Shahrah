import { z } from "zod";

import { tenantAdminUserSchema } from "@/lib/contracts/tenant-users";
import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const userListSchema = z.object({
  data: z.array(tenantAdminUserSchema).optional(),
  total: z.number().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
  next_offset: z.number().optional(),
});

const inviteResponseSchema = z.object({
  link: z.string().optional(),
  invitation: z.object({
    email: z.string().optional(),
    role: z.string().optional(),
  }).optional(),
});

const roleDefinitionSchema = z.object({
  id: z.string(),
  name: z.string(),
  display_name: z.string(),
  description: z.string().optional(),
  is_system: z.boolean(),
  // Carried so the member drawer can show effective permissions without a second fetch.
  permissions: z.array(z.string()).optional(),
});

const roleListSchema = z.object({ data: z.array(roleDefinitionSchema).optional() });

const seatUsageSchema = z.object({
  used: z.number(),
  active_members: z.number().optional(),
  pending_invitations: z.number().optional(),
  limit: z.number().nullable(),
});

export type SeatUsage = z.infer<typeof seatUsageSchema>;
export type RoleDefinition = z.infer<typeof roleDefinitionSchema>;

export function getUserSeats() {
  return requestParsedJSON(seatUsageSchema, "/api/users/seats");
}

export function listUsers(params: URLSearchParams) {
  return requestParsedJSON(userListSchema, `/api/users?${params.toString()}`);
}

export function listUserRoles() {
  return requestParsedJSON(roleListSchema, "/api/roles?limit=200&offset=0");
}

export function inviteUser(body: { email: string; role: string; first_name?: string; last_name?: string }) {
  return requestParsedJSON(inviteResponseSchema, "/api/users/invite", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export async function updateUserProfile(
  userID: string,
  body: { first_name: string; last_name: string },
): Promise<void> {
  await requestJSON(`/api/users/${encodeURIComponent(userID)}/profile`, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}

export async function setUserActive(userID: string, active: boolean): Promise<void> {
  await requestJSON(`/api/users/${encodeURIComponent(userID)}/${active ? "activate" : "deactivate"}`, {
    method: "POST",
  });
}

export async function assignUserRole(email: string, role: string): Promise<void> {
  await requestJSON("/api/users/roles", {
    method: "PATCH",
    body: JSON.stringify({ email, roles: [role], resourceType: "tenant" }),
  });
}

export function parseUserListPayload(payload: unknown) {
  const parsed = userListSchema.safeParse(payload);
  return parsed.success ? parsed.data : { data: [], total: 0 };
}

/** Parses `/api/v1/users/seats`; a malformed payload hides the usage instead of guessing. */
export function parseSeatUsage(payload: unknown): SeatUsage | null {
  const parsed = seatUsageSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}
