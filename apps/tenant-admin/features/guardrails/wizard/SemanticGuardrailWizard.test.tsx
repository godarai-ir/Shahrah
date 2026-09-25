import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import en from "@/messages/modules/guardrails/en.json";

import { APIClientError } from "@/lib/http/api-client";

import { GuardrailWizard } from "./GuardrailWizard";

const api = vi.hoisted(() => ({
  listGuardrails: vi.fn(),
  getGuardrail: vi.fn(),
  createGuardrail: vi.fn(),
  updateGuardrail: vi.fn(),
  testSemanticGuardrail: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  usePathname: () => "/guardrails/new",
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), refresh: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  const messages = {
    guardrails: (await import("@/messages/modules/guardrails/en.json")).default,
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

vi.mock("@/components/app-ui/feedback-toast", () => ({ showFeedbackToast: vi.fn() }));

vi.mock("../api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../api")>()),
  ...api,
}));

function clickContinue() {
  fireEvent.click(screen.getByRole("button", { name: new RegExp(en.wizard.actions.next) }));
}

function openSemanticSettings() {
  render(<GuardrailWizard mode="create" />);
  fireEvent.click(screen.getByRole("button", { name: en.types.semantic_policy }));
  clickContinue();
}

beforeEach(() => {
  vi.clearAllMocks();
  api.listGuardrails.mockResolvedValue({ data: [], total: 0, limit: 50, offset: 0 });
  api.testSemanticGuardrail.mockResolvedValue({
    decision: "block",
    language: "english",
    rules: [
      { key: "jailbreak", probability: 0.97, decision: "block" },
      { key: "system_prompt_extraction", probability: 0.04, decision: "pass" },
    ],
    latency_ms: 312,
    truncated: false,
  });
});

describe("semantic policy wizard", () => {
  it("starts from the pack, defaults to audit, and tests on the gateway", async () => {
    openSemanticSettings();

    expect(screen.getByRole("radio", { name: en.wizard.config.semantic.sources.pack })).toHaveAttribute("aria-checked", "true");
    expect(screen.getByText(en.wizard.config.semantic.packRules.jailbreak.title)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: en.wizard.test.semantic.run }));

    await waitFor(() => expect(screen.getByRole("status")).toHaveTextContent(en.wizard.test.semantic.decision.block));
    expect(api.testSemanticGuardrail).toHaveBeenCalledWith(
      expect.objectContaining({ config: expect.objectContaining({ pack: "security-baseline@1" }), text: en.wizard.test.samples.semantic_policy }),
      expect.any(AbortSignal),
    );
    expect(screen.getByText("97%")).toBeInTheDocument();

    clickContinue();

    expect(screen.getByRole("radio", { name: new RegExp(en.enforcement.audit) })).toHaveAttribute("aria-checked", "true");
    expect(screen.getByRole("heading", { level: 3, name: en.wizard.behavior.semantic.thresholdsTitle })).toBeInTheDocument();
    expect(screen.getByText(en.wizard.behavior.semantic.failure.audit)).toBeInTheDocument();
  });

  it("picks an industry pack and its output rules", async () => {
    openSemanticSettings();
    const semantic = en.wizard.config.semantic;

    expect(screen.getByRole("radio", { name: semantic.targets.output })).toBeDisabled();
    expect(screen.getByText(semantic.packTargetLocked)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("radio", { name: semantic.packs.banking.title }));
    expect(screen.getByText(semantic.packRules.social_engineering.title)).toBeInTheDocument();
    expect(screen.getByText(semantic.packBetaNote)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("radio", { name: semantic.targets.output }));
    expect(screen.getByText(semantic.packRules.personal_financial_advice.title)).toBeInTheDocument();
    expect(screen.getByText(semantic.packRuleReviewOnly)).toBeInTheDocument();
    expect(screen.queryByText(semantic.packRules.jailbreak.title)).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: en.wizard.test.semantic.run }));

    await waitFor(() => expect(api.testSemanticGuardrail).toHaveBeenCalled());
    expect(api.testSemanticGuardrail).toHaveBeenCalledWith(
      expect.objectContaining({ config: expect.objectContaining({ pack: "banking@1", target: "output" }) }),
      expect.any(AbortSignal),
    );

    clickContinue();
    fireEvent.click(screen.getByRole("radio", { name: new RegExp(en.enforcement.enforce) }));
    expect(screen.getByText(en.wizard.behavior.semantic.betaPackAudit)).toBeInTheDocument();
  });

  it("requires a statement for custom rules and blocks inverted thresholds", () => {
    openSemanticSettings();

    fireEvent.click(screen.getByRole("radio", { name: en.wizard.config.semantic.sources.custom }));
    clickContinue();
    expect(screen.getByRole("alert")).toHaveTextContent(en.validation.semanticStatement);

    fireEvent.change(screen.getByLabelText(en.wizard.config.semantic.statement), { target: { value: "The reply promises a refund or a discount." } });
    expect(screen.getByText(en.wizard.config.semantic.warnings.compound)).toBeInTheDocument();
    clickContinue();

    fireEvent.change(screen.getAllByLabelText(en.wizard.behavior.semantic.review)[0], { target: { value: "0.9" } });
    clickContinue();

    expect(screen.getAllByText(en.validation.semanticThresholds).length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { level: 2, name: en.wizard.behavior.heading })).toBeInTheDocument();
  });

  it("explains when the gateway has no evaluator configured", async () => {
    api.testSemanticGuardrail.mockRejectedValue(new APIClientError("semantic guardrails are not configured", { status: 503, payload: null }));
    openSemanticSettings();

    fireEvent.click(screen.getByRole("button", { name: en.wizard.test.semantic.run }));

    await waitFor(() => expect(screen.getByRole("alert")).toHaveTextContent(en.wizard.test.semantic.errors.notConfigured));
  });
});
