import { z } from "zod";

import { requestParsedJSON } from "@/lib/http/api-client";

import { paidSeatType, setMemberSeatResultSchema, type PaidSeatType } from "@/lib/seats";

export type { PaidSeatType } from "@/lib/seats";

export const seatOfferSchema = z.object({
  type: z.enum(["free", "pro", "max"]),
  monthly_credits: z.number(),
  lifetime: z.boolean(),
  pool_access: z.boolean(),
  discount_percent: z.string(),
  yearly_discount_percent: z.string().optional(),
  monthly_price_irr: z.number(),
  yearly_price_irr: z.number(),
  active: z.boolean(),
});

export const seatTypeCountSchema = z.object({
  type: paidSeatType,
  purchased: z.number(),
  assigned: z.number(),
  available: z.number(),
  renewing: z.number(),
});

export const seatRenewalLineSchema = z.object({
  block_id: z.string(),
  seat: paidSeatType,
  quantity: z.number(),
  paid_through: z.string(),
});

export const seatRenewalQuoteSchema = z.object({
  lines: z.array(seatRenewalLineSchema),
  period_end: z.string(),
  price_irr: z.number(),
  tax_percent: z.string(),
  tax_irr: z.number(),
  payable_irr: z.number(),
});

export const seatSummarySchema = z.object({
  enforced: z.boolean(),
  subscription_status: z.string().optional().default(""),
  /** "free" until the first paid seat purchase starts the "seats" subscription. */
  subscription_kind: z.enum(["free", "seats"]).optional().default("free"),
  billing_interval: z.string().optional(),
  period_start: z.string().nullish(),
  period_end: z.string().nullish(),
  free_limit: z.number().nullable(),
  free_used: z.number(),
  unassigned: z.number(),
  pending_invitations: z.number(),
  paid: z.array(seatTypeCountSchema),
  offers: z.array(seatOfferSchema),
  pending_requests: z.number(),
  renewal_due: seatRenewalQuoteSchema.nullish(),
  /** When the running subscription's renewal window opens. */
  renewal_opens_at: z.string().nullish(),
  tax_percent: z.string(),
});

export const seatRenewalPreviewMemberSchema = z.object({
  user_id: z.string(),
  email: z.string().optional().default(""),
  name: z.string().optional().default(""),
  assigned_at: z.string().nullish(),
  /** What the member holds after losing the paid seat. */
  fallback: z.enum(["free", "unassigned"]),
});

export const seatRenewalPreviewLineSchema = z.object({
  seat: paidSeatType,
  /** Seats of the type the tenant holds, and how many are already paid for the next period. */
  held: z.number(),
  renewed: z.number(),
  keep: z.number(),
  assigned: z.number(),
  unit_price_irr: z.number(),
  losing: z.array(seatRenewalPreviewMemberSchema).nullish().transform((value) => value ?? []),
});

export const seatRenewalPreviewSchema = z.object({
  period_end: z.string(),
  next_period_end: z.string(),
  lines: z.array(seatRenewalPreviewLineSchema),
  price_irr: z.number(),
  tax_percent: z.string(),
  tax_irr: z.number(),
  payable_irr: z.number(),
});

export const seatQuoteSchema = z.object({
  seat: paidSeatType,
  quantity: z.number(),
  billing_interval: z.string(),
  period_start: z.string(),
  period_end: z.string(),
  remaining_days: z.number(),
  period_days: z.number(),
  list_price_irr_per_credit: z.number(),
  unit_period_price_irr: z.number(),
  price_irr: z.number(),
  tax_percent: z.string(),
  tax_irr: z.number(),
  payable_irr: z.number(),
  /** Paying this quote starts the seat subscription; the period begins now. */
  starts_subscription: z.boolean().optional().default(false),
});

export const seatOrderSchema = z.object({
  id: z.string(),
  kind: z.enum(["purchase", "renewal"]),
  seat: paidSeatType,
  quantity: z.number(),
  remaining_days: z.number(),
  period_days: z.number(),
  payable_irr: z.number(),
  status: z.string(),
  payment_url: z.string().optional(),
  failure_reason: z.string().optional(),
  created_at: z.string(),
  completed_at: z.string().nullish(),
});

export const seatRequestSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  email: z.string().optional().default(""),
  name: z.string().optional().default(""),
  current_seat: z.string(),
  requested_seat: paidSeatType,
  reason: z.string().optional().default(""),
  block_reason: z.string().optional().default(""),
  status: z.enum(["pending", "approved", "denied", "canceled"]),
  decision_note: z.string().optional().default(""),
  decided_at: z.string().nullish(),
  created_at: z.string(),
});

