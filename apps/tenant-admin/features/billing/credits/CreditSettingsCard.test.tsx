import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { updateCreditSettings } from "./api";
import { CreditSettingsCard } from "./CreditSettingsCard";

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

vi.mock("@/components/app-ui/feedback-toast", () => ({ showFeedbackToast: vi.fn() }));

vi.mock("./api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./api")>()),
  updateCreditSettings: vi.fn(),
}));

const settings = { billing_unit: "credit", overage_allowed: true, overage_enabled: false, overage_cap_credits: null };

describe("CreditSettingsCard", () => {
  beforeEach(() => {
    vi.mocked(updateCreditSettings).mockReset();
    vi.mocked(updateCreditSettings).mockImplementation(async (input = {}) => ({
      billing_unit: "credit",
      overage_allowed: true,
      overage_enabled: input.overage_enabled ?? false,
      overage_cap_credits: input.overage_cap_credits || null,
    }));
  });

  it("enables overage with a cap", async () => {
    render(<CreditSettingsCard initialSettings={settings} canManage />);
    fireEvent.click(screen.getByRole("switch", { name: "Allow overage" }));
    fireEvent.change(screen.getByLabelText("Overage cap (credits)"), { target: { value: "250.5" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(updateCreditSettings).toHaveBeenCalledWith({ overage_enabled: true, overage_cap_credits: "250.5" }));
  });

  it("rejects a cap with more than two decimals", () => {
    render(<CreditSettingsCard initialSettings={{ ...settings, overage_enabled: true }} canManage />);
    fireEvent.change(screen.getByLabelText("Overage cap (credits)"), { target: { value: "1.234" } });
    expect(screen.getByText("Enter a positive number with at most two decimals.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
  });

  it("locks the switch when the plan does not allow overage", () => {
    render(<CreditSettingsCard initialSettings={{ ...settings, overage_allowed: false }} canManage />);
    expect(screen.getByRole("switch", { name: "Allow overage" })).toBeDisabled();
    expect(screen.getByText("Your current plan does not support overage.")).toBeInTheDocument();
  });
});
