import { useTranslations } from "next-intl";

import type { PrincipalEntitySelectLabels } from "@/app/components/principal-search-input";
export function buildPrincipalSelectorLabels(
  t: ReturnType<typeof useTranslations>,
  ux: ReturnType<typeof useTranslations>,
): PrincipalEntitySelectLabels {
  const isEnglish = t("principal.user").trim().toLowerCase() === "user";
  return {
    users: isEnglish ? "Users" : "کاربران",
    teams: isEnglish ? "Teams" : "تیم‌ها",
    orgUnits: isEnglish ? "Org units" : "واحدهای سازمانی",
    virtualAccounts: isEnglish ? "Virtual accounts" : "اکانت‌های مجازی",
    externalIdentities: isEnglish ? "External identities" : "هویت‌های خارجی",
    everyone: t("principal.everyone"),
    searchPlaceholder: t("principal.searchPlaceholder"),
    triggerPlaceholder: t("principal.searchPlaceholder"),
    loading: ux("actions.loading"),
    empty: t("principal.searchHint"),
    minSearchHint: ux("access.minSearchHint"),
    everyoneWarning: ux("access.everyoneWarning"),
  };
}


