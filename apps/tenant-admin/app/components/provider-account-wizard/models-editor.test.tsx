import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ModelsEditor } from "./models-editor";
import type { ModelDraft, ProviderCatalogModel } from "./types";

const tMock = (key: string) => key;
const copyTextToClipboardMock = vi.hoisted(() => vi.fn().mockResolvedValue(undefined));

vi.mock("next-intl", () => ({
  useTranslations: () => tMock,
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

describe("ModelsEditor", () => {
  it("copies a valid pending-account FQN and opens its code guide in compact create mode", async () => {
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

    fireEvent.click(screen.getByRole("button", { name: "FQN" }));

    await waitFor(() => {
      expect(copyTextToClipboardMock).toHaveBeenCalledWith("ai21:ai21-brisk-orbit:jamba-large-1.7");
    });

    fireEvent.click(screen.getByRole("button", { name: "models.card.code" }));

    const dialog = screen.getByRole("dialog");
    expect(within(dialog).getByText("models.curlGuide.title")).toBeInTheDocument();
    expect(within(dialog).getByText("ai21:ai21-brisk-orbit:jamba-large-1.7")).toBeInTheDocument();
  });
});
