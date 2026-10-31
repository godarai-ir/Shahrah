import type {
  AccountDetailsResponse,
  ProviderCatalogModel,
  Readiness,
  TenantProviderAccount,
  TenantProviderAccountModel,
} from "./types";
import { hasCatalogPricing } from "./utils";

export function deriveAccountListReadiness(account: TenantProviderAccount): Readiness {
  if (!account.has_credential && !account.use_platform_managed_credentials && account.auth_mode !== "no_credential") {
    return "needs_credentials";
  }
  return "ready";
}

export function deriveAccountDetailsReadiness(
  account: TenantProviderAccount,
  models: TenantProviderAccountModel[],
  access: AccountDetailsResponse["access"] | undefined,
  catalogModels: ProviderCatalogModel[],
): Readiness {
  if (!account.has_credential && !account.use_platform_managed_credentials && account.auth_mode !== "no_credential") {
    return "needs_credentials";
  }
  if ((access?.managers ?? []).length === 0) {
    return "needs_access";
  }

  const catalogByModel = new Map(catalogModels.map((model) => [model.model, model]));
  const hasUnresolvedPricing = models.some((model) => {
    if (!model.enabled) return false;
    if (model.pricing_mode === "private") {
      return (
        typeof model.private_pricing?.input_usd_per_1m !== "number" &&
        typeof model.private_pricing?.output_usd_per_1m !== "number"
      );
    }
    return !hasCatalogPricing(catalogByModel.get(model.model));
  });
  return hasUnresolvedPricing ? "pricing_incomplete" : "ready";
}

export function mergeAccountIntoList(
  accounts: TenantProviderAccount[],
  nextAccount: TenantProviderAccount,
): TenantProviderAccount[] {
  let found = false;
  const next = accounts.map((item) => {
    if (item.id !== nextAccount.id) return item;
    found = true;
    return nextAccount;
  });
  return found ? next : [nextAccount, ...next];
}


