import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ListOrdered } from "lucide-react";

import { GatewayPolicyRulesList } from "../components";
import type { PolicyRule } from "../policy-rules";
import { RateLimitRuleForm } from "./RateLimitRuleForm";
import { createRateLimitRule, updateRateLimitRule } from "./ruleApi";
import type { RateLimitRule } from "./ruleTypes";

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    gatewayPolicies: (await import("@/messages/modules/gatewayPolicies/en.json")).default,
    rateLimitRules: (await import("@/messages/modules/rateLimitRules/en.json")).default,
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
  createRateLimitRule: vi.fn(async () => ({})),
  updateRateLimitRule: vi.fn(async () => ({})),
}));

function rateRule(overrides: Partial<RateLimitRule> = {}): RateLimitRule {
  return {
    id: "1",
    rule_id: "checkout-rpm",
    name: "Checkout RPM",
    enabled: true,
    priority: 10,
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

function renderForm(props: Partial<React.ComponentProps<typeof RateLimitRuleForm>> = {}) {
  const onSaved = vi.fn();
  const view = render(inSheet(<RateLimitRuleForm rules={[]} onCancel={vi.fn()} onSaved={onSaved} {...props} />));
  return { onSaved, unmount: view.unmount };
}

describe("RateLimitRuleForm", () => {
  beforeEach(() => {
    vi.mocked(createRateLimitRule).mockClear();
    vi.mocked(updateRateLimitRule).mockClear();
  });

  it("states that empty target groups mean everything and previews the sentence", () => {
    renderForm();
    expect(screen.getAllByText("All requesters").length).toBeGreaterThan(0);
    expect(screen.getAllByText("All models").length).toBeGreaterThan(0);
    expect(screen.getByText("For all requesters on all models: at most 60 requests per minute")).toBeInTheDocument();
    expect(screen.getByText(/No metadata condition/)).toBeInTheDocument();
  });

  it("warns about overlapping live rules and says so when there are none", () => {
    const other: PolicyRule = { type: "rate", rule: rateRule({ rule_id: "tenant-wide", name: "Tenant wide" }) };
    const { unmount } = renderForm({ rules: [other] });
    expect(screen.getByText(/Overlaps with 1 other rule: Tenant wide/)).toBeInTheDocument();
    unmount();

    renderForm({ rules: [{ type: "rate", rule: rateRule({ enabled: false }) }] });
    expect(screen.getByText("Does not overlap any other active rule.")).toBeInTheDocument();
  });

  it("upgrades a legacy subject and saves it in the shape the gateway reads", async () => {
    const legacy = rateRule({
      conditions: { combinator: "AND", conditions: [{ field: "subject", operator: "in", values: ["team:t-1"] }] },
    });
    const { onSaved } = renderForm({ initialRule: legacy });
    expect(screen.getByText(/stored in an old format/)).toBeInTheDocument();

    const save = screen.getByRole("button", { name: "Save" });
    await waitFor(() => expect(save).toBeEnabled(), { timeout: 5000 });
    fireEvent.click(save);

    await waitFor(() => expect(updateRateLimitRule).toHaveBeenCalledOnce());
    const [ruleID, body] = vi.mocked(updateRateLimitRule).mock.calls[0];
    expect(ruleID).toBe("checkout-rpm");
    expect(body.conditions?.conditions).toEqual([{ field: "subject", operator: "in", subjectKind: "team", values: ["t-1"] }]);
    await waitFor(() => expect(onSaved).toHaveBeenCalled());
  });

  it("turns soft limit and stop-evaluation into explicit choices", async () => {
    renderForm({ initialRule: rateRule() });
    fireEvent.click(screen.getByRole("radio", { name: /Count only/ }));
    fireEvent.click(screen.getByRole("radio", { name: /Stop evaluation/ }));
    const save = screen.getByRole("button", { name: "Save" });
    await waitFor(() => expect(save).toBeEnabled(), { timeout: 5000 });
    fireEvent.click(save);

    await waitFor(() => expect(updateRateLimitRule).toHaveBeenCalledOnce());
    expect(vi.mocked(updateRateLimitRule).mock.calls[0][1]).toMatchObject({ soft_limit: true, conflict_behavior: "first_match" });
  });

  it("keeps save disabled until the rule has an ID", () => {
    renderForm();
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
  });
});
