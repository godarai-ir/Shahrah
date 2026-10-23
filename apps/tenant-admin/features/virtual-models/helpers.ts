import type { useTranslations } from "next-intl";

import type { VirtualModelTarget } from "@/features/virtual-models/api";

import type { ProviderModelOption } from "./ProviderModelPicker";

export function buildTargetPreview(target: VirtualModelTarget, option?: ProviderModelOption) {
  const fqnParts = target.model.trim().split(":");
  const hasFQNParts = fqnParts.length === 3;
  const provider = option?.provider || target.provider || (hasFQNParts ? fqnParts[0].trim() : "");
  const accountName = option?.accountName || (hasFQNParts ? fqnParts[1].trim() : "");
  let modelName = option?.model || (hasFQNParts ? fqnParts[2].trim() : target.model.trim());
  const providerPrefix = `${provider}/`;

  if (modelName.toLocaleLowerCase().startsWith(providerPrefix.toLocaleLowerCase())) {
    modelName = modelName.slice(providerPrefix.length);
  }

  return { provider, accountName, modelName };
}

export function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

export function formatStrategyLabel(
  value: string,
  t: ReturnType<typeof useTranslations<"virtualModels">>,
) {
  return value === "weighted" ? t("form.weightedValue") : t("form.priorityValue");
}
