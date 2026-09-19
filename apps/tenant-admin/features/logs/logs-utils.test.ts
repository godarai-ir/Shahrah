import { describe, expect, it } from "vitest";

import { EMPTY_OBSERVE_FILTERS } from "@/lib/observe-filters";

import { DEFAULT_FILTERS, type Filters } from "./logs-types";
import { filtersFromSearchParams, isSameFilterSet, logsFiltersFromObserve, normalizeFilterSet, observeFromLogsFilters } from "./logs-utils";

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

  it("matches saved views that stored the legacy range values", () => {
    expect(isSameFilterSet(filters({ range: "last7d" }), { range: "7d" })).toBe(true);
    expect(isSameFilterSet(filters(), { range: "24h" })).toBe(true);
  });

  it("drops keys that are not filters", () => {
    expect(normalizeFilterSet({ provider: "openai", limit: "25" } as never)).toEqual(normalizeFilterSet({ provider: "openai" }));
  });
});

describe("shared observe filters on logs", () => {
  it("maps the request ID lookup and status onto the logs filters and back", () => {
    const logs = filtersFromSearchParams(new URLSearchParams("q=req_1&status=502&model=gpt-5&range=last1h&error_code=timeout"));

    expect(logs).toMatchObject({ request_id: "req_1", status_code: "502", model: "gpt-5", range: "last1h" });
    expect(observeFromLogsFilters(logs, "timeout")).toEqual({
      q: "req_1",
      range: "last1h",
      from: "",
      to: "",
      model: "gpt-5",
      provider: "",
      status: "502",
      errorCode: "timeout",
    });
  });

  it("keeps logs-only filters when the shared ones change", () => {
    const next = logsFiltersFromObserve(filters({ route: "/v1/chat", model: "old" }), { ...EMPTY_OBSERVE_FILTERS, model: "new" });
    expect(next).toMatchObject({ route: "/v1/chat", model: "new" });
  });
});
