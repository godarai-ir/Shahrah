import { describe, expect, it } from "vitest";

import { APIClientError } from "@/lib/http/api-client";

import { parseSeatSummary, purchaseRequiredQuote } from "./api";
import { freeSeatsLeft, personalUsagePercent, previewSeatChange, seatErrorMessage } from "./domain";

describe("previewSeatChange", () => {
  it.each([
    [{ seat: "unassigned", free_seat_used: false }, "free", { allowed: true, timing: "immediate", needsPaidSeat: false }],
    [{ seat: "unassigned", free_seat_used: true }, "free", { allowed: false, reason: "free_used" }],
    [{ seat: "free", free_seat_used: true }, "pro", { allowed: true, timing: "immediate", needsPaidSeat: true }],
    [{ seat: "pro", free_seat_used: true }, "max", { allowed: true, timing: "immediate", needsPaidSeat: true }],
    [{ seat: "max", free_seat_used: true }, "pro", { allowed: true, timing: "period_end", needsPaidSeat: false }],
    [{ seat: "pro", free_seat_used: true }, "unassigned", { allowed: true, timing: "period_end", needsPaidSeat: false }],
    [{ seat: "free", free_seat_used: true }, "unassigned", { allowed: true, timing: "immediate", needsPaidSeat: false }],
    [{ seat: "pro", free_seat_used: true }, "free", { allowed: false, reason: "back_to_free" }],
    [{ seat: "max", free_seat_used: true }, "max", { allowed: false, reason: "same" }],
  ] as const)("%o to %s", (member, target, expected) => {
    expect(previewSeatChange(member, target)).toEqual(expected);
  });
});

describe("seat helpers", () => {
  it("measures personal credit use", () => {
    expect(personalUsagePercent({ personal_granted_credits: "2000.00", personal_remaining_credits: "500.00" })).toBe(75);
    expect(personalUsagePercent({ personal_granted_credits: "0.00", personal_remaining_credits: "0.00" })).toBe(0);
  });

  it("counts free seats left", () => {
    expect(freeSeatsLeft({ free_limit: 10, free_used: 7 })).toBe(3);
    expect(freeSeatsLeft({ free_limit: 3, free_used: 5 })).toBe(0);
    expect(freeSeatsLeft({ free_limit: null, free_used: 5 })).toBeNull();
  });

  it("translates known error codes and keeps the server text otherwise", () => {
    const t = (key: string) => `t:${key}`;
    expect(seatErrorMessage(t, new APIClientError("x", { status: 409, code: "free_seat_used", payload: {} }))).toBe("t:errors.free_seat_used");
    expect(seatErrorMessage(t, new APIClientError("Server said no", { status: 500, code: "other", payload: {} }))).toBe("Server said no");
  });
});

describe("seat API contracts", () => {
  it("reads a purchase-required quote from a 409 body", () => {
    const quote = {
      seat: "pro", quantity: 1, billing_interval: "monthly", period_start: "2026-10-01T00:00:00Z", period_end: "2026-11-01T00:00:00Z",
      remaining_days: 12, period_days: 31, list_price_irr_per_credit: 10000, unit_period_price_irr: 18000000,
      price_irr: 6967742, tax_percent: "10.00", tax_irr: 696774, payable_irr: 7664516,
    };
    expect(purchaseRequiredQuote({ error: { code: "seat_purchase_required", quote } })?.payable_irr).toBe(7664516);
    expect(purchaseRequiredQuote({ error: { code: "seat_purchase_required" } })).toBeNull();
    expect(purchaseRequiredQuote({ error: { code: "free_seats_full" } })).toBeUndefined();
  });

  it("parses a seat summary", () => {
    const summary = parseSeatSummary({
      enforced: true, subscription_status: "active",
      billing_interval: "monthly", period_start: "2026-10-01T00:00:00Z", period_end: "2026-11-01T00:00:00Z",
      free_limit: 10, free_used: 4, unassigned: 1, pending_invitations: 0,
      paid: [{ type: "pro", purchased: 2, assigned: 1, available: 1, renewing: 2 }],
      offers: [], pending_requests: 1, tax_percent: "10.00",
    });
    expect(summary?.paid[0]?.available).toBe(1);
  });
});
