import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { CreditUsageSummary } from "./api";
import { CreditShadowCard, type CreditShadowCardLabels } from "./CreditShadowCard";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

const labels: CreditShadowCardLabels = {
  title: "Usage in credits",
  description: "Comparison only",
  badge: "Shadow mode",
  unit: "credits",
  total: "Credits used",
  cacheSaved: "Saved by cache",
  byModel: "Top models",
  byTeam: "Top teams",
  noTeam: "No team",
  emptyTitle: "No usage yet",
  emptyDescription: "Send a request",
  requestCount: (count) => `${count} requests`,
};

function summary(overrides: Partial<CreditUsageSummary> = {}): CreditUsageSummary {
  return {
    available: true,
    shadow_mode: true,
    preset: "current_month",
    period_start: "2026-09-22T20:30:00Z",
    period_end: "2026-10-12T00:00:00Z",
    total_credits: "1234.5",
    cache_saved_credits: "0.9",
    request_count: 3,
    by_model: [{ key: "anthropic/claude-sonnet-5", credits: "1200", request_count: 2 }],
    by_team: [
      { key: "team-1", label: "Platform", credits: "1000", request_count: 1 },
      { key: "", credits: "234.5", request_count: 2 },
    ],
    by_product: [],
    ...overrides,
  };
}

describe("CreditShadowCard", () => {
  it("shows totals and top models and teams with two-decimal credits", () => {
    render(<CreditShadowCard summary={summary()} labels={labels} locale="en" />);

    expect(screen.getByText("Shadow mode")).toBeInTheDocument();
    expect(screen.getByText("1,234.50 credits")).toBeInTheDocument();
    expect(screen.getByText("0.90 credits")).toBeInTheDocument();
    expect(screen.getByText("anthropic/claude-sonnet-5")).toBeInTheDocument();

    const teams = screen.getByRole("heading", { name: "Top teams" }).closest("section");
    expect(teams).not.toBeNull();
    expect(within(teams as HTMLElement).getByText("Platform")).toBeInTheDocument();
    expect(within(teams as HTMLElement).getByText("No team")).toBeInTheDocument();
    expect(within(teams as HTMLElement).getByText("234.50 credits")).toBeInTheDocument();
  });

  it("uses Persian digits for the fa locale", () => {
    render(<CreditShadowCard summary={summary()} labels={labels} locale="fa" />);

    expect(screen.getByText("۱٬۲۳۴٫۵۰ credits")).toBeInTheDocument();
  });

  it("renders the empty state when no requests were recorded", () => {
    render(<CreditShadowCard summary={summary({ request_count: 0, by_model: [], by_team: [] })} labels={labels} locale="en" />);

    expect(screen.getByText("No usage yet")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Top models" })).not.toBeInTheDocument();
  });

  it("links the public pricing changelog when given one", () => {
    render(<CreditShadowCard summary={summary()} labels={labels} locale="en" changelog={{ href: "https://godarai.ir/pricing/changelog", label: "Changelog" }} />);

    expect(screen.getByRole("link", { name: "Changelog" })).toHaveAttribute("href", "https://godarai.ir/pricing/changelog");
  });
});
