import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { SeatSummary } from "./api";
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

function renderWorkspace(current: SeatSummary = summary) {
  return render(
    <SeatsWorkspace
      summary={current}
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
  it("summarizes the seat stock without a members list", () => {
    renderWorkspace();

    expect(screen.getByText("0 free seats left")).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "Members" })).not.toBeInTheDocument();
  });

  it("sends the buy action to the checkout page", () => {
    renderWorkspace({ ...summary, subscription_kind: "free", billing_interval: undefined, period_start: null, period_end: null });

    expect(screen.getByText("No seat subscription yet")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Buy seats" })).toHaveAttribute("href", "/billing/seats/buy");
  });
});
