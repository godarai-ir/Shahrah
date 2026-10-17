import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { CreditUsageSummary } from "../api";
import type { CreditBalance } from "./api";
import { CreditOverview } from "./CreditOverview";

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = { billing: (await import("@/messages/modules/billing/en.json")).default };
  const translators = new Map<string, unknown>();
  return {
    ...actual,
    useLocale: () => "en",
    useTranslations: (namespace: keyof typeof messages) => {
      if (!translators.has(namespace)) translators.set(namespace, actual.createTranslator({ locale: "en", messages, namespace }));
      return translators.get(namespace);
    },
  };
});

vi.mock("@/hooks/use-locale-format", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/hooks/use-locale-format")>()),
  useAppLocale: () => "en",
}));

const inDays = (days: number) => new Date(Date.now() + days * 24 * 60 * 60 * 1000 + 60_000).toISOString();

function balance(overrides: Partial<CreditBalance> = {}): CreditBalance {
  return {
    available_credits: "498.00",
    debt_credits: "0.00",
    lots: [{ lot_id: "l1", source: "trial", amount_credits: "500.00", remaining_credits: "498.00", expires_at: inDays(14), created_at: "2026-10-17T00:00:00Z" }],
    forecast: { status: "ok", window_days: 7, consumed_credits: "14.00", daily_average_credits: "2.00", depletes_at: inDays(30) },
    settings: { billing_unit: "credit", overage_allowed: false, overage_enabled: false, overage_cap_credits: null },
    ...overrides,
  };
}

const usage: CreditUsageSummary = {
  available: true,
  preset: "current_month",
  period_start: "2026-10-01T00:00:00Z",
  period_end: "2026-11-01T00:00:00Z",
  total_credits: "40.00",
  cache_saved_credits: "10.00",
  request_count: 12,
  by_model: [],
  by_team: [],
  by_user: [],
  by_product: [{ key: "semantic_cache", credits: "1.50", request_count: 3 }],
};

describe("CreditOverview", () => {
  it("shows the balance, this month's usage, net cache savings, and the forecast", () => {
    render(<CreditOverview balance={balance()} usage={usage} />);
    expect(screen.getAllByText("498.00 credits").length).toBeGreaterThan(0);
    expect(screen.getByText("40.00 credits")).toBeInTheDocument();
    expect(screen.getByText("8.50 credits")).toBeInTheDocument();
    expect(screen.getByText("In about 30 days")).toBeInTheDocument();
    expect(screen.getByText("At 2.00 credits a day over the last 7 days")).toBeInTheDocument();
    expect(screen.getByText("Trial credits")).toBeInTheDocument();
  });

  it("states when there is no usage to forecast from and flags short history", () => {
    const { unmount } = render(<CreditOverview balance={balance({ forecast: { status: "no_usage", window_days: 7, consumed_credits: "0.00", daily_average_credits: "0.00" } })} usage={usage} />);
    expect(screen.getByText("No estimate")).toBeInTheDocument();
    expect(screen.getByText("No usage in the last 7 days.")).toBeInTheDocument();
    unmount();

    render(<CreditOverview balance={balance({ forecast: { status: "uncertain", window_days: 7, consumed_credits: "1.00", daily_average_credits: "2.00", depletes_at: inDays(3) } })} usage={usage} />);
    expect(screen.getByText("Little data")).toBeInTheDocument();
  });

  it("warns about debt and reports a missing balance", () => {
    const { unmount } = render(<CreditOverview balance={balance({ debt_credits: "12.00" })} usage={null} />);
    expect(screen.getByText("Credit debt")).toBeInTheDocument();
    expect(screen.getAllByText("Unavailable").length).toBe(2);
    unmount();

    render(<CreditOverview balance={null} usage={null} />);
    expect(screen.getByText("The credit balance could not be loaded. Reload the page.")).toBeInTheDocument();
  });
});
