import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import en from "@/messages/modules/guardrails/en.json";

import { GUARDRAIL_CATALOG } from "./catalog";
import { GuardrailsPage } from "./GuardrailsPage";
import type { TenantGuardrail } from "./types";

const api = vi.hoisted(() => ({ listGuardrails: vi.fn(), deleteGuardrail: vi.fn() }));

vi.mock("next/navigation", () => ({
  usePathname: () => "/guardrails",
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

vi.mock("./api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./api")>()),
  ...api,
}));

function guardrailOf(type: TenantGuardrail["type"], index: number): TenantGuardrail {
  return {
    id: `g${index}`,
    slug: `guardrail-${index}`,
    selector: `guardrails/guardrail-${index}`,
    name: `Guardrail ${index}`,
    description: "",
    type,
    operation: "validate",
    enforcement: "enforce",
    enabled: index % 2 === 0,
    config: {},
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
}

const items = GUARDRAIL_CATALOG.map((entry, index) => guardrailOf(entry.type, index));

beforeEach(() => {
  vi.clearAllMocks();
  api.listGuardrails.mockResolvedValue({ data: items, total: items.length, limit: 50, offset: 0 });
});

describe("GuardrailsPage", () => {
  it("renders the catalog icon next to every guardrail type", async () => {
    const { container } = render(<GuardrailsPage docsHref="/docs/guardrails" />);
    await screen.findByText("Guardrail 0");

    for (const entry of GUARDRAIL_CATALOG) {
      expect(container.querySelector(`[data-guardrail-type-icon="${entry.type}"]`)).not.toBeNull();
    }
  });

  it("shows the behavior as one line and the status as text beside a dot", async () => {
    render(<GuardrailsPage docsHref="/docs/guardrails" />);
    const row = (await screen.findByText("Guardrail 0")).closest("tr")!;

    expect(within(row).getByText(`${en.operations.validate} · ${en.enforcement.enforce}`)).toBeInTheDocument();
    expect(within(row).getByText(en.status.enabled)).toBeInTheDocument();

    const disabledRow = screen.getByText("Guardrail 1").closest("tr")!;
    expect(within(disabledRow).getByText(en.status.disabled)).toBeInTheDocument();
  });

  it("offers to create a guardrail from the empty state", async () => {
    api.listGuardrails.mockResolvedValue({ data: [], total: 0, limit: 50, offset: 0 });
    render(<GuardrailsPage docsHref="/docs/guardrails" />);

    await screen.findByText(en.empty);
    const links = screen.getAllByRole("link", { name: new RegExp(en.actions.new) });
    expect(links.map((link) => link.getAttribute("href"))).toEqual(["/guardrails/new", "/guardrails/new"]);
  });
});
