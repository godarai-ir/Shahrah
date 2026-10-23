import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { CreditUsageSummary } from "../api";
import { getCreditUsageSummary } from "./api";
import { CreditUsagePanel } from "./CreditUsagePanel";

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

vi.mock("./api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./api")>()),
  getCreditUsageSummary: vi.fn(),
}));

function summary(overrides: Partial<CreditUsageSummary> = {}): CreditUsageSummary {
  return {
    available: true,
    preset: "current_month",
    period_start: "2026-10-01T00:00:00Z",
    period_end: "2026-11-01T00:00:00Z",
    total_credits: "40.00",
    cache_saved_credits: "0.00",
    request_count: 4,
    by_model: [{ key: "openai/gpt-4o-mini", credits: "30.00", request_count: 3 }],
    by_team: [],
    by_user: [{ key: "dev@example.com", label: "dev@example.com", credits: "10.00", request_count: 1 }],
    by_product: [],
    ...overrides,
  };
}

describe("CreditUsagePanel", () => {
  it("lists the top models with their share and switches to users", async () => {
    render(<CreditUsagePanel initialSummary={summary()} />);
    const modelTable = screen.getByRole("table");
    expect(within(modelTable).getByText("openai/gpt-4o-mini")).toBeInTheDocument();
    expect(within(modelTable).getByText("75%")).toBeInTheDocument();

    fireEvent.mouseDown(screen.getByRole("tab", { name: "User" }));
    fireEvent.click(screen.getByRole("tab", { name: "User" }));
    await waitFor(() => expect(screen.getByText("dev@example.com")).toBeInTheDocument());
    expect(getCreditUsageSummary).not.toHaveBeenCalled();
  });

  it("shows the empty state when the period has no requests", () => {
    render(<CreditUsagePanel initialSummary={summary({ request_count: 0, by_model: [], by_user: [] })} />);
    expect(screen.getByText("No usage yet")).toBeInTheDocument();
  });
});
