import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import en from "@/messages/modules/guardrails/en.json";

import { APIClientError } from "@/lib/http/api-client";

import type { TenantGuardrail } from "../types";
import { GuardrailWizard } from "./GuardrailWizard";

const navigation = vi.hoisted(() => ({ push: vi.fn(), replace: vi.fn() }));
const api = vi.hoisted(() => ({
  listGuardrails: vi.fn(),
  getGuardrail: vi.fn(),
  createGuardrail: vi.fn(),
  updateGuardrail: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  usePathname: () => "/guardrails/new",
  useRouter: () => ({ push: navigation.push, replace: navigation.replace, refresh: vi.fn() }),
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

const savedGuardrail: TenantGuardrail = {
  id: "g1",
  slug: "prompt-injection",
  selector: "guardrails/prompt-injection",
  name: "Prompt injection terms",
  description: "",
  type: "keyword_blocklist",
  operation: "validate",
  enforcement: "enforce",
  enabled: true,
  config: { terms: ["ignore previous instructions"] },
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function clickContinue() {
  fireEvent.click(screen.getByRole("button", { name: new RegExp(en.wizard.actions.next) }));
}

beforeEach(() => {
  vi.clearAllMocks();
  api.listGuardrails.mockResolvedValue({ data: [savedGuardrail], total: 1, limit: 50, offset: 0 });
  api.getGuardrail.mockResolvedValue(savedGuardrail);
  api.createGuardrail.mockImplementation(async (body) => ({ ...savedGuardrail, ...body }));
  api.updateGuardrail.mockResolvedValue(savedGuardrail);
});

describe("GuardrailWizard", () => {
  it("shows the step error and stays when continuing without a type", () => {
    render(<GuardrailWizard mode="create" />);

    clickContinue();

    expect(screen.getByRole("alert")).toHaveTextContent(en.validation.type);
    expect(screen.getByRole("heading", { level: 2, name: en.wizard.type.heading })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Guardrail type/ })).toHaveAttribute("aria-current", "step");
  });

  it("blocks the settings step until its config is valid", () => {
    render(<GuardrailWizard mode="create" />);

    fireEvent.click(screen.getByRole("button", { name: "Blocked terms" }));
    expect(screen.getByRole("button", { name: "Blocked terms" })).toHaveAttribute("aria-pressed", "true");
    clickContinue();
    expect(screen.getByRole("heading", { level: 2, name: "Blocked terms settings" })).toBeInTheDocument();

    clickContinue();

    expect(screen.getByRole("alert")).toHaveTextContent(en.validation.terms);
    expect(screen.getByRole("heading", { level: 2, name: "Blocked terms settings" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Add jailbreak" }));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    clickContinue();
    expect(screen.getByRole("heading", { level: 2, name: en.wizard.behavior.heading })).toBeInTheDocument();
  });

  it("locks the operation of validate-only types", () => {
    render(<GuardrailWizard mode="create" />);

    fireEvent.click(screen.getByRole("button", { name: "Not null" }));
    clickContinue();
    clickContinue();

    expect(screen.getByRole("radio", { name: "Validate" })).toHaveAttribute("aria-checked", "true");
    expect(screen.getByRole("radio", { name: "Rewrite content" })).toBeDisabled();
    expect(screen.getByText("Not null supports only “Validate”")).toBeInTheDocument();
  });

  it("creates a guardrail and shows the success page", async () => {
    render(<GuardrailWizard mode="create" />);
    await waitFor(() => expect(api.listGuardrails).toHaveBeenCalled());

    fireEvent.click(screen.getByRole("button", { name: "Not null" }));
    clickContinue();
    clickContinue();
    clickContinue();
    fireEvent.change(screen.getByLabelText(en.form.name), { target: { value: "Empty Guard" } });
    expect(screen.getByLabelText(en.form.identifier)).toHaveValue("empty-guard");
    fireEvent.click(screen.getByRole("button", { name: en.wizard.actions.create }));

    await screen.findByRole("heading", { name: "Guardrail “Empty Guard” was created" });
    expect(api.createGuardrail).toHaveBeenCalledWith(expect.objectContaining({ slug: "empty-guard", name: "Empty Guard", type: "not_null", operation: "validate" }));
    expect(screen.getByRole("link", { name: en.wizard.done.back })).toHaveAttribute("href", "/guardrails?created=empty-guard");
  });

  it("rejects a slug that is already in the list", async () => {
    render(<GuardrailWizard mode="create" />);
    await waitFor(() => expect(api.listGuardrails).toHaveBeenCalled());

    fireEvent.click(screen.getByRole("button", { name: "Not null" }));
    clickContinue();
    clickContinue();
    clickContinue();
    fireEvent.change(screen.getByLabelText(en.form.name), { target: { value: "Prompt injection" } });
    await waitFor(() => expect(screen.getByLabelText(en.form.identifier)).toHaveValue("prompt-injection"));
    expect(screen.getByLabelText(en.form.identifier)).toHaveAccessibleDescription(expect.stringContaining(en.messages.duplicate));
    fireEvent.click(screen.getByRole("button", { name: en.wizard.actions.create }));

    const alerts = screen.getAllByRole("alert").map((alert) => alert.textContent);
    expect(alerts).toHaveLength(2);
    expect(alerts.every((text) => text?.includes(en.messages.duplicate))).toBe(true);
    expect(api.createGuardrail).not.toHaveBeenCalled();
  });

  it("keeps the user on the review step when the server reports a conflict", async () => {
    api.createGuardrail.mockRejectedValue(new APIClientError("conflict", { status: 409, payload: null }));
    render(<GuardrailWizard mode="create" />);

    fireEvent.click(screen.getByRole("button", { name: "Not null" }));
    clickContinue();
    clickContinue();
    clickContinue();
    fireEvent.change(screen.getByLabelText(en.form.name), { target: { value: "Fresh" } });
    fireEvent.click(screen.getByRole("button", { name: en.wizard.actions.create }));

    expect(await screen.findByRole("alert")).toHaveTextContent(en.messages.duplicate);
    expect(screen.getByRole("heading", { level: 2, name: en.wizard.review.title })).toBeInTheDocument();
  });

  it("starts editing on the settings step with the type and slug locked", async () => {
    render(<GuardrailWizard mode="edit" slug="prompt-injection" />);

    expect(await screen.findByRole("heading", { level: 2, name: "Blocked terms settings" })).toBeInTheDocument();
    expect(api.getGuardrail).toHaveBeenCalledWith("prompt-injection");
    expect(screen.getByRole("button", { name: /Guardrail type/ })).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: /Name and review/ }));

    expect(screen.getByLabelText(en.form.identifier)).toBeDisabled();
    expect(screen.getByLabelText(en.form.identifier)).toHaveValue("prompt-injection");
    fireEvent.click(screen.getByRole("button", { name: en.wizard.actions.saveChanges }));

    await waitFor(() => expect(api.updateGuardrail).toHaveBeenCalledWith("prompt-injection", expect.not.objectContaining({ slug: expect.anything() })));
    expect(navigation.push).toHaveBeenCalledWith("/guardrails?updated=prompt-injection");
  });

  it("asks before discarding a dirty form", () => {
    render(<GuardrailWizard mode="create" />);

    fireEvent.click(screen.getByRole("button", { name: en.wizard.actions.cancel }));
    expect(navigation.push).toHaveBeenCalledWith("/guardrails");

    navigation.push.mockClear();
    fireEvent.click(screen.getByRole("button", { name: "Not null" }));
    fireEvent.click(screen.getByRole("button", { name: en.wizard.actions.cancel }));

    expect(screen.getByRole("alertdialog", { name: en.wizard.discard.title })).toBeInTheDocument();
    expect(navigation.push).not.toHaveBeenCalled();
  });
});
