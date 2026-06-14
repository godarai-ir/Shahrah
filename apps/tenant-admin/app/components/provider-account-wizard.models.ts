export type WizardCatalogModel = {
  model: string;
  types?: string[];
  public_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
};

export type WizardAccountModel = {
  model: string;
  enabled: boolean;
  types?: string[];
  pricing_mode: "public" | "private";
  private_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
};

export type WizardModelDraft = {
  enabled: boolean;
  compared: boolean;
  selectedCapabilities: string[];
  pricingMode: "public" | "private";
  inputPrice: string;
  outputPrice: string;
};

export type WizardModelSavePayload = {
  model: string;
  effective_capabilities: string[];
  pricing_mode: "public" | "private";
  private_pricing?:
    | {
        input_usd_per_1m?: number;
        output_usd_per_1m?: number;
      }
    | undefined;
};

export type WizardModelValidationErrorReason =
  | "private_pricing_required"
  | "effective_capabilities_required";

export type WizardModelValidationError = {
  model: string;
  reason: WizardModelValidationErrorReason;
};

function normalizeCapabilities(values: string[] | undefined): string[] {
  if (!Array.isArray(values)) return [];
  const seen = new Set<string>();
  const next: string[] = [];
  for (const item of values) {
    const normalized = item.trim().toLowerCase();
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    next.push(normalized);
  }
  return next;
}

function createDefaultDraft(model: WizardCatalogModel): WizardModelDraft {
  const capabilities = normalizeCapabilities(model.types);
  return {
    enabled: false,
    compared: false,
    selectedCapabilities: capabilities.length === 1 ? capabilities : [],
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
      selectedCapabilities: normalizeCapabilities(item.types),
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
    const autoSelectedCapabilities =
      !allSelected && draft.selectedCapabilities.length === 0 ? createDefaultDraft(model).selectedCapabilities : draft.selectedCapabilities;
    next[model.model] = {
      ...draft,
      enabled: !allSelected,
      selectedCapabilities: autoSelectedCapabilities,
    };
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
): { models: WizardModelSavePayload[]; validationErrors: WizardModelValidationError[] } {
  const models = Object.entries(drafts)
    .filter(([, draft]) => draft.enabled)
    .map(([model, draft]) => ({
      model,
      effective_capabilities: normalizeCapabilities(draft.selectedCapabilities),
      pricing_mode: draft.pricingMode,
      private_pricing:
        draft.pricingMode === "private"
          ? {
              input_usd_per_1m: toNumberOrUndefined(draft.inputPrice),
              output_usd_per_1m: toNumberOrUndefined(draft.outputPrice),
            }
          : undefined,
    }));

  const validationErrors: WizardModelValidationError[] = [];
  for (const item of models) {
    if (item.effective_capabilities.length === 0) {
      validationErrors.push({ model: item.model, reason: "effective_capabilities_required" });
    }
    if (item.pricing_mode !== "private") continue;
    const input = item.private_pricing?.input_usd_per_1m;
    const output = item.private_pricing?.output_usd_per_1m;
    if (typeof input !== "number" && typeof output !== "number") {
      validationErrors.push({ model: item.model, reason: "private_pricing_required" });
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
