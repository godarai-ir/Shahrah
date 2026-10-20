import { APIClientError } from "@/lib/http/api-client";

import type { SeatMember, SeatSummary, SeatType } from "./api";

/** Error codes the seat API returns; each has a translated message. */
export const SEAT_ERROR_CODES = [
  "invalid_seat_type",
  "invalid_seat_quantity",
  "seat_interval_not_supported",
  "seat_unchanged",
  "seat_back_to_free",
  "free_seat_used",
  "free_seats_full",
  "seats_need_active_subscription",
  "seat_purchase_required",
  "member_not_found",
  "no_pending_seat_change",
  "seat_request_not_found",
  "seat_request_closed",
  "seat_order_not_payable",
  "seat_renewal_not_due",
  "invalid_seat_renewal",
  "invalid_seat_pool_limit",
  "payment_provider_unavailable",
] as const;

type Translate = (key: string, values?: Record<string, string | number>) => string;

/** A translated message for a seat API error, or the server's own text. */
export function seatErrorMessage(t: Translate, error: unknown): string {
  if (error instanceof APIClientError && error.code && (SEAT_ERROR_CODES as readonly string[]).includes(error.code)) {
    return t(`errors.${error.code}`);
  }

  return error instanceof Error && error.message.trim() ? error.message : t("errors.unknown");
}

const SEAT_RANK: Record<SeatType, number> = { unassigned: 0, free: 1, pro: 2, max: 3 };

export type SeatChangePreview =
  | { allowed: true; timing: "immediate" | "period_end"; needsPaidSeat: boolean }
  | { allowed: false; reason: "same" | "back_to_free" | "free_used" };

/**
 * Mirrors the server's seat change rules so the dialog can explain a change
 * before it is made: upgrades are immediate, paid downgrades wait for the end
 * of the billing period, paid seats never go back to free, and the free seat
 * is given once.
 */
export function previewSeatChange(member: Pick<SeatMember, "seat" | "free_seat_used">, target: SeatType): SeatChangePreview {
  const from = member.seat;
  if (from === target) return { allowed: false, reason: "same" };
  if (target === "free" && (from === "pro" || from === "max")) return { allowed: false, reason: "back_to_free" };
  if (target === "free" && member.free_seat_used) return { allowed: false, reason: "free_used" };
  if (SEAT_RANK[target] > SEAT_RANK[from]) return { allowed: true, timing: "immediate", needsPaidSeat: target === "pro" || target === "max" };
  return { allowed: true, timing: from === "pro" || from === "max" ? "period_end" : "immediate", needsPaidSeat: false };
}

/** Share of a member's personal credits already spent, 0 to 100. */
export function personalUsagePercent(member: Pick<SeatMember, "personal_granted_credits" | "personal_remaining_credits">): number {
  const granted = Number(member.personal_granted_credits);
  const remaining = Number(member.personal_remaining_credits);
  if (!Number.isFinite(granted) || granted <= 0) return 0;
  return Math.min(100, Math.max(0, ((granted - remaining) / granted) * 100));
}

export const SEAT_TONE: Record<SeatType, "info" | "success" | "neutral" | "warning"> = {
  max: "success",
  pro: "info",
  free: "neutral",
  unassigned: "warning",
};

/** Free seats left on the plan, or null when the plan has no cap. */
export function freeSeatsLeft(summary: Pick<SeatSummary, "free_limit" | "free_used">): number | null {
  if (summary.free_limit === null) return null;
  return Math.max(0, summary.free_limit - summary.free_used);
}

/** True while seat limits are announced but not enforced yet. */
export function inGracePeriod(summary: Pick<SeatSummary, "enabled" | "enforced" | "enforced_from">, now = new Date()): boolean {
  if (!summary.enabled || summary.enforced || !summary.enforced_from) return false;
  return new Date(summary.enforced_from).getTime() > now.getTime();
}
