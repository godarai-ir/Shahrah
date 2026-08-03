import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ModelsEditor } from "./models-editor";
import type { ModelDraft, ProviderCatalogModel, TenantProviderAccount } from "./types";

const tMock = (key: string) => key;
const copyTextToClipboardMock = vi.hoisted(() => vi.fn().mockResolvedValue(undefined));

vi.mock("next-intl", () => ({
  useTranslations: () => tMock,
  useLocale: () => "en",
}));

vi.mock("./utils", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./utils")>();
  return {
    ...actual,
    requestJSON: vi.fn().mockResolvedValue({
      data: {
        currencyUnit: "USD",
        priceScale: "PER_MILLION_TOKENS",
        exchangeRate: { usdToIrr: 0 },
      },
    }),
  };
});

vi.mock("@/components/app-ui/CopyableValue", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/components/app-ui/CopyableValue")>();
  return {
    ...actual,
    copyTextToClipboard: copyTextToClipboardMock,
  };
});

const model: ProviderCatalogModel = {
  provider: "ai21",
  model: "ai21/jamba-large-1.7",
  display_name: "Jamba Large 1.7",
  mode: ["chat"],
  types: ["chat"],
  context_window: 256_000,
  input_modalities: ["text"],
  output_modalities: ["text"],
  capabilities: ["function_calling"],
  public_pricing: {
    input_usd_per_1m: 2,
    output_usd_per_1m: 8,
  },
};

const draft: ModelDraft = {
  enabled: true,
  selectedCapabilities: ["chat"],
  pricingMode: "public",
  inputPrice: "2",
  outputPrice: "8",
};

const account: TenantProviderAccount = {
  id: "account-1",
  provider: "ai21",
  name: "ai21-brisk-orbit",
  auth_mode: "api_key",
  has_credential: true,
  use_platform_managed_credentials: false,
};

describe("ModelsEditor", () => {
  it("hides FQN and code actions while the provider account is not created", () => {
    render(
      <ModelsEditor
        t={tMock as never}
        account={null}
        pendingAccountName="ai21-brisk-orbit"
        pendingProvider="ai21"
        setBanner={vi.fn()}
        providerModels={[model]}
        modelDrafts={{ [model.model]: draft }}
        setModelDrafts={vi.fn()}
        isSaving={false}
        isLoadingModels={false}
        canManage
        onSave={vi.fn()}
        variant="compact"
      />,
    );

    expect(screen.queryByRole("button", { name: "FQN" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "models.card.code" })).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("keeps FQN and code actions available for an existing provider account", async () => {
    copyTextToClipboardMock.mockClear();
    render(
      <ModelsEditor
        t={tMock as never}
        account={account}
        setBanner={vi.fn()}
        providerModels={[model]}
        modelDrafts={{ [model.model]: draft }}
        setModelDrafts={vi.fn()}
        isSaving={false}
        isLoadingModels={false}
        canManage
        onSave={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "FQN" }));

    await waitFor(() => {
      expect(copyTextToClipboardMock).toHaveBeenCalledWith("ai21:ai21-brisk-orbit:jamba-large-1.7");
    });

    fireEvent.click(screen.getByRole("button", { name: "models.card.code" }));

    const dialog = screen.getByRole("dialog");
    expect(within(dialog).getByText("models.curlGuide.title")).toBeInTheDocument();
    expect(within(dialog).getByLabelText("models.curlGuide.modelFqnLabel")).toHaveValue("ai21:ai21-brisk-orbit:jamba-large-1.7");
    expect(within(dialog).getByRole("tablist")).toBeInTheDocument();
    expect(within(dialog).getByRole("tab", { name: "models.curlGuide.shells.powershell" })).toHaveAttribute("data-state", "active");
  });

  it("renders provider models in pages of 25 while preserving accessible model labels", () => {
    const providerModels = Array.from({ length: 107 }, (_, index) => ({
      ...model,
      model: `model-${index}`,
      display_name: `Model ${index}`,
    }));
    const modelDrafts = Object.fromEntries(
      providerModels.map((item) => [item.model, { ...draft, enabled: false }]),
    );

    render(
      <ModelsEditor
        t={tMock as never}
        account={null}
        pendingAccountName="ai21-brisk-orbit"
        pendingProvider="ai21"
        setBanner={vi.fn()}
        providerModels={providerModels}
        modelDrafts={modelDrafts}
        setModelDrafts={vi.fn()}
        isSaving={false}
        isLoadingModels={false}
        canManage
        onSave={vi.fn()}
        variant="compact"
      />,
    );

    expect(document.querySelectorAll("tbody tr")).toHaveLength(25);
    expect(screen.getByRole("checkbox", { name: "common.enabled Model 0" })).toBeInTheDocument();
    expect(screen.queryByRole("checkbox", { name: "common.enabled Model 25" })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "models.pagination.next" }));
    expect(document.querySelectorAll("tbody tr")).toHaveLength(25);
    expect(screen.getByRole("checkbox", { name: "common.enabled Model 25" })).toBeInTheDocument();
  });
});
