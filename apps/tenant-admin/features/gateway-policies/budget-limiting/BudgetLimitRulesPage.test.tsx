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
  "actions.edit": "Edit",
  "actions.close": "Close",
  "table.rowActions": "Actions for {ruleId}",
  "form.fields.ruleId": "Technical identifier",
  "alerts.error": "Error",
  "alerts.success": "Success",
  empty: "No budget limiting rules have been created yet.",
  "upgradeDialog.title": "This feature is not available on your current plan",
  "upgradeDialog.description": "Creating budget limit rules is not included in your current commercial plan. Upgrade your plan to use this feature.",
  "upgradeDialog.cancel": "Close",
  "upgradeDialog.action": "Upgrade / buy plan",
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

  it("opens the same drawer with the selected rule for editing", async () => {
    getBudgetLimitEntitlementsMock.mockResolvedValue({ features: { "budget_limit_rules.enabled": true } });
    listBudgetLimitRulesMock.mockResolvedValue({
      data: [{
        id: "rule-1",
        rule_id: "monthly-team-budget",
        enabled: true,
        priority: 100,
        conditions: { combinator: "AND", conditions: [] },
        limit_to: 1_000_000,
        unit: "cost_per_month",
        block_on_budget_exceed: true,
        alerts: { enabled: false },
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      }],
      total: 1,
      limit: 20,
      offset: 0,
    });

    render(<BudgetLimitRulesPage />);

    const actions = await screen.findByRole("button", { name: "Actions for monthly-team-budget" });
    fireEvent.pointerDown(actions, { button: 0, ctrlKey: false });
    fireEvent.click(await screen.findByRole("menuitem", { name: "Edit" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Technical identifier" })).toHaveValue("monthly-team-budget");
    expect(pushMock).not.toHaveBeenCalled();
  });
});
