import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { SeatSummary } from "./api";
import type { BuySeatsTarget } from "./BuySeatsSheet";
import { SeatsWorkspace } from "./SeatsWorkspace";

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

vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh: vi.fn() }) }));

const summary: SeatSummary = {
  enforced: true,
  subscription_status: "active",
  subscription_kind: "seats",
  billing_interval: "monthly",
  period_start: "2026-10-01T00:00:00Z",
  period_end: "2026-11-01T00:00:00Z",
  free_limit: 5,
  free_used: 5,
  unassigned: 1,
  pending_invitations: 0,
  paid: [
    { type: "pro", purchased: 0, assigned: 0, available: 0, renewing: 0 },
    { type: "max", purchased: 0, assigned: 0, available: 0, renewing: 0 },
  ],
  offers: [],
  pending_requests: 0,
  renewal_due: null,
  tax_percent: "10.00",
};

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

function renderWorkspace(initialBuyTarget?: BuySeatsTarget, current: SeatSummary = summary) {
  return render(
    <SeatsWorkspace
      summary={current}
      requests={[]}
      orders={[]}
      settings={null}
      canManageMembers
      canManageBilling
      paymentResult={null}
      initialBuyTarget={initialBuyTarget}
    />,
  );
}

describe("SeatsWorkspace", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("summarizes the seat stock without a members list", () => {
    renderWorkspace();

    expect(screen.getByText("0 free seats left")).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "Members" })).not.toBeInTheDocument();
  });

  it("opens checkout for the member the users list sent here", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => jsonResponse(200, quote)));
    renderWorkspace({ seat: "pro", assignUserID: "user-1", assignLabel: "Sara" });

    expect(await screen.findByText("Buy a seat for Sara")).toBeInTheDocument();
    expect(await screen.findByText("12 of 31 days")).toBeInTheDocument();
    expect(screen.queryByText("Billing interval")).not.toBeInTheDocument();
  });

  it("lets a free organization buy its first paid seat and pick the interval", async () => {
    const freeSummary: SeatSummary = {
      ...summary,
      subscription_kind: "free",
      billing_interval: undefined,
      period_start: null,
      period_end: null,
    };
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
    renderWorkspace(undefined, freeSummary);

    expect(screen.getByText("No seat subscription yet")).toBeInTheDocument();

    const buy = screen.getByRole("button", { name: "Buy seats" });
    expect(buy).toBeEnabled();
    buy.click();

    expect(await screen.findByText("Billing interval")).toBeInTheDocument();
    expect(screen.getByText(/This payment starts your seat subscription/)).toBeInTheDocument();
    await waitFor(() =>
      expect(fetchMock).toHaveBeenLastCalledWith(
        "/api/billing/seats/quote",
        expect.objectContaining({ body: expect.stringContaining('"billing_interval":"monthly"') }),
      ),
    );
  });
});
