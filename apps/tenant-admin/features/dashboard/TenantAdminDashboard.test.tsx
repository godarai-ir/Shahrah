import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { AnalyticsPoint } from "./api";
import { HOUR_MS } from "./dashboard-metrics";

const getDashboardTimeseries = vi.fn();

vi.mock("next-intl", () => ({
  useLocale: () => "en",
  useTranslations: (namespace: string) => (key: string) => `${namespace}.${key}`,
}));
vi.mock("next/dynamic", () => ({ default: () => () => null }));
vi.mock("@/components/app-ui/RefreshButton", () => ({ RefreshButton: () => null }));
vi.mock("./api", () => ({
  getDashboardSubscription: vi.fn(),
  getDashboardEntitlements: vi.fn(),
  getDashboardTimeseries: (...args: unknown[]) => getDashboardTimeseries(...args),
}));

import TenantAdminDashboard from "./TenantAdminDashboard";

function hourly(lastHourErrorRate: number): AnalyticsPoint[] {
  const lastCompleteStart = Math.floor(Date.now() / HOUR_MS) * HOUR_MS - HOUR_MS;
  return Array.from({ length: 48 }, (_, index) => ({
    timestamp: new Date(lastCompleteStart - (47 - index) * HOUR_MS).toISOString(),
    request_count: 1_000,
    error_count: index === 47 ? Math.round(lastHourErrorRate * 1_000) : 10,
  }));
}

function mockSignals(points: AnalyticsPoint[]) {
  getDashboardTimeseries.mockImplementation((window: { intervalSeconds: number }) =>
    Promise.resolve(window.intervalSeconds === 3_600 ? points : []),
  );
}

const renderDashboard = () =>
  render(
    <TenantAdminDashboard
      initialSubscription={{ status: "active", billing_interval: "monthly", plan_display_name: "Team" }}
      initialEntitlements={null}
      initialError={null}
    />,
  );

describe("TenantAdminDashboard", () => {
  beforeEach(() => {
    getDashboardTimeseries.mockReset();
  });

  it("raises the error-rate alert when the last hour spikes past the threshold", async () => {
    mockSignals(hourly(0.2));
    renderDashboard();

    expect(await screen.findByText("dashboard.alerts.errorRateSpike.title")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /dashboard\.alerts\.errorRateSpike\.action/ })).toHaveAttribute(
      "href",
      "/logs?status=4xx,5xx",
    );
  });

  it("does not raise the alert for a steady error rate", async () => {
    mockSignals(hourly(0.01));
    renderDashboard();

    expect(await screen.findByText("dashboard.metrics.errorRate.label")).toBeInTheDocument();
    expect(screen.queryByText("dashboard.alerts.errorRateSpike.title")).not.toBeInTheDocument();
  });

  it("drops the plan card and hides P95 until the backend reports latency", async () => {
    mockSignals(hourly(0.01));
    renderDashboard();

    await waitFor(() => expect(screen.getByText("dashboard.metrics.errorRate.label")).toBeInTheDocument());
    expect(screen.queryByText("dashboard.overviewEyebrow")).not.toBeInTheDocument();
    expect(screen.queryByText("dashboard.metrics.latencyP95.label")).not.toBeInTheDocument();
    expect(screen.getByText("dashboard.metrics.quota.label")).toBeInTheDocument();
  });
});
