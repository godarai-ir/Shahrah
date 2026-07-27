import {
  clonePricing,
  isPricingResolved,
  type ModelPrice,
} from "@/app/components/pricing";

export type WizardCatalogModel = {
  model: string;
  types?: string[];
  public_pricing?: ModelPrice;
};

export type WizardAccountModel = {
  model: string;
  enabled: boolean;
  types?: string[];
  pricing_mode: "public" | "private";
  private_pricing?: ModelPrice;
};

export type WizardModelDraft = {
  enabled: boolean;
  selectedCapabilities: string[];
  pricingMode: "public" | "private";
  privatePricing?: ModelPrice;
};

export type WizardModelSavePayload = {
  model: string;
  effective_capabilities: string[];
  pricing_mode: "public" | "private";
  private_pricing?: ModelPrice;
};

export type WizardModelValidationErrorReason =
  | "private_pricing_required"
  | "effective_capabilities_required";

export type WizardModelValidationError = {
  model: string;
  reason: WizardModelValidationErrorReason;
};

export type WizardProviderAccountPricingPolicy = {
  auth_mode?: string;
  has_credential?: boolean;
  use_platform_managed_credentials?: boolean;
} | null | undefined;

export function canUseCustomModelPricing(account: WizardProviderAccountPricingPolicy): boolean {
  return Boolean(
    account &&
    account.has_credential &&
    !account.use_platform_managed_credentials &&
    account.auth_mode !== "no_credential",
  );
}

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
    selectedCapabilities: capabilities.length === 1 ? capabilities : [],
    pricingMode: "public",
    privatePricing: clonePricing(model.public_pricing),
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
      selectedCapabilities: normalizeCapabilities(item.types),
      pricingMode: item.pricing_mode ?? "public",
      privatePricing: clonePricing(item.private_pricing),
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
  options: { canUseCustomPricing?: boolean } = {},
): { models: WizardModelSavePayload[]; validationErrors: WizardModelValidationError[] } {
  const canUseCustomPricing = options.canUseCustomPricing ?? true;
  const models = Object.entries(drafts)
    .filter(([, draft]) => draft.enabled)
    .map(([model, draft]) => {
      const pricingMode = canUseCustomPricing ? draft.pricingMode : "public";
      return {
        model,
        effective_capabilities: normalizeCapabilities(draft.selectedCapabilities),
        pricing_mode: pricingMode,
        private_pricing:
          pricingMode === "private" ? buildPrivatePricing(draft) : undefined,
      };
    });

  const validationErrors: WizardModelValidationError[] = [];
  for (const item of models) {
    if (item.effective_capabilities.length === 0) {
      validationErrors.push({ model: item.model, reason: "effective_capabilities_required" });
    }
    if (item.pricing_mode !== "private") continue;
    if (!isPricingResolved(item.private_pricing)) {
      validationErrors.push({ model: item.model, reason: "private_pricing_required" });
    }
  }

  return { models, validationErrors };
}

function buildPrivatePricing(draft: WizardModelDraft): ModelPrice | undefined {
  return clonePricing(draft.privatePricing);
}
