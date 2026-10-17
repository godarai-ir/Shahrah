import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { createCreditPackOrder, type CreditPack, type CreditPackOrder } from "./api";
import { CreditPacksPanel } from "./CreditPacksPanel";

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

vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh: vi.fn() }) }));

vi.mock("./api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./api")>()),
  createCreditPackOrder: vi.fn(),
}));

const pack: CreditPack = {
  slug: "pack-1000",
  version: 1,
  credits: 1000,
  discount_percent: "10",
  validity_months: 12,
  price_irr: 9_000_000,
  tax_percent: "10",
  tax_irr: 900_000,
  payable_irr: 9_900_000,
};

function order(overrides: Partial<CreditPackOrder> = {}): CreditPackOrder {
  return {
    id: "order-1",
    pack_slug: "pack-1000",
    credits: 1000,
    validity_months: 12,
    payable_irr: 9_900_000,
    provider: "bank_transfer",
    status: "AWAITING_TRANSFER",
    created_at: "2026-10-17T10:00:00Z",
    ...overrides,
  };
}

describe("CreditPacksPanel", () => {
  beforeEach(() => vi.mocked(createCreditPackOrder).mockReset());

  it("orders a pack by bank transfer and opens the receipt form with the fixed amount", async () => {
    vi.mocked(createCreditPackOrder).mockResolvedValue(order());
    render(<CreditPacksPanel packs={[pack]} orders={[]} canManage paymentResult={null} loadError={null} />);

    expect(screen.getByText("10% off")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Bank transfer" }));

    await waitFor(() => expect(createCreditPackOrder).toHaveBeenCalledWith(expect.objectContaining({ pack_slug: "pack-1000", provider: "bank_transfer" })));
    expect(await screen.findByText("The pack order's total; transfer exactly this amount.")).toBeInTheDocument();
  });

  it("offers a receipt for orders awaiting a transfer and keeps buying read-only without access", () => {
    render(<CreditPacksPanel packs={[pack]} orders={[order(), order({ id: "order-2", status: "COMPLETED", provider: "zarinpal" })]} canManage={false} paymentResult="success" loadError={null} />);

    expect(screen.getByText("Payment succeeded")).toBeInTheDocument();
    expect(screen.getByText("Buying credits needs billing management access.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pay online" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "Submit receipt" })).not.toBeInTheDocument();
    expect(screen.getByText("Awaiting transfer")).toBeInTheDocument();
    expect(screen.getByText("Completed")).toBeInTheDocument();
  });
});
