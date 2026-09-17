import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { TenantProviderAccount } from "@/features/provider-accounts/types";

import { ModelCatalogExplorer } from "./model-catalog-explorer";
import type { CatalogModel, ModelAvailability } from "./types";

vi.mock("next-intl", () => ({
  useLocale: () => "en-US",
  useTranslations: (namespace: string) => (key: string, values?: Record<string, unknown>) => {
    if (values) return `${namespace}.${key}:${JSON.stringify(values)}`;
    return `${namespace}.${key}`;
  },
}));

function account(overrides: Partial<TenantProviderAccount>): TenantProviderAccount {
  return {
    id: "acc-1",
    provider: "openai",
    name: "openai:main",
    auth_mode: "api_key",
    has_credential: true,
    use_platform_managed_credentials: false,
    ...overrides,
  } as TenantProviderAccount;
}

function availability(overrides: Partial<ModelAvailability> = {}): ModelAvailability {
  return {
    account_id: "acc-1",
    account_name: "openai:main",
    fqn: "openai:main/gpt-4.1",
    credential_source: "byok",
    effective_capabilities: ["chat"],
    pricing_mode: "public",
    pricing_resolved: true,
    readiness: "ready",
    effective_pricing: { input_usd_per_1m: 2, output_usd_per_1m: 8 },
    capability_readiness: [{ endpoint: "/v1/chat/completions", state: "contract_ready", live_validated: true }],
    ...overrides,
  };
}

const models: CatalogModel[] = [
  { provider: "openai", model: "gpt-4.1", display_name: "GPT 4.1", types: ["chat"], context_window: 1_047_576, availability: [availability()] },
  { provider: "openai", model: "gpt-4o", display_name: "GPT 4o", types: ["chat"], availability: [availability({ pricing_resolved: false })] },
  {
    provider: "openai",
    model: "gpt-3.5-turbo",
    display_name: "GPT 3.5 Turbo",
    types: ["chat"],
    is_deprecated: true,
    replacement_model: "gpt-4.1",
    availability: [availability()],
  },
];

function renderExplorer(overrides: Partial<React.ComponentProps<typeof ModelCatalogExplorer>> = {}) {
  const props: React.ComponentProps<typeof ModelCatalogExplorer> = {
    accounts: [
      account({ id: "acc-1", provider: "openai", name: "openai:main" }),
      account({ id: "acc-2", provider: "anthropic", name: "anthropic:prod" }),
      account({ id: "acc-3", provider: "openai", name: "openai:backup" }),
    ],
    docsHref: "/docs",
    summaries: {},
    models,
    providerBySlug: new Map(),
    selectedAccountID: "acc-1",
    modelAccountID: "acc-1",
    selectedModel: null,
    query: "",
    readinessFilter: "",
    loadingAccounts: false,
    loadingModels: false,
    refreshingModels: false,
    modelLoadError: null,
    onQueryChange: vi.fn(),
    onReadinessFilterChange: vi.fn(),
    onSelectAccount: vi.fn(),
    onSelectModel: vi.fn(),
    onTestConnection: vi.fn(),
    onDeleteAccount: vi.fn(),
    ...overrides,
  };
  render(<ModelCatalogExplorer {...props} />);
  return props;
}

describe("ModelCatalogExplorer", () => {
  it("groups accounts under provider headings with a menu on every account row", () => {
    const props = renderExplorer();
    const openai = screen.getByRole("region", { name: /openai/i });
    expect(within(openai).getByText("backup")).toBeInTheDocument();
    expect(within(openai).getByText("main")).toBeInTheDocument();
    expect(screen.getByRole("region", { name: /anthropic/i })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /modelCatalog\.actions\.openMenu/ })).toHaveLength(3);

    fireEvent.click(screen.getByRole("button", { name: /^prod/ }));
    expect(props.onSelectAccount).toHaveBeenCalledWith(expect.objectContaining({ id: "acc-2", provider: "anthropic" }));
  });

  it("summarizes readiness per row and hides prices that are not resolved", () => {
    renderExplorer();
    const rows = screen.getAllByRole("row").slice(1);
    expect(rows).toHaveLength(3);

    expect(within(rows[0]).getByText("modelCatalog.readiness.ready")).toBeInTheDocument();
    expect(within(rows[0]).getByText("$2")).toBeInTheDocument();
    expect(within(rows[0]).getByText("1M")).toBeInTheDocument();

    expect(within(rows[1]).getByText("modelCatalog.readiness.pricing")).toBeInTheDocument();
    expect(within(rows[1]).getAllByText("modelCatalog.models.unknownPrice")).toHaveLength(2);

    expect(within(rows[2]).getByText("modelCatalog.readiness.deprecated")).toBeInTheDocument();
    expect(within(rows[2]).getByText(/modelCatalog\.models\.replacedBy:\{"model":"gpt-4\.1"\}/)).toBeInTheDocument();
  });

  it("filters by readiness from the URL and reports chip changes", () => {
    const props = renderExplorer({ readinessFilter: "pricing" });
    expect(screen.getAllByRole("row")).toHaveLength(2);

    const group = screen.getByRole("group", { name: "modelCatalog.readiness.filterLabel" });
    const pricingChip = within(group).getByRole("button", { name: /modelCatalog\.readiness\.pricing/ });
    expect(pricingChip).toHaveAttribute("aria-pressed", "true");
    expect(within(group).getByRole("button", { name: /modelCatalog\.filters\.any/ })).toHaveTextContent("3");

    fireEvent.click(pricingChip);
    expect(props.onReadinessFilterChange).toHaveBeenCalledWith("");
  });

  it("searches locally and mirrors the query to the URL", () => {
    const props = renderExplorer();
    fireEvent.change(screen.getByRole("searchbox"), { target: { value: "4o" } });
    expect(props.onQueryChange).toHaveBeenCalledWith("4o");
    expect(screen.getAllByRole("row")).toHaveLength(2);
  });

  it("opens a model exactly once from its name button", () => {
    const props = renderExplorer();
    fireEvent.click(screen.getByRole("button", { name: /modelCatalog\.actions\.modelDetails.*GPT 4o/ }));
    expect(props.onSelectModel).toHaveBeenCalledWith(expect.objectContaining({ model: "gpt-4o" }));
    expect(props.onSelectModel).toHaveBeenCalledTimes(1);
  });
});
