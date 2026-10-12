import { z } from "zod";

/** A virtual user (service identity) as returned by the control plane. */
export const virtualUserSchema = z.object({
  id: z.string(),
  tenant_id: z.string(),
  alias: z.string(),
  display_name: z.string().optional(),
  org_unit_id: z.string().optional(),
  team_id: z.string().optional(),
  active: z.boolean(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type VirtualUser = z.infer<typeof virtualUserSchema>;
