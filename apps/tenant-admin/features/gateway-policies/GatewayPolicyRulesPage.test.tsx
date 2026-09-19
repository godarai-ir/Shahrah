import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { BudgetLimitRule } from "./budget-limiting/ruleTypes";
import { GatewayPolicyRulesPage } from "./GatewayPolicyRulesPage";
import type { RateLimitRule } from "./rate-limiting/ruleTypes";

const navigation = vi.hoisted(() => ({ search: "", replace: vi.fn() }));

vi.mock("next/navigation", () => ({
  usePathname: () => "/gateway-policies",
  useRouter: () => ({ replace: navigation.replace, push: vi.fn(), refresh: vi.fn() }),
  useSearchParams: () => new URLSearchParams(navigation.search),
}));

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    gatewayPolicies: (await import("@/messages/modules/gatewayPolicies/en.json")).default,
    rateLimitRules: (await import("@/messages/modules/rateLimitRules/en.json")).default,
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

vi.mock("./lookups", () => ({
  loadPolicyModelOptions: vi.fn(async () => ({ providerAllowedModels: [], virtualModels: [] })),
  searchPolicyPrincipals: vi.fn(async () => []),
}));

function rateRule(ruleID: string, overrides: Partial<RateLimitRule> = {}): RateLimitRule {
  return {
    id: ruleID,
    rule_id: ruleID,
    name: ruleID,
    enabled: true,
    priority: 100,
    conditions: { combinator: "AND", conditions: [] },
    limit: { dimension: "requests", max: 600, window: { value: 1, unit: "minute" } },
    conflict_behavior: "most_restrictive",
    response_behavior: "return_429",
    soft_limit: false,
    retry_after_header: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

const budgetRule: BudgetLimitRule = {
  id: "b1",
  rule_id: "monthly-cap",
  name: "Monthly cap",
  enabled: true,
  priority: 100,
  conditions: { combinator: "AND", conditions: [] },
  limit_to: 5_000_000,
  unit: "cost_per_month",
  block_on_budget_exceed: true,
  alerts: { enabled: false },
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const rateRules = [
  rateRule("tenant-wide", { name: "Tenant wide", priority: 50 }),
  rateRule("checkout", {
    name: "Checkout team",
    priority: 10,
    conditions: {
      combinator: "AND",
      conditions: [
        { field: "subject", operator: "in", subjectKind: "team", values: ["t-checkout"], valueLabels: { "t-checkout": "Checkout" } },
        { field: "model", operator: "in", values: ["gpt-4o-mini"] },
      ],
    },
  }),
  rateRule("legacy", {
    name: "Legacy rule",
    priority: 90,
    conditions: { combinator: "AND", conditions: [{ field: "subject", operator: "in", values: ["user:u-1"] }] },
  }),
];

vi.mock("./rate-limiting/ruleApi", () => ({
  listRateLimitRules: vi.fn(async () => ({ data: rateRules, total: rateRules.length, limit: 1000, offset: 0 })),
  getRateLimitEntitlements: vi.fn(async () => ({ features: { "rate_limit_rules.enabled": true, "budget_limit_rules.enabled": true } })),
  createRateLimitRule: vi.fn(),
  deleteRateLimitRule: vi.fn(),
  updateRateLimitRuleStatus: vi.fn(),
  updateRateLimitRule: vi.fn(),
}));

vi.mock("./budget-limiting/ruleApi", () => ({
  listBudgetLimitRules: vi.fn(async () => ({ data: [budgetRule], total: 1, limit: 1000, offset: 0 })),
  createBudgetLimitRule: vi.fn(),
  deleteBudgetLimitRule: vi.fn(),
  updateBudgetLimitRuleStatus: vi.fn(),
  updateBudgetLimitRule: vi.fn(),
}));

describe("GatewayPolicyRulesPage", () => {
  beforeEach(() => {
    navigation.search = "";
    navigation.replace.mockClear();
  });

  it("lists rate rules in evaluation order as sentences, with budgets in their own section", async () => {
    render(<GatewayPolicyRulesPage />);

    const rateSection = await screen.findByRole("region", { name: /Rate limits · 3 rules/ });
    const names = within(rateSection).getAllByRole("listitem").map((row) => row.querySelector("button")?.textContent);
    expect(names).toEqual(["Checkout team", "Tenant wide", "Legacy rule"]);
    expect(within(rateSection).getByText("For team “Checkout” on gpt-4o-mini: at most 600 requests per minute")).toBeInTheDocument();

    const budgetSection = screen.getByRole("region", { name: /Budget caps · 1 rule/ });
    expect(within(budgetSection).getByText(/at most 5,000,000 IRR spent per month/)).toBeInTheDocument();
  });

  it("flags overlaps and legacy subjects on the row", async () => {
    render(<GatewayPolicyRulesPage />);
    const rateSection = await screen.findByRole("region", { name: /Rate limits/ });
    expect(within(rateSection).getByText("Overlaps with 1 other rule: Tenant wide")).toBeInTheDocument();
    expect(within(rateSection).getByText(/old format that the gateway never matches/)).toBeInTheDocument();
  });

  it("keeps the type filter in the URL and narrows the list", async () => {
    const { rerender } = render(<GatewayPolicyRulesPage />);
    fireEvent.click(await screen.findByRole("radio", { name: "Budget caps" }));
    expect(navigation.replace).toHaveBeenCalledWith("/gateway-policies?type=budget", { scroll: false });

    navigation.search = "type=budget";
    rerender(<GatewayPolicyRulesPage />);
    await waitFor(() => expect(screen.queryByRole("region", { name: /Rate limits/ })).toBeNull());
    expect(screen.getByRole("region", { name: /Budget caps/ })).toBeInTheDocument();
  });

  it("simulates an empty request against the loaded rules", async () => {
    render(<GatewayPolicyRulesPage />);
    const result = await screen.findByRole("region", { name: "Simulation result" });
    expect(within(result).getByRole("button", { name: "Tenant wide" })).toBeInTheDocument();
    expect(within(result).getByRole("button", { name: "Monthly cap" })).toBeInTheDocument();
    expect(within(result).queryByRole("button", { name: "Checkout team" })).toBeNull();
    expect(within(result).getAllByText("Enforced")).toHaveLength(2);
  });
});
