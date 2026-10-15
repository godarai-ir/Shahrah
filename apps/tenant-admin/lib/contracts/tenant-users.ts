import { z } from "zod";

/** A tenant member as returned by the control plane users API. */
export const tenantAdminUserSchema = z.object({
  id: z.string(),
  tenant_slug: z.string(),
  email: z.string().optional(),
  mobile: z.string().optional(),
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  role: z.string(),
  roles: z.array(z.string()).optional(),
  active: z.boolean(),
  last_login_at: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type TenantAdminUser = z.infer<typeof tenantAdminUserSchema>;
