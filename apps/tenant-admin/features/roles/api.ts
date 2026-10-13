import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";
import { effectiveEntitlementsSchema } from "@/lib/contracts/roles";

export type RolePayload = {
  name?: string;
  display_name: string;
  description: string;
  permissions: string[];
};

export function getRoleEntitlements() {
  return requestParsedJSON(effectiveEntitlementsSchema, "/api/entitlements");
}

export async function createRole(body: RolePayload & { name: string }): Promise<void> {
  await requestJSON("/api/roles", { method: "POST", body: JSON.stringify(body) });
}

export async function updateRole(id: string, body: RolePayload): Promise<void> {
  await requestJSON(`/api/roles/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}
