import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { createTranslator } from "next-intl";
import { afterEach, describe, expect, it } from "vitest";

import messages from "@/messages/modules/logs/en.json";

import { createLogFormatters } from "./logs-utils";
import type { Summary } from "./logs-types";
import { RANGE_SUMMARY_STORAGE_KEY, RangeSummary } from "./RangeSummary";

const t = createTranslator({ locale: "en", messages: { logs: messages }, namespace: "logs" }) as never;
const formatters = createLogFormatters("en");

const summary: Summary = {
  request_count: 1_234,
  error_rate: 0.018,
  latency_p95_ms: 1_800,
  cost_usd: 12.5,
  cache_hit_rate: 0.129,
};

function renderSummary(props: { summaryUnavailable?: boolean } = {}) {
  return render(
    <RangeSummary summary={summary} summaryUnavailable={props.summaryUnavailable ?? false} formatters={formatters} t={t}>
      <div data-testid="charts" />
    </RangeSummary>,
  );
}

function summaryLine(container: HTMLElement) {
  const element = container.querySelector("summary");
  if (!element) throw new Error("summary element missing");
  return element;
}

describe("RangeSummary", () => {
  afterEach(() => window.localStorage.clear());

  it("shows all five headline numbers on the closed summary line", () => {
    const { container } = renderSummary();
    const line = summaryLine(container);

    expect(container.querySelector("details")).not.toHaveAttribute("open");
    expect(line.querySelectorAll("[data-metric]")).toHaveLength(5);
    expect(line).toHaveTextContent("Requests1,234");
    expect(line).toHaveTextContent("Errors1.8%");
    expect(line).toHaveTextContent("P951,800 ms");
    expect(line).toHaveTextContent("Cache12.9%");
    expect(within(line).getByText("1.8%")).toHaveClass("text-destructive");
    expect(screen.queryByText(messages.cards.p99, { exact: false })).not.toBeInTheDocument();
    expect(screen.queryByTestId("charts")).not.toBeInTheDocument();
  });

  it("renders the metric cards and children once opened, and remembers it", async () => {
    const { container } = renderSummary();
    const details = container.querySelector("details") as HTMLDetailsElement;

    details.open = true;
    fireEvent(details, new Event("toggle"));

    await waitFor(() => expect(screen.getByTestId("charts")).toBeInTheDocument());
    expect(screen.getByText(messages.cards.p95)).toBeInTheDocument();
    expect(window.localStorage.getItem(RANGE_SUMMARY_STORAGE_KEY)).toBe("true");
  });

  it("opens from the stored preference after mount", async () => {
    window.localStorage.setItem(RANGE_SUMMARY_STORAGE_KEY, "true");
    const { container } = renderSummary();

    await waitFor(() => expect(container.querySelector("details")).toHaveAttribute("open"));
    expect(screen.getByTestId("charts")).toBeInTheDocument();
  });

  it("keeps the summary line informative when the summary is unavailable", () => {
    const { container } = renderSummary({ summaryUnavailable: true });

    expect(summaryLine(container)).toHaveTextContent(messages.analyticsUnavailable);
  });
});
