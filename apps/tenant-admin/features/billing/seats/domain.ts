import type { SeatSummary } from "./api";

/** Free seats left on the plan, or null when the plan has no cap. */
export function freeSeatsLeft(summary: Pick<SeatSummary, "free_limit" | "free_used">): number | null {
  if (summary.free_limit === null) return null;
  return Math.max(0, summary.free_limit - summary.free_used);
}

/**
 * Whether the renewal window of the running seat subscription is open at `now`
 * (epoch milliseconds): from the moment the control plane says it opens until
 * the period ends.
 */
export function renewalWindowOpen(summary: Pick<SeatSummary, "renewal_opens_at" | "period_end">, now: number): boolean {
  const opensAt = summary.renewal_opens_at ? Date.parse(summary.renewal_opens_at) : Number.NaN;
  const endsAt = summary.period_end ? Date.parse(summary.period_end) : Number.NaN;

  return Number.isFinite(opensAt) && Number.isFinite(endsAt) && now >= opensAt && now < endsAt;
}
