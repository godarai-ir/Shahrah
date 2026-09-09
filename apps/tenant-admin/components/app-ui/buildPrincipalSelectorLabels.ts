import { useTranslations } from "next-intl";

import type { PrincipalEntitySelectLabels } from "@/components/app-ui/PrincipalEntitySelect";

export function buildPrincipalSelectorLabels(
  t: ReturnType<typeof useTranslations>,
  ux: ReturnType<typeof useTranslations>,
): PrincipalEntitySelectLabels {
  return {
    users: t("principal.users"),
    teams: t("principal.teams"),
    orgUnits: t("principal.orgUnits"),
    virtualAccounts: t("principal.virtualAccounts"),
    externalIdentities: t("principal.externalIdentities"),
    everyone: t("principal.everyone"),
    searchPlaceholder: t("principal.searchPlaceholder"),
    triggerPlaceholder: t("principal.searchPlaceholder"),
    loading: ux("actions.loading"),
    empty: t("principal.searchHint"),
    minSearchHint: ux("access.minSearchHint"),
    everyoneWarning: ux("access.everyoneWarning"),
  };
}


