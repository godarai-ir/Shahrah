import { describe, expect, it } from "vitest";

import type { TraceSpan } from "./detail-contracts";
import { buildWaterfallRows, sortSpans } from "./trace-waterfall";

function span(id: string, startedAt: string, durationMS: number): TraceSpan {
  return { span_id: id, trace_id: "tr_1", name: id, kind: "internal", started_at: startedAt, duration_ms: durationMS, status: "ok", attributes: {} };
}

describe("buildWaterfallRows", () => {
  it("returns an empty array for no spans instead of NaN rows", () => {
    expect(buildWaterfallRows([])).toEqual([]);
  });

  it("offsets spans from the earliest start and gives concurrent spans the same offset", () => {
    const rows = buildWaterfallRows(sortSpans([
      span("provider", "2026-09-19T10:00:00.100Z", 300),
      span("root", "2026-09-19T10:00:00.000Z", 400),
      span("policy", "2026-09-19T10:00:00.100Z", 50),
    ]));

    expect(rows.map((row) => row.span.span_id)).toEqual(["root", "provider", "policy"]);
    expect(rows.map((row) => row.offsetMS)).toEqual([0, 100, 100]);
    expect(rows[1].leftPct).toBe(rows[2].leftPct);
    expect(rows[0]).toMatchObject({ totalMS: 400, leftPct: 0, widthPct: 100 });
    expect(rows[1].leftPct).toBeCloseTo(25);
    expect(rows[1].widthPct).toBeCloseTo(75);
  });

  it("keeps zero-length spans visible without a zero-length trace dividing by zero", () => {
    const rows = buildWaterfallRows([span("a", "2026-09-19T10:00:00.000Z", 0), span("b", "2026-09-19T10:00:00.000Z", 0)]);

    expect(rows).toHaveLength(2);
    for (const row of rows) {
      expect(row.totalMS).toBe(1);
      expect(Number.isFinite(row.leftPct)).toBe(true);
      expect(row.widthPct).toBeGreaterThan(0);
      expect(row.leftPct + row.widthPct).toBeLessThanOrEqual(100);
    }
  });

  it("places a span with an unreadable start at the origin", () => {
    const rows = buildWaterfallRows([span("ok", "2026-09-19T10:00:00.000Z", 100), span("bad", "not-a-date", 10)]);
    expect(rows.find((row) => row.span.span_id === "bad")).toMatchObject({ offsetMS: 0, leftPct: 0 });
  });
});

describe("buildWaterfallRows at the end of the window", () => {
  it("keeps a zero-length span at the end visible", () => {
    const rows = buildWaterfallRows([span("root", "2026-09-19T10:00:00.000Z", 400), span("end", "2026-09-19T10:00:00.400Z", 0)]);
    const end = rows.find((row) => row.span.span_id === "end")!;
    expect(end.widthPct).toBeGreaterThanOrEqual(1.5);
    expect(end.leftPct + end.widthPct).toBeLessThanOrEqual(100);
  });
});
