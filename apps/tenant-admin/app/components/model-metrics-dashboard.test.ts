import { describe, expect, it } from "vitest";

import { buildChartData, normalizeStringList } from "./model-metrics-dashboard";

describe("buildChartData", () => {
  it("uses raw request counts for the request chart", () => {
    const data = buildChartData(
      [
        {
          timestamp: "2026-06-21T08:30:00.000Z",
          modelName: "openai/gpt-4.1-mini",
          count_model_name: 7,
          p99_latency_ms: 1200,
        },
      ],
      "modelName",
      "p99",
      "Unknown",
    );

    expect(data.points).toHaveLength(1);
    expect(data.points[0].series0).toBe(7);
  });

  it("includes dates in axis labels when the selected range spans multiple days", () => {
    const data = buildChartData(
      [
        {
          timestamp: "2026-06-21T08:30:00.000Z",
          modelName: "openai/gpt-4.1-mini",
          count_model_name: 3,
        },
      ],
      "modelName",
      "p99",
      "Unknown",
      "en-US",
      "2026-01-20T08:32:00.000Z",
      "2026-06-21T08:32:00.000Z",
    );

    expect(data.points[0].label).toContain("Jun 21");
  });
});

describe("metrics filter option helpers", () => {
  it("dedupes and sorts historical facet values", () => {
    expect(normalizeStringList(["openai:gpt-4o", " anthropic/claude-3-5 ", "openai:gpt-4o", 42, ""])).toEqual([
      "anthropic/claude-3-5",
      "openai:gpt-4o",
    ]);
  });
});
