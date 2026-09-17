import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const getDashboardAnalytics = vi.fn();

vi.mock("next-intl", () => ({
  useLocale: () => "en",
  useTranslations: (namespace: string) => (key: string) => `${namespace}.${key}`,
}));
vi.mock("next/dynamic", () => ({ default: () => () => <div data-testid="chart" /> }));
// getDashboardTimeseries closes over the real getDashboardAnalytics, so route it through the mock too.
vi.mock("@/features/dashboard/api", async () => {
  const actual = await vi.importActual<typeof import("./api")>("./api");
  return {
    ...actual,
    getDashboardAnalytics: (...args: unknown[]) => getDashboardAnalytics(...args),
    getDashboardTimeseries: (window: { from: string; to: string; intervalSeconds: number }, signal?: AbortSignal) =>
      getDashboardAnalytics(
        new URLSearchParams({ type: "timeseries", interval_seconds: String(window.intervalSeconds) }),
        signal,
      ),
  };
});

import { DashboardAnalytics } from "./DashboardAnalytics";

const intervalsRequested = () =>
  getDashboardAnalytics.mock.calls
    .map(([query]) => query as URLSearchParams)
    .filter((query) => query.get("type") === "timeseries")
    .map((query) => query.get("interval_seconds"));

describe("DashboardAnalytics", () => {
  beforeEach(() => {
    getDashboardAnalytics.mockReset();
  });

  it("refetches with the selected range's interval", async () => {
    getDashboardAnalytics.mockResolvedValue([]);
    const onRangeChange = vi.fn();
    const { rerender } = render(<DashboardAnalytics range="24h" onRangeChange={onRangeChange} refreshKey={0} />);
    await waitFor(() => expect(intervalsRequested()).toEqual(["3600"]));

    rerender(<DashboardAnalytics range="7d" onRangeChange={onRangeChange} refreshKey={0} />);
    await waitFor(() => expect(intervalsRequested()).toEqual(["3600", "21600"]));

    rerender(<DashboardAnalytics range="30d" onRangeChange={onRangeChange} refreshKey={0} />);
    await waitFor(() => expect(intervalsRequested()).toEqual(["3600", "21600", "86400"]));
  });

  it("shows the unavailable state when analytics fail", async () => {
    getDashboardAnalytics.mockRejectedValue(new Error("boom"));
    render(<DashboardAnalytics range="24h" onRangeChange={vi.fn()} refreshKey={0} />);

    expect(await screen.findByText("dashboard.analytics.unavailableTitle")).toBeInTheDocument();
    expect(screen.getByText("dashboard.analytics.unavailableDescription")).toBeInTheDocument();
    expect(screen.queryByText("dashboard.charts.trafficTitle")).not.toBeInTheDocument();
  });
});
