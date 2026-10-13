import { z } from "zod";

import { requestParsedJSON } from "@/lib/http/api-client";

/** A key as the tenant-wide /api/v1/keys listing returns it. */
const tenantKeySchema = z.object({
  id: z.string(),
  key_alias: z.string(),
  key_name: z.string().optional(),
  key_type: z.string().optional(),
  team_id: z.string().optional(),
  org_unit_id: z.string().optional(),
  virtual_user_id: z.string().optional(),
  created_by_user_id: z.string().optional(),
  spend: z.number().optional(),
  blocked: z.boolean().optional(),
  created_at: z.string().optional(),
  expires_at: z.string().optional(),
  last_used_at: z.string().optional(),
  revoked_at: z.string().optional(),
});

const tenantKeyListSchema = z.object({ data: z.array(tenantKeySchema).optional() });

export type TenantKey = z.infer<typeof tenantKeySchema>;

export function listTenantKeys() {
  return requestParsedJSON(tenantKeyListSchema, "/api/access-keys");
}

export function parseTenantKeyListPayload(payload: unknown) {
  const parsed = tenantKeyListSchema.safeParse(payload);
  return parsed.success ? parsed.data : { data: [] };
}
