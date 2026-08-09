import { z } from "zod";

export const roleDefinitionSchema = z.object({
  id: z.string(),
  tenant_id: z.string(),
  name: z.string(),
  display_name: z.string(),
  description: z.string().optional(),
  is_system: z.boolean(),
  permissions: z.array(z.string()),
  created_at: z.string(),
  updated_at: z.string(),
});

export const roleListSchema = z.object({ data: z.array(roleDefinitionSchema).optional() });

export const permissionItemSchema = z.object({
  key: z.string(),
  group: z.string(),
  description: z.string().optional().transform((value) => value ?? ""),
});

export const mePermissionsResponseSchema = z.object({
  catalog: z.array(permissionItemSchema).optional(),
});

export const effectiveEntitlementsSchema = z.object({
  features: z.record(z.string(), z.unknown()).optional(),
}).nullable();

export type RoleDefinition = z.infer<typeof roleDefinitionSchema>;
export type PermissionItem = z.infer<typeof permissionItemSchema>;
export type EffectiveEntitlements = z.infer<typeof effectiveEntitlementsSchema>;