const seatOrderListSchema = z.object({ data: z.array(seatOrderSchema) });
const seatRequestListSchema = z.object({ data: z.array(seatRequestSchema) });

/** The 409 body of a paid upgrade without a seat in stock. */
export const seatPurchaseRequiredSchema = z.object({
  error: z.object({ code: z.literal("seat_purchase_required"), quote: seatQuoteSchema.optional() }),
});

export type SeatOffer = z.infer<typeof seatOfferSchema>;
export type SeatTypeCount = z.infer<typeof seatTypeCountSchema>;
export type SeatRenewalQuote = z.infer<typeof seatRenewalQuoteSchema>;
export type SeatSummary = z.infer<typeof seatSummarySchema>;
export type SeatRenewalPreview = z.infer<typeof seatRenewalPreviewSchema>;
export type SeatRenewalPreviewLine = z.infer<typeof seatRenewalPreviewLineSchema>;
export type SeatRenewalPreviewMember = z.infer<typeof seatRenewalPreviewMemberSchema>;
export type SeatQuote = z.infer<typeof seatQuoteSchema>;
export type SeatOrder = z.infer<typeof seatOrderSchema>;
export type SeatRequest = z.infer<typeof seatRequestSchema>;

export function getSeatSummary() {
  return requestParsedJSON(seatSummarySchema, "/api/billing/seats");
}

export type SeatInterval = "monthly" | "yearly";

export function quoteSeats(input: { seat: PaidSeatType; quantity: number; billing_interval?: SeatInterval }) {
  return requestParsedJSON(seatQuoteSchema, "/api/billing/seats/quote", { method: "POST", body: JSON.stringify(input) });
}

export function updateSeatRenewal(input: { seat: PaidSeatType; quantity: number }) {
  return requestParsedJSON(seatSummarySchema, "/api/billing/seats/renewal", { method: "PUT", body: JSON.stringify(input) });
}

/** What renewing the given number of seats per type costs and which members lose their seat; a missing type uses the saved count. */
export function previewSeatRenewal(keep: Partial<Record<PaidSeatType, number>>, signal?: AbortSignal) {
  const query = new URLSearchParams();
  for (const [seat, count] of Object.entries(keep)) query.set(seat, String(count));

  return requestParsedJSON(seatRenewalPreviewSchema, `/api/billing/seats/renewal/preview?${query.toString()}`, { signal });
}

export function createSeatOrder(input: {
  kind: "purchase" | "renewal";
  seat?: PaidSeatType;
  quantity?: number;
  billing_interval?: SeatInterval;
  assign_user_id?: string;
  request_id?: string;
  return_url: string;
}) {
  return requestParsedJSON(seatOrderSchema, "/api/billing/seats/orders", { method: "POST", body: JSON.stringify({ provider: "", ...input }) });
}

export async function listSeatRequests() {
  return (await requestParsedJSON(seatRequestListSchema, "/api/billing/seats/requests?limit=100")).data;
}

export function approveSeatRequest(requestID: string) {
  return requestParsedJSON(setMemberSeatResultSchema, `/api/billing/seats/requests/${encodeURIComponent(requestID)}/approve`, { method: "POST" });
}

export function denySeatRequest(requestID: string, note: string) {
  return requestParsedJSON(seatRequestSchema, `/api/billing/seats/requests/${encodeURIComponent(requestID)}/deny`, {
    method: "POST",
    body: JSON.stringify({ note }),
  });
}

export function parseSeatSummary(payload: unknown): SeatSummary | null {
  const parsed = seatSummarySchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

export function parseSeatRenewalPreview(payload: unknown): SeatRenewalPreview | null {
  const parsed = seatRenewalPreviewSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

export function parseSeatRequests(payload: unknown): SeatRequest[] {
  const parsed = seatRequestListSchema.safeParse(payload);
  return parsed.success ? parsed.data.data : [];
}

export function parseSeatOrders(payload: unknown): SeatOrder[] {
  const parsed = seatOrderListSchema.safeParse(payload);
  return parsed.success ? parsed.data.data : [];
}

/** The quote attached to a seat_purchase_required error, if the error is one. */
export function purchaseRequiredQuote(payload: unknown): SeatQuote | null | undefined {
  const parsed = seatPurchaseRequiredSchema.safeParse(payload);
  if (!parsed.success) return undefined;
  return parsed.data.error.quote ?? null;
}
