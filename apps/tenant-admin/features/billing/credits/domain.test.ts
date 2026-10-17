import { describe, expect, it } from "vitest";

import type { CreditUsageSummary } from "../api";
import { creditShare, forecastDaysLeft, netCacheSavedCredits } from "./domain";

const NOW = new Date("2026-10-17T12:00:00Z");

function summary(overrides: Partial<CreditUsageSummary> = {}): CreditUsageSummary {
  return {
    available: true,
    preset: "current_month",
    period_start: "2026-10-01T00:00:00Z",
    period_end: "2026-11-01T00:00:00Z",
    total_credits: "100.00",
    cache_saved_credits: "12.00",
    request_count: 10,
    by_model: [],
    by_team: [],
    by_user: [],
    by_product: [],
    ...overrides,
  };
}

describe("forecastDaysLeft", () => {
  it("has no end without usage or expiry, and counts to expiry without usage", () => {
    const idle = { status: "no_usage" as const, window_days: 7, consumed_credits: "0.00", daily_average_credits: "0.00" };
    expect(forecastDaysLeft(idle, NOW)).toBeNull();
    expect(forecastDaysLeft({ ...idle, depletes_at: "2026-10-31T12:00:00Z", limited_by: "expiry" }, NOW)).toBe(14);
  });

  it("counts whole days left and stops at zero", () => {
    const base = { status: "ok" as const, window_days: 7, consumed_credits: "70.00", daily_average_credits: "10.00" };
    expect(forecastDaysLeft({ ...base, depletes_at: "2026-10-20T18:00:00Z" }, NOW)).toBe(3);
    expect(forecastDaysLeft({ ...base, depletes_at: "2026-10-16T00:00:00Z" }, NOW)).toBe(0);
  });
});

describe("netCacheSavedCredits", () => {
  it("subtracts the semantic cache charge and never goes negative", () => {
    expect(netCacheSavedCredits(summary({ by_product: [{ key: "semantic_cache", credits: "2.50", request_count: 4 }] }))).toBe(9.5);
    expect(netCacheSavedCredits(summary({ cache_saved_credits: "1.00", by_product: [{ key: "semantic_cache", credits: "2.00", request_count: 1 }] }))).toBe(0);
    expect(netCacheSavedCredits(summary())).toBe(12);
  });
});

describe("creditShare", () => {
  it("is a fraction of the total", () => {
    expect(creditShare("25.00", "100.00")).toBe(0.25);
    expect(creditShare("5.00", "0.00")).toBe(0);
  });
});
