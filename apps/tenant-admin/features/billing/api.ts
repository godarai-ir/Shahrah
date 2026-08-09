import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";
import {
  buildBillingMonthlyPath,
  buildBillingSummaryPath,
  type BillingSummaryPreset,
  type JalaliMonthSelection,
} from "./domain";

const providerUsageSchema = z.object({
  provider_name: z.string(), request_count: z.number(), input_tokens: z.number(), output_tokens: z.number(),
  raw_cost_irr: z.number(), discount_irr: z.number(), cache_savings_irr: z.number(), final_cost_irr: z.number(),
  share: z.number(),
});
export const billingSummarySchema = z.object({
  calendar: z.literal("jalali"),
  preset: z.enum(["current_month", "current_quarter", "current_year", "last_3_months", "last_6_months", "last_12_months"]),
  start: z.string(), end: z.string(), total_usage_irr: z.number(), discount_irr: z.number(),
  cache_savings_irr: z.number(), gift_credit_irr: z.number(), providers: z.array(providerUsageSchema),
});
export const monthlyUsageSchema = z.object({
  calendar: z.literal("jalali"), jalali_year: z.number(), jalali_month: z.number(), start: z.string(), end: z.string(),
  total_usage_irr: z.number(), gift_credit_irr: z.number(), discount_irr: z.number(), cache_savings_irr: z.number(),
  providers: z.array(providerUsageSchema),
});
export const bankTransferSchema = z.object({
  id: z.string(), public_id: z.string().optional(), amount_irr: z.number(), tracking_number: z.string(),
  bank_transaction_at: z.string(), receipt_file_url: z.string().optional(), status: z.string(),
  rejection_reason: z.string().optional(), created_at: z.string(),
});
export const billingProfileSchema = z.object({
  id: z.string().optional(), profile_type: z.string(), data: z.record(z.string(), z.unknown()),
});
export const invoiceSchema = z.object({
  id: z.string(), invoice_number: z.string(), status: z.string(), period_start: z.string(), period_end: z.string(),
  raw_usage_irr: z.number(), discount_irr: z.number(), gift_credit_irr: z.number(), cache_savings_irr: z.number(),
  final_amount_irr: z.number(), net_adjustment_irr: z.number(), created_at: z.string(),
});
export const walletTransactionSchema = z.object({
  id: z.string(), type: z.string(), amount_irr: z.number(), balance_after_irr: z.number().optional(),
  status: z.string(), created_at: z.string(),
});
export const discountRedemptionSchema = z.object({
  id: z.string(), code: z.string(), discount_type: z.string(), value_irr: z.number().optional(),
  percent: z.number().optional(), status: z.string(), used_amount_irr: z.number(), created_at: z.string(),
});
export const walletSchema = z.object({
  id: z.string().optional(), wallet_id: z.string().optional(), tenant_id: z.string().optional(),
  balance_irr: z.number().optional(), ledger_balance_irr: z.number().optional(), available_balance_irr: z.number().optional(),
  reserved_irr: z.number().optional(), committed_unsettled_irr: z.number().optional(), pending_irr: z.number().optional(),
  sync_lag_seconds: z.number().optional(), sync_status: z.string().optional(), window_id: z.string().optional(), as_of: z.string().optional(),
});

const currentExchangeRateSchema = z.object({
  base: z.literal("USD"), quote: z.literal("IRR"), rate: z.number().positive(),
  source: z.string().refine((value) => value.trim().toLowerCase().startsWith("nobitex:")),
  fetchedAt: z.string().refine((value) => !Number.isNaN(new Date(value).getTime())),
  status: z.enum(["fresh", "stale"]),
});
const topUpIntentSchema = z.object({ payment_url: z.string().optional() });

export type CurrentExchangeRate = z.infer<typeof currentExchangeRateSchema>;

export function getCurrentExchangeRate(signal?: AbortSignal) {
  return requestParsedJSON(currentExchangeRateSchema, "/api/exchange-rates/current", { signal });
}

export function createTopUpIntent(input: {
  amount_irr: number;
  provider: string;
  return_url: string;
  metadata: Record<string, string>;
}) {
  return requestParsedJSON(topUpIntentSchema, "/api/wallet/topup", {
    method: "POST", body: JSON.stringify(input),
  });
}

export async function submitBankTransfer(form: FormData) {
  await requestJSON("/api/wallet/bank-transfers", { method: "POST", body: form });
}

export async function getBillingReports(preset: BillingSummaryPreset, month: JalaliMonthSelection) {
  const [summary, monthly] = await Promise.all([
    requestParsedJSON(billingSummarySchema, buildBillingSummaryPath(preset)),
    requestParsedJSON(monthlyUsageSchema, buildBillingMonthlyPath(month)),
  ]);
  return { summary, monthly };
}

export async function updateBillingProfile(input: { profile_type: string; data: Record<string, string> }) {
  await requestJSON("/api/billing/profile", { method: "PUT", body: JSON.stringify(input) });
}

export function parseBillingSummary(payload: unknown) {
  return parseNullable(billingSummarySchema, payload);
}

export function parseMonthlyUsage(payload: unknown) {
  return parseNullable(monthlyUsageSchema, payload);
}

export function parseBillingProfile(payload: unknown) {
  return parseNullable(billingProfileSchema, payload);
}

export function parseWallet(payload: unknown) {
  return parseNullable(walletSchema, payload);
}

export function parseBankTransfers(payload: unknown) {
  return parseList(bankTransferSchema, payload);
}

export function parseInvoices(payload: unknown) {
  return parseList(invoiceSchema, payload);
}

export function parseWalletTransactions(payload: unknown) {
  return parseList(walletTransactionSchema, payload);
}

export function parseDiscountRedemptions(payload: unknown) {
  return parseList(discountRedemptionSchema, payload);
}

function parseNullable<T>(schema: z.ZodType<T>, payload: unknown): T | null {
  const parsed = schema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

function parseList<T>(itemSchema: z.ZodType<T>, payload: unknown): T[] {
  const listSchema = z.array(itemSchema);
  const direct = listSchema.safeParse(payload);
  if (direct.success) return direct.data;
  const enveloped = z.object({ data: listSchema }).safeParse(payload);
  return enveloped.success ? enveloped.data.data : [];
}
