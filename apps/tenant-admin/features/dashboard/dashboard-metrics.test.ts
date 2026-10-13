import { describe, expect, it } from "vitest";

import {
  DAY_MS,
  HOUR_MS,
  summarizeDailyCost,
  summarizeErrorRate,
} from "./dashboard-metrics";

const now = new Date("2026-09-17T12:30:00Z");

function dailySeries(days: number, requestsPerDay: number) {
  const todayStart = Math.floor(now.getTime() / DAY_MS) * DAY_MS;
  return Array.from({ length: days }, (_, index) => ({
    timestamp: new Date(todayStart - (days - index) * DAY_MS).toISOString(),
    request_count: requestsPerDay,
    cost_usd: 2,
  }));
}

function hourlySeries(errorRates: number[]) {
  const lastCompleteStart = Math.floor(now.getTime() / HOUR_MS) * HOUR_MS - HOUR_MS;
  return errorRates.map((rate, index) => ({
    timestamp: new Date(lastCompleteStart - (errorRates.length - 1 - index) * HOUR_MS).toISOString(),
    request_count: 1_000,
    error_count: Math.round(rate * 1_000),
  }));
}

describe("summarizeErrorRate", () => {
  it("flags a spike above threshold and twice the baseline", () => {
    const summary = summarizeErrorRate(hourlySeries([...Array(47).fill(0.01), 0.12]), now);
    expect(summary?.spike).not.toBeNull();
    expect(summary?.sparkline).toHaveLength(9);
  });

  it("does not flag a high but steady rate", () => {
    expect(summarizeErrorRate(hourlySeries(Array(48).fill(0.08)), now)?.spike).toBeNull();
  });

  it("does not flag a jump that stays under the threshold", () => {
    expect(summarizeErrorRate(hourlySeries([...Array(47).fill(0.005), 0.04]), now)?.spike).toBeNull();
  });
});

describe("summarizeDailyCost", () => {
  it("compares today with the previous seven days", () => {
    expect(summarizeDailyCost(dailySeries(9, 1), now)).toMatchObject({ today: 0, average7d: 2 });
  });
});
