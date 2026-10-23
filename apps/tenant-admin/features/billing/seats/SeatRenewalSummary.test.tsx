import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { SeatSummary } from "./api";
import { SeatRenewalSummary } from "./SeatRenewalSummary";

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    billing: (await import("@/messages/modules/billing/en.json")).default,
    billingSeats: (await import("@/messages/modules/billingSeats/en.json")).default,
  };
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

const summary: SeatSummary = {
  enforced: true,
  subscription_status: "active",
  subscription_kind: "seats",
  billing_interval: "monthly",
  period_start: "2026-10-01T00:00:00Z",
  period_end: "2026-11-01T00:00:00Z",
  renewal_opens_at: "2026-10-25T00:00:00Z",
  free_limit: 5,
  free_used: 0,
  unassigned: 0,
  pending_invitations: 0,
  paid: [
    { type: "pro", purchased: 3, assigned: 2, available: 1, renewing: 2 },
    { type: "max", purchased: 0, assigned: 0, available: 0, renewing: 0 },
  ],
  offers: [],
  pending_requests: 0,
  renewal_due: null,
  tax_percent: "10.00",
};

describe("SeatRenewalSummary", () => {
  afterEach(() => vi.useRealTimers());

  it("shows nothing when there are no paid seats", () => {
    const { container } = render(<SeatRenewalSummary summary={{ ...summary, paid: summary.paid.map((item) => ({ ...item, purchased: 0 })) }} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("points at the renewal page and says when it opens", () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-10-10T00:00:00Z"));
    render(<SeatRenewalSummary summary={summary} />);

    expect(screen.getByText(/Renewal opens on/)).toBeInTheDocument();
    expect(screen.getByText("3 seats · 2 given to members")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "View renewal" })).toHaveAttribute("href", "/billing/seats/renew");
  });

  it("makes renewing the main action while it is open and seats are unpaid", () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-10-28T00:00:00Z"));
    render(
      <SeatRenewalSummary
        summary={{
          ...summary,
          renewal_due: { lines: [], period_end: "2026-12-01T00:00:00Z", price_irr: 1, tax_percent: "10.00", tax_irr: 0, payable_irr: 1 },
        }}
      />,
    );

    expect(screen.getByText(/Renewal is open until/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Renew seats" })).toHaveAttribute("href", "/billing/seats/renew");
  });
});
