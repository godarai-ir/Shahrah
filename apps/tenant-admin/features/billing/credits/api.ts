import { z } from "zod";

import { requestParsedJSON } from "@/lib/http/api-client";

import { creditUsageSummarySchema } from "../api";

const creditAmount = z.string();

export const creditSettingsSchema = z.object({
  billing_unit: z.string(),
  overage_allowed: z.boolean(),
  overage_enabled: z.boolean(),
  overage_cap_credits: creditAmount.nullable(),
  /** Seat settings; absent on control planes without member seats. */
  seat_upgrade_requests_enabled: z.boolean().optional(),
  seat_pool_limit_pro_credits: creditAmount.nullish(),
  seat_pool_limit_max_credits: creditAmount.nullish(),
});

const creditLotSchema = z.object({
  lot_id: z.string(),
  source: z.string(),
  amount_credits: creditAmount,
  remaining_credits: creditAmount,
  expires_at: z.string().nullish(),
  created_at: z.string(),
});

export const creditForecastSchema = z.object({
  status: z.enum(["ok", "uncertain", "no_usage"]),
  window_days: z.number(),
  consumed_credits: creditAmount,
  daily_average_credits: creditAmount,
  depletes_at: z.string().nullish(),
  /** "expiry" when the credits expire before spending would use them up. */
  limited_by: z.enum(["usage", "expiry"]).optional(),
});

export const creditBalanceSchema = z.object({
  available_credits: creditAmount,
  debt_credits: creditAmount,
  lots: z.array(creditLotSchema),
  forecast: creditForecastSchema,
  settings: creditSettingsSchema,
});

export const creditPackSchema = z.object({
  slug: z.string(),
  version: z.number(),
  credits: z.number(),
  discount_percent: z.string(),
  validity_months: z.number(),
  price_irr: z.number(),
  tax_percent: z.string(),
  tax_irr: z.number(),
  payable_irr: z.number(),
});

export const creditPackOrderSchema = z.object({
  id: z.string(),
  pack_slug: z.string(),
  credits: z.number(),
  validity_months: z.number(),
  payable_irr: z.number(),
  provider: z.string(),
  status: z.string(),
  payment_url: z.string().optional(),
  failure_reason: z.string().optional(),
  created_at: z.string(),
  completed_at: z.string().nullish(),
});

const packListSchema = z.object({ data: z.array(creditPackSchema) });
const orderListSchema = z.object({ data: z.array(creditPackOrderSchema) });

export type CreditSettings = z.infer<typeof creditSettingsSchema>;
export type CreditLot = z.infer<typeof creditLotSchema>;
export type CreditForecast = z.infer<typeof creditForecastSchema>;
export type CreditBalance = z.infer<typeof creditBalanceSchema>;
export type CreditPack = z.infer<typeof creditPackSchema>;
export type CreditPackOrder = z.infer<typeof creditPackOrderSchema>;

/** Report windows the credit usage summary accepts. */
export const CREDIT_USAGE_PRESETS = ["current_month", "last_3_months", "last_6_months", "last_12_months", "current_year"] as const;
export type CreditUsagePreset = (typeof CREDIT_USAGE_PRESETS)[number];

/** The pack bank-transfer provider; such an order waits for a reviewed transfer instead of a gateway. */
export const BANK_TRANSFER_PROVIDER = "bank_transfer";

export function getCreditUsageSummary(preset: CreditUsagePreset, signal?: AbortSignal) {
  return requestParsedJSON(creditUsageSummarySchema, `/api/billing/credits/summary?preset=${encodeURIComponent(preset)}`, { signal });
}

export async function listCreditPackOrders() {
  return (await requestParsedJSON(orderListSchema, "/api/billing/credits/orders?limit=20")).data;
}

export function createCreditPackOrder(input: { pack_slug: string; provider: string; return_url: string }) {
  return requestParsedJSON(creditPackOrderSchema, "/api/billing/credits/orders", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

/** An empty limit string removes it. Overage is changed only by support. */
export function updateCreditSettings(input: {
  seat_upgrade_requests_enabled?: boolean;
  seat_pool_limit_pro_credits?: string;
  seat_pool_limit_max_credits?: string;
}) {
  return requestParsedJSON(creditSettingsSchema, "/api/billing/credits/settings", {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function parseCreditBalance(payload: unknown): CreditBalance | null {
  const parsed = creditBalanceSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

export function parseCreditSettings(payload: unknown): CreditSettings | null {
  const parsed = creditSettingsSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

export function parseCreditPacks(payload: unknown): CreditPack[] {
  const parsed = packListSchema.safeParse(payload);
  return parsed.success ? parsed.data.data : [];
}

export function parseCreditPackOrders(payload: unknown): CreditPackOrder[] {
  const parsed = orderListSchema.safeParse(payload);
  return parsed.success ? parsed.data.data : [];
}
