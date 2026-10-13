import { z } from "zod";

export const tenantSubscriptionSchema = z.object({
  plan_slug: z.string().optional(),
  plan_display_name: z.string().optional(),
  status: z.string(),
  billing_interval: z.string(),
  current_period_start: z.string().optional(),
  current_period_end: z.string().optional(),
  trial_started_at: z.string().optional(),
  trial_ends_at: z.string().optional(),
});

export const effectiveEntitlementsSchema = z.object({
  subscription: tenantSubscriptionSchema.nullable().optional(),
  features: z.record(z.string(), z.unknown()).optional(),
  overrides: z.record(z.string(), z.unknown()).optional(),
});


export type TenantSubscription = z.infer<typeof tenantSubscriptionSchema>;
export type EffectiveEntitlements = z.infer<typeof effectiveEntitlementsSchema>;
