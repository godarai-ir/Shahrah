import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { BudgetLimitRulesPage } from "./BudgetLimitRulesPage";

const pushMock = vi.fn();
const getBudgetLimitEntitlementsMock = vi.fn();
const listBudgetLimitRulesMock = vi.fn();

const translations: Record<string, string> = {
  title: "Budget Limiting Rules",
  description: "Budget limiting description",
  searchPlaceholder: "Search",
  "actions.new": "Add New Rule",
  "alerts.error": "Error",
  "alerts.success": "Success",
  empty: "No budget limiting rules have been created yet.",
  "upgradeDialog.title": "This feature is not available on your current plan",
  "upgradeDialog.description": "Creating budget limit rules is not included in your current commercial plan. Upgrade your plan to use this feature.",
  "upgradeDialog.cancel": "Close",
  "upgradeDialog.action": "Upgrade / buy plan",
};

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string, values?: Record<string, string | number>) => {
    const template = translations[key] ?? key;
    if (!values) return template;
    return Object.entries(values).reduce((result, [name, value]) => result.replace(`{${name}}`, String(value)), template);
  },
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

vi.mock("./ruleApi", () => ({
  createBudgetLimitRule: vi.fn(),
  deleteBudgetLimitRule: vi.fn(),
  getBudgetLimitEntitlements: () => getBudgetLimitEntitlementsMock(),
  listBudgetLimitRules: (...args: unknown[]) => listBudgetLimitRulesMock(...args),
  updateBudgetLimitRuleStatus: vi.fn(),
}));

describe("BudgetLimitRulesPage", () => {
  beforeEach(() => {
    pushMock.mockReset();
    getBudgetLimitEntitlementsMock.mockReset();
    listBudgetLimitRulesMock.mockReset();

    listBudgetLimitRulesMock.mockResolvedValue({
      data: [],
      total: 0,
      limit: 20,
      offset: 0,
    });
  });

  afterEach(() => {
    cleanup();
  });

  it("shows an upgrade dialog instead of navigating when the feature is disabled", async () => {
    getBudgetLimitEntitlementsMock.mockResolvedValue({
      features: {
        "budget_limit_rules.enabled": false,
      },
    });

    render(<BudgetLimitRulesPage />);

    const button = await screen.findByRole("button", { name: "Add New Rule" });
    await waitFor(() => expect(button).toBeEnabled());

    fireEvent.click(button);

    expect(await screen.findByText("This feature is not available on your current plan")).toBeInTheDocument();
    expect(pushMock).not.toHaveBeenCalled();
  });

  it("opens the create drawer when the feature is enabled", async () => {
    getBudgetLimitEntitlementsMock.mockResolvedValue({
      features: {
        "budget_limit_rules.enabled": true,
      },
    });

    render(<BudgetLimitRulesPage />);

    const button = await screen.findByRole("button", { name: "Add New Rule" });
    await waitFor(() => expect(button).toBeEnabled());

    fireEvent.click(button);

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "form.titleCreate", level: 1 })).toBeInTheDocument();
    expect(pushMock).not.toHaveBeenCalled();
  });
});
