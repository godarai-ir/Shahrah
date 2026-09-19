import { describe, expect, it } from "vitest";

import { DEFAULT_FILTERS, type Filters } from "./logs-types";
import { isSameFilterSet, normalizeFilterSet } from "./logs-utils";

function filters(overrides: Partial<Filters> = {}): Filters {
  return { ...DEFAULT_FILTERS, ...overrides };
}

describe("isSameFilterSet", () => {
  it("ignores key order", () => {
    const a = { provider: "openai", model: "gpt-5", range: "7d" };
    const b = { range: "7d", model: "gpt-5", provider: "openai" };

    expect(isSameFilterSet(a, b)).toBe(true);
  });

  it("treats missing, empty, null and whitespace values alike", () => {
    expect(isSameFilterSet(filters({ provider: "openai" }), { provider: "openai", model: "", route: null, team_id: "  " })).toBe(true);
  });

  it("fills blanks with defaults", () => {
    expect(isSameFilterSet(filters(), {})).toBe(true);
    expect(isSameFilterSet(filters(), { range: "", sort_order: "" })).toBe(true);
    expect(isSameFilterSet(filters(), { range: "7d" })).toBe(false);
  });

  it("detects a real difference", () => {
    expect(isSameFilterSet(filters({ provider: "openai" }), filters({ provider: "anthropic" }))).toBe(false);
    expect(isSameFilterSet(filters({ provider: "openai" }), filters())).toBe(false);
  });

  it("ignores derived from/to on relative ranges but not on custom ones", () => {
    const stale = { from: "2026-01-01T00:00:00.000Z", to: "2026-01-02T00:00:00.000Z" };

    expect(isSameFilterSet(filters(), stale)).toBe(true);
    expect(isSameFilterSet(filters({ range: "custom", ...stale }), { range: "custom", ...stale })).toBe(true);
    expect(isSameFilterSet(filters({ range: "custom", ...stale }), { range: "custom", from: stale.from })).toBe(false);
  });

  it("drops keys that are not filters", () => {
    expect(normalizeFilterSet({ provider: "openai", limit: "25" } as never)).toEqual(normalizeFilterSet({ provider: "openai" }));
  });
});
