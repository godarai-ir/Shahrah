import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const tenantAdminUserSchema = z.object({
  id: z.string(),
  tenant_slug: z.string(),
  email: z.string().optional(),
  mobile: z.string().optional(),
  role: z.string(),
  roles: z.array(z.string()).optional(),
  active: z.boolean(),
  last_login_at: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

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
});

const roleListSchema = z.object({ data: z.array(roleDefinitionSchema).optional() });
const effectiveEntitlementsSchema = z.object({
  features: z.record(z.string(), z.unknown()).optional(),
}).nullable();

export type TenantAdminUser = z.infer<typeof tenantAdminUserSchema>;
export type RoleDefinition = z.infer<typeof roleDefinitionSchema>;
export type EffectiveEntitlements = z.infer<typeof effectiveEntitlementsSchema>;

export function getUserEntitlements() {
  return requestParsedJSON(effectiveEntitlementsSchema, "/api/entitlements");
}

export function listUsers(params: URLSearchParams) {
  return requestParsedJSON(userListSchema, `/api/users?${params.toString()}`);
}

export function listUserRoles() {
  return requestParsedJSON(roleListSchema, "/api/roles?limit=200&offset=0");
}

export function inviteUser(body: { email: string; role: string }) {
  return requestParsedJSON(inviteResponseSchema, "/api/users/invite", {
    method: "POST",
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
