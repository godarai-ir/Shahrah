import type { SeatSummary } from "./api";

/** Free seats left on the plan, or null when the plan has no cap. */
export function freeSeatsLeft(summary: Pick<SeatSummary, "free_limit" | "free_used">): number | null {
  if (summary.free_limit === null) return null;
  return Math.max(0, summary.free_limit - summary.free_used);
}
