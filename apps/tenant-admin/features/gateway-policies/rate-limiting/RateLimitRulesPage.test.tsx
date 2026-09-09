import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { RateLimitRulesPage } from "./RateLimitRulesPage";

const pushMock = vi.fn();
const getRateLimitEntitlementsMock = vi.fn();
const listRateLimitRulesMock = vi.fn();

const translations: Record<string, string> = {
  title: "Rate Limiting Rules",
  description: "Rate limiting description",
  searchPlaceholder: "Search",
  resultCount: "{count} rules",
  "actions.new": "Add New Rule",
  "actions.close": "Close",
  "alerts.error": "Error",
  "alerts.success": "Success",
  "emptyState.title": "No rate limiting rules yet",
  "emptyState.description": "Create the first rule.",
  "form.addTitle": "Add New Rule",
  "form.headerDescription": "Create a rate limit rule.",
  "upgradeDialog.title": "This feature is not available on your current plan",
  "upgradeDialog.description": "Upgrade your plan to use this feature.",
  "upgradeDialog.cancel": "Close",
  "upgradeDialog.action": "Upgrade",
};

vi.mock("next-intl", () => ({
  useLocale: () => "fa",
  useTranslations: () => (key: string, values?: Record<string, string | number>) => {
    const template = translations[key] ?? key;
    if (!values) return template;
    return Object.entries(values).reduce((result, [name, value]) => result.replace(`{${name}}`, String(value)), template);
  },
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

vi.mock("./RateLimitRuleForm", () => ({
  RateLimitRuleForm: () => <div>Rate limit form</div>,
}));

vi.mock("./ruleApi", () => ({
  createRateLimitRule: vi.fn(),
  deleteRateLimitRule: vi.fn(),
  getRateLimitEntitlements: () => getRateLimitEntitlementsMock(),
  listRateLimitRules: (...args: unknown[]) => listRateLimitRulesMock(...args),
  updateRateLimitRuleStatus: vi.fn(),
}));

describe("RateLimitRulesPage", () => {
  beforeEach(() => {
    pushMock.mockReset();
    getRateLimitEntitlementsMock.mockReset();
    listRateLimitRulesMock.mockReset();
    listRateLimitRulesMock.mockResolvedValue({ data: [], total: 0, limit: 20, offset: 0 });
  });

  afterEach(() => cleanup());

  it("shows the upgrade dialog when the feature is disabled", async () => {
    getRateLimitEntitlementsMock.mockResolvedValue({ features: { "rate_limit_rules.enabled": false } });
    render(<RateLimitRulesPage />);

    const button = await screen.findByRole("button", { name: "Add New Rule" });
    await waitFor(() => expect(button).toBeEnabled());
    fireEvent.click(button);

    expect(await screen.findByText("This feature is not available on your current plan")).toBeInTheDocument();
    expect(pushMock).not.toHaveBeenCalled();
  });

  it("opens the create drawer when the feature is enabled", async () => {
    getRateLimitEntitlementsMock.mockResolvedValue({ features: { "rate_limit_rules.enabled": true } });
    render(<RateLimitRulesPage />);

    const button = await screen.findByRole("button", { name: "Add New Rule" });
    await waitFor(() => expect(button).toBeEnabled());
    fireEvent.click(button);

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Rate limit form")).toBeInTheDocument();
    expect(pushMock).not.toHaveBeenCalled();
  });
});
