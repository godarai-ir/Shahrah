import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ListOrdered } from "lucide-react";

import { GatewayPolicyRulesList } from "../components";
import { BudgetLimitRuleForm } from "./BudgetLimitRuleForm";
import { createBudgetLimitRule, updateBudgetLimitRule } from "./ruleApi";
import type { BudgetLimitRule } from "./ruleTypes";

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    gatewayPolicies: (await import("@/messages/modules/gatewayPolicies/en.json")).default,
    budgetLimitRules: (await import("@/messages/modules/budgetLimitRules/en.json")).default,
    common: (await import("@/messages/modules/common/en.json")).default,
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

vi.mock("../lookups", () => ({
  loadPolicyModelOptions: vi.fn(async () => ({ providerAllowedModels: [], virtualModels: [] })),
  searchPolicyPrincipals: vi.fn(async () => []),
}));

vi.mock("./ruleApi", () => ({
  createBudgetLimitRule: vi.fn(async () => ({})),
  updateBudgetLimitRule: vi.fn(async () => ({})),
}));

function budgetRule(overrides: Partial<BudgetLimitRule> = {}): BudgetLimitRule {
  return {
    id: "1",
    rule_id: "monthly-cap",
    enabled: true,
    priority: 100,
    conditions: { combinator: "AND", conditions: [] },
    limit: "250.50",
    unit: "cost_per_month",
    on_exceed: "block",
    alert_thresholds: [80, 100],
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

/** The form's actions portal into the sheet footer, so it is rendered inside the real sheet. */
function inSheet(content: React.ReactNode) {
  const noop = () => undefined;
  return (
    <GatewayPolicyRulesList
      title="Rules"
      createAction={null}
      alertLabels={{ error: "Error", success: "Success" }}
      toolbar={null}
      filtered={false}
      itemCount={0}
      loading={false}
      loadingLabel="Loading"
      emptyIcon={ListOrdered}
      emptyTitle="Empty"
      emptyDescription="Empty"
      ruleSheet={{ open: true, onOpenChange: noop, title: "Rule", description: "Rule", closeLabel: "Close", content }}
      deleteDialog={{ open: false, onOpenChange: noop, title: "", description: "", cancelLabel: "", confirmLabel: "", pendingLabel: "", pending: false, onConfirm: noop }}
      upgradeDialog={{ open: false, onOpenChange: noop, title: "", description: "", cancelLabel: "", actionLabel: "" }}
    >
      {null}
    </GatewayPolicyRulesList>
  );
}

function renderForm(props: Partial<React.ComponentProps<typeof BudgetLimitRuleForm>> = {}) {
  const onSaved = vi.fn();
  render(inSheet(<BudgetLimitRuleForm rules={[]} onCancel={vi.fn()} onSaved={onSaved} {...props} />));
  return { onSaved };
}

describe("BudgetLimitRuleForm", () => {
  beforeEach(() => {
    vi.mocked(createBudgetLimitRule).mockClear();
    vi.mocked(updateBudgetLimitRule).mockClear();
  });

  it("previews the limit in credits", () => {
    renderForm({ initialRule: budgetRule() });
    expect(screen.getByText("For all requesters on all models: at most 250.5 credits per month")).toBeInTheDocument();
  });

  it("saves the credit limit, alert-only behavior, and thresholds", async () => {
    renderForm({ initialRule: budgetRule() });
    fireEvent.change(screen.getByLabelText("Budget limit (credits)"), { target: { value: "120.25" } });
    fireEvent.click(screen.getByRole("radio", { name: /Record an alert only/ }));
    const thresholds = screen.getByLabelText("Alert thresholds (% of limit)");
    fireEvent.change(thresholds, { target: { value: "۵۰" } });
    fireEvent.keyDown(thresholds, { key: "Enter" });

    const save = screen.getByRole("button", { name: "Save" });
    await waitFor(() => expect(save).toBeEnabled(), { timeout: 5000 });
    fireEvent.click(save);

    await waitFor(() => expect(updateBudgetLimitRule).toHaveBeenCalledOnce());
    const [ruleID, body] = vi.mocked(updateBudgetLimitRule).mock.calls[0];
    expect(ruleID).toBe("monthly-cap");
    expect(body).toMatchObject({ limit: "120.25", on_exceed: "alert", alert_thresholds: [50, 80, 100] });
    expect(body).not.toHaveProperty("limit_to");
    expect(body).not.toHaveProperty("block_on_budget_exceed");
  });

  it("reads a rule saved before on_exceed from its blocking flag", () => {
    renderForm({ initialRule: budgetRule({ on_exceed: undefined, block_on_budget_exceed: false }) });
    expect(screen.getByRole("radio", { name: /Record an alert only/ })).toBeChecked();
  });

  it("rejects more than two decimals", async () => {
    renderForm({ initialRule: budgetRule() });
    fireEvent.change(screen.getByLabelText("Budget limit (credits)"), { target: { value: "1.255" } });
    await waitFor(() => expect(screen.getByText("Enter credits with at most two decimals.")).toBeInTheDocument());
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
  });
});
