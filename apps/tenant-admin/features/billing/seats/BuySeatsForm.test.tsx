import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { SeatOffer } from "./api";
import { BuySeatsForm, type BuySeatsTarget } from "./BuySeatsForm";

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

const offers: SeatOffer[] = [
  { type: "pro", monthly_credits: 1800, lifetime: false, pool_access: true, discount_percent: "0", monthly_price_irr: 18_000_000, yearly_price_irr: 180_000_000, active: true },
  { type: "max", monthly_credits: 9000, lifetime: false, pool_access: true, discount_percent: "0", monthly_price_irr: 90_000_000, yearly_price_irr: 900_000_000, active: true },
];

const quote = {
  seat: "pro",
  quantity: 1,
  billing_interval: "monthly",
  period_start: "2026-10-01T00:00:00Z",
  period_end: "2026-11-01T00:00:00Z",
  remaining_days: 12,
  period_days: 31,
  list_price_irr_per_credit: 10_000,
  unit_period_price_irr: 18_000_000,
  price_irr: 6_967_742,
  tax_percent: "10.00",
  tax_irr: 696_774,
  payable_irr: 7_664_516,
};

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

function renderForm(target: BuySeatsTarget = { seat: "pro" }) {
  return render(<BuySeatsForm target={target} offers={offers} />);
}

describe("BuySeatsForm", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("shows each seat's credits and buys one for the member the users list sent", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => jsonResponse(200, quote)));
    renderForm({ seat: "pro", assignUserID: "user-1", assignLabel: "Sara" });

    expect(screen.getByText("Buy a seat for Sara")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "Pro" })).toBeChecked();
    expect(screen.getByText("1,800 personal credits a month")).toBeInTheDocument();
    expect(await screen.findByText("12 of 31 days")).toBeInTheDocument();
    expect(screen.getByText("VAT (10%)")).toBeInTheDocument();
    expect(screen.queryByText("Billing interval")).not.toBeInTheDocument();
  });

  it("lets a free organization pick the interval of its first paid seat", async () => {
    const fetchMock = vi.fn(async (_url: string, init?: RequestInit) => {
      const body = JSON.parse(String(init?.body ?? "{}")) as { billing_interval?: string };
      const interval = body.billing_interval ?? "monthly";
      return jsonResponse(200, {
        ...quote,
        billing_interval: interval,
        remaining_days: 31,
        period_days: 31,
        price_irr: quote.unit_period_price_irr,
        starts_subscription: true,
      });
    });
    vi.stubGlobal("fetch", fetchMock);
    renderForm();

    expect(await screen.findByText("Billing interval")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: "Monthly" })).toBeChecked();
    expect(screen.getByText(/This payment starts your seat subscription/)).toBeInTheDocument();
    await waitFor(() =>
      expect(fetchMock).toHaveBeenLastCalledWith(
        "/api/billing/seats/quote",
        expect.objectContaining({ body: expect.stringContaining('"billing_interval":"monthly"') }),
      ),
    );
  });
});
