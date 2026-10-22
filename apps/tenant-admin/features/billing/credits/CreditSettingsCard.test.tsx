import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

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

const settings = { billing_unit: "credit", overage_allowed: true, overage_enabled: false, overage_cap_credits: null };

describe("CreditSettingsCard", () => {
  it("shows overage as read-only and points to support", () => {
    render(<CreditSettingsCard settings={settings} />);
    expect(screen.getByText("Off")).toBeInTheDocument();
    expect(screen.getByText(/Only GodarAI support can turn overage on/)).toBeInTheDocument();
    expect(screen.queryByRole("switch")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Save" })).not.toBeInTheDocument();
  });

  it("shows the cap when support enabled overage", () => {
    render(<CreditSettingsCard settings={{ ...settings, overage_enabled: true, overage_cap_credits: "250.50" }} />);
    expect(screen.getByText("Enabled")).toBeInTheDocument();
    expect(screen.getByText("Overage cap")).toBeInTheDocument();
    expect(screen.queryByText("No cap")).not.toBeInTheDocument();
  });

  it("says when there is no cap", () => {
    render(<CreditSettingsCard settings={{ ...settings, overage_enabled: true }} />);
    expect(screen.getByText("No cap")).toBeInTheDocument();
  });
});
