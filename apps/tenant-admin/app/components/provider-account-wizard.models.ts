export type WizardCatalogModel = {
  model: string;
  public_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
};

export type WizardAccountModel = {
  model: string;
  enabled: boolean;
  pricing_mode: "public" | "private";
  private_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
};

export type WizardModelDraft = {
  enabled: boolean;
  compared: boolean;
  pricingMode: "public" | "private";
  inputPrice: string;
  outputPrice: string;
};

export type WizardModelSavePayload = {
  model: string;
  pricing_mode: "public" | "private";
  private_pricing?:
    | {
        input_usd_per_1m?: number;
        output_usd_per_1m?: number;
      }
    | undefined;
};

function createDefaultDraft(model: WizardCatalogModel): WizardModelDraft {
  return {
    enabled: false,
    compared: false,
    pricingMode: "public",
    inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
    outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
  };
}

export function buildDefaultDrafts(
  models: WizardCatalogModel[],
  current: Record<string, WizardModelDraft>,
): Record<string, WizardModelDraft> {
  const next: Record<string, WizardModelDraft> = {};
  for (const model of models) {
    next[model.model] = current[model.model] ?? createDefaultDraft(model);
  }
  return next;
}

export function buildDraftsForAccount(
  models: WizardAccountModel[],
  catalogModels: WizardCatalogModel[],
): Record<string, WizardModelDraft> {
  const next = buildDefaultDrafts(catalogModels, {});
  for (const item of models) {
    next[item.model] = {
      enabled: item.enabled,
      compared: false,
      pricingMode: item.pricing_mode ?? "public",
      inputPrice: item.private_pricing?.input_usd_per_1m?.toString() ?? "",
      outputPrice: item.private_pricing?.output_usd_per_1m?.toString() ?? "",
    };
  }
  return next;
}

export function toggleVisibleModelSelection(
  visibleModels: WizardCatalogModel[],
  drafts: Record<string, WizardModelDraft>,
): Record<string, WizardModelDraft> {
  const allSelected =
    visibleModels.length > 0 && visibleModels.every((model) => Boolean(drafts[model.model]?.enabled));
  const next = { ...drafts };
  for (const model of visibleModels) {
    const draft = next[model.model] ?? createDefaultDraft(model);
    next[model.model] = { ...draft, enabled: !allSelected };
  }
  return next;
}

export function clearVisibleModelSelection(
  visibleModels: WizardCatalogModel[],
  drafts: Record<string, WizardModelDraft>,
): Record<string, WizardModelDraft> {
  const next = { ...drafts };
  for (const model of visibleModels) {
    const draft = next[model.model];
    if (!draft) continue;
    next[model.model] = { ...draft, enabled: false };
  }
  return next;
}

export function buildSelectedModelPayload(
  drafts: Record<string, WizardModelDraft>,
): { models: WizardModelSavePayload[]; validationErrors: string[] } {
  const models = Object.entries(drafts)
    .filter(([, draft]) => draft.enabled)
    .map(([model, draft]) => ({
      model,
      pricing_mode: draft.pricingMode,
      private_pricing:
        draft.pricingMode === "private"
          ? {
              input_usd_per_1m: toNumberOrUndefined(draft.inputPrice),
              output_usd_per_1m: toNumberOrUndefined(draft.outputPrice),
            }
          : undefined,
    }));

  const validationErrors: string[] = [];
  for (const item of models) {
    if (item.pricing_mode !== "private") continue;
    const input = item.private_pricing?.input_usd_per_1m;
    const output = item.private_pricing?.output_usd_per_1m;
    if (typeof input !== "number" && typeof output !== "number") {
      validationErrors.push(item.model);
    }
  }

  return { models, validationErrors };
}

function toNumberOrUndefined(value: string): number | undefined {
  const raw = value.trim();
  if (!raw) return undefined;
  const parsed = Number(raw);
  if (!Number.isFinite(parsed)) return undefined;
  return parsed;
}
