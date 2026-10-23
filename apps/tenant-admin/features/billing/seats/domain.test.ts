import { describe, expect, it } from "vitest";

import { parseSeatSummary, purchaseRequiredQuote } from "./api";
import { freeSeatsLeft, renewalWindowOpen } from "./domain";

describe("seat helpers", () => {
  it("counts free seats left", () => {
    expect(freeSeatsLeft({ free_limit: 10, free_used: 7 })).toBe(3);
    expect(freeSeatsLeft({ free_limit: 3, free_used: 5 })).toBe(0);
    expect(freeSeatsLeft({ free_limit: null, free_used: 5 })).toBeNull();
  });
});

describe("renewal window", () => {
  const summary = { renewal_opens_at: "2026-10-25T00:00:00Z", period_end: "2026-11-01T00:00:00Z" };
  const at = (iso: string) => Date.parse(iso);

  it("is open from the opening moment until the period ends", () => {
    expect(renewalWindowOpen(summary, at("2026-10-24T23:59:59Z"))).toBe(false);
    expect(renewalWindowOpen(summary, at("2026-10-25T00:00:00Z"))).toBe(true);
    expect(renewalWindowOpen(summary, at("2026-10-31T23:59:59Z"))).toBe(true);
    expect(renewalWindowOpen(summary, at("2026-11-01T00:00:00Z"))).toBe(false);
  });

  it("is closed without a running subscription", () => {
    expect(renewalWindowOpen({ renewal_opens_at: null, period_end: null }, at("2026-10-28T00:00:00Z"))).toBe(false);
    expect(renewalWindowOpen({ period_end: "2026-11-01T00:00:00Z" }, at("2026-10-28T00:00:00Z"))).toBe(false);
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
