import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { SeatMember, SeatSummary } from "./api";
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
  billing_interval: "monthly",
  period_start: "2026-10-01T00:00:00Z",
  period_end: "2026-11-01T00:00:00Z",
  free_limit: 10,
  free_used: 10,
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

const member: SeatMember = {
  user_id: "user-1",
  email: "sara@example.com",
  name: "Sara",
  role: "member",
  active: true,
  seat: "free",
  pending_seat: "",
  pending_effective_at: null,
  free_seat_used: true,
  personal_granted_credits: "500.00",
  personal_remaining_credits: "125.00",
  pool_used_credits: "0.00",
  pool_limit_credits: null,
  assigned_at: "2026-10-02T00:00:00Z",
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

function renderWorkspace() {
  return render(
    <SeatsWorkspace
      summary={summary}
      members={[member]}
      requests={[]}
      orders={[]}
      settings={null}
      canManageMembers
      canManageBilling
      paymentResult={null}
    />,
  );
}

describe("SeatsWorkspace", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("shows each member's personal credits", () => {
    renderWorkspace();

    expect(screen.getByText("0 free seats left")).toBeInTheDocument();
    expect(screen.getByText("Sara")).toBeInTheDocument();
    expect(screen.getByText(/125\.00 .* of 500\.00/)).toBeInTheDocument();
  });

  it("offers to buy a seat for the member when an upgrade has none in stock", async () => {
    const fetchMock = vi.fn(async (url: string) =>
      url.includes("/seats/members/")
        ? jsonResponse(409, { error: { code: "seat_purchase_required", type: "seats", message: "no seat", quote } })
        : jsonResponse(200, quote),
    );
    vi.stubGlobal("fetch", fetchMock);
    renderWorkspace();

    fireEvent.pointerDown(screen.getByRole("button", { name: "Actions for Sara" }), { button: 0, ctrlKey: false });
    fireEvent.click(await screen.findByRole("menuitem", { name: "Change seat" }));
    fireEvent.click(await screen.findByRole("button", { name: "Apply" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("/api/billing/seats/members/user-1", expect.objectContaining({ method: "PUT", body: JSON.stringify({ seat: "pro" }) })),
    );
    expect(await screen.findByText("No seat of this type is available")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Buy and assign/ }));
    expect(await screen.findByText("Buy a seat for Sara")).toBeInTheDocument();
    expect(await screen.findByText("12 of 31 days")).toBeInTheDocument();
  });
});
