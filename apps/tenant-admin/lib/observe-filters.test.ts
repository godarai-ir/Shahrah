import { describe, expect, it } from "vitest";

import {
  EMPTY_OBSERVE_FILTERS,
  mergeObserveSearchParams,
  observeFiltersFromSearchParams,
  observeHref,
  observeQueryString,
  resolveObserveRange,
  unsupportedObserveFilters,
  type ObserveFilters,
} from "./observe-filters";

function parse(query: string) {
  return observeFiltersFromSearchParams(new URLSearchParams(query));
}

describe("observe filters URL round trip", () => {
  it("restores every shared filter from its own query string", () => {
    const filters: ObserveFilters = {
      q: "req_123",
      range: "custom",
      from: "2026-09-01T00:00:00.000Z",
      to: "2026-09-02T00:00:00.000Z",
      model: "gpt-4.1-mini",
      provider: "openai",
      status: "429",
      errorCode: "rate_limited",
    };

    expect(parse(observeQueryString(filters))).toEqual(filters);
  });

  it("writes nothing for the default state", () => {
    expect(observeQueryString(EMPTY_OBSERVE_FILTERS)).toBe("");
  });

  it("drops unknown parameters", () => {
    const query = observeQueryString(parse("model=m1&tracePrivacy=open&sort_by=duration_ms&foo=bar"));
    expect(query).toBe("model=m1");
  });

  it("rejects a status that is not a three-digit HTTP code", () => {
    expect(parse("status=abc").status).toBe("");
    expect(parse("status=5000").status).toBe("");
    expect(parse("status=503").status).toBe("503");
  });

  it("falls back to the default range when a custom range lacks bounds", () => {
    expect(parse("range=custom&from=2026-09-01T00:00:00Z").range).toBe("last24h");
  });

  it("reads a bare from/to pair from older metrics links as a custom range", () => {
    const filters = parse("from=2026-09-01T00:00:00Z&to=2026-09-01T01:00:00Z");
    expect(filters.range).toBe("custom");
    expect(filters.from).toBe("2026-09-01T00:00:00.000Z");
  });

  it("maps the legacy logs range values", () => {
    expect(parse("range=6h").range).toBe("last6h");
    expect(parse("range=7d").range).toBe("last7d");
  });

  it("ignores from/to under a relative range", () => {
    const filters = parse("range=last1h&from=2026-09-01T00:00:00Z&to=2026-09-01T01:00:00Z");
    expect(filters).toMatchObject({ range: "last1h", from: "", to: "" });
  });
});

describe("observeHref", () => {
  it("keeps a relative range when switching tabs", () => {
    expect(observeHref("traces", { ...EMPTY_OBSERVE_FILTERS, range: "last7d", model: "m1" })).toBe("/traces?range=last7d&model=m1");
  });

  it("keeps a custom range with its bounds", () => {
    const href = observeHref("metrics", {
      ...EMPTY_OBSERVE_FILTERS,
      range: "custom",
      from: "2026-09-01T00:00:00.000Z",
      to: "2026-09-02T00:00:00.000Z",
    });
    const params = new URL(href, "http://localhost").searchParams;
    expect(params.get("range")).toBe("custom");
    expect(params.get("from")).toBe("2026-09-01T00:00:00.000Z");
    expect(params.get("to")).toBe("2026-09-02T00:00:00.000Z");
  });

  it("returns the bare route without filters and appends view-specific extras", () => {
    expect(observeHref("logs", EMPTY_OBSERVE_FILTERS)).toBe("/logs");
    expect(observeHref("traces", EMPTY_OBSERVE_FILTERS, { trace: "tr_1" })).toBe("/traces?trace=tr_1");
  });
});

describe("mergeObserveSearchParams", () => {
  it("replaces shared parameters and keeps the view's own ones", () => {
    const merged = mergeObserveSearchParams(new URLSearchParams("trace=tr_1&model=old&range=last1h"), { ...EMPTY_OBSERVE_FILTERS, model: "new" });
    expect(merged.toString()).toBe("trace=tr_1&model=new");
  });
});

describe("filter support", () => {
  it("reports an error code as not applied on logs and metrics", () => {
    const filters = { ...EMPTY_OBSERVE_FILTERS, errorCode: "timeout", model: "m1" };
    expect(unsupportedObserveFilters("logs", filters)).toEqual(["errorCode"]);
    expect(unsupportedObserveFilters("metrics", filters)).toEqual(["errorCode"]);
    expect(unsupportedObserveFilters("traces", filters)).toEqual([]);
  });
});

describe("resolveObserveRange", () => {
  it("anchors relative presets at now", () => {
    const now = new Date("2026-09-19T12:00:00.000Z");
    expect(resolveObserveRange({ range: "last1h", from: "", to: "" }, now)).toEqual({
      from: "2026-09-19T11:00:00.000Z",
      to: "2026-09-19T12:00:00.000Z",
    });
  });
});
