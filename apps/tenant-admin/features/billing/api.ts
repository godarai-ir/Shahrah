import { z } from "zod";

import { requestJSON } from "@/lib/http/api-client";

export const billingProfileSchema = z.object({
  id: z.string().optional(), profile_type: z.string(), data: z.record(z.string(), z.unknown()),
});

const creditUsageItemSchema = z.object({
  key: z.string(), label: z.string().optional(), credits: z.string(), request_count: z.number(),
});
export const creditUsageSummarySchema = z.object({
  available: z.boolean(), preset: z.string(),
  period_start: z.string(), period_end: z.string(), total_credits: z.string(), cache_saved_credits: z.string(),
  request_count: z.number(), by_model: z.array(creditUsageItemSchema), by_team: z.array(creditUsageItemSchema),
  by_user: z.array(creditUsageItemSchema).default([]), by_product: z.array(creditUsageItemSchema),
});
export type CreditUsageSummary = z.infer<typeof creditUsageSummarySchema>;
export type CreditUsageItem = z.infer<typeof creditUsageItemSchema>;

/** Submits a bank transfer receipt; with credit_pack_order_id it pays that pack order. */
export async function submitBankTransfer(form: FormData) {
  await requestJSON("/api/wallet/bank-transfers", { method: "POST", body: form });
}

export async function updateBillingProfile(input: { profile_type: string; data: Record<string, string> }) {
  await requestJSON("/api/billing/profile", { method: "PUT", body: JSON.stringify(input) });
}

export function parseCreditUsageSummary(payload: unknown) {
  return parseNullable(creditUsageSummarySchema, payload);
}

export function parseBillingProfile(payload: unknown) {
  return parseNullable(billingProfileSchema, payload);
}

function parseNullable<T>(schema: z.ZodType<T>, payload: unknown): T | null {
  const parsed = schema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}
