"use client";

import { useLocale, useTranslations } from "next-intl";

import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import type { CredentialTestResult } from "@/features/provider-accounts/types";

export function WIFCredentialStatus({
  result,
}: {
  result: Partial<CredentialTestResult> | null;
}) {
  const locale = useLocale();
  const t = useTranslations("providerAccounts.ux.connectionTest");
  if (
    !result ||
    (!result.wif_expires_at &&
      !result.wif_last_refresh_at &&
      !result.wif_effective_scopes?.length)
  ) {
    return null;
  }

  return (
    <DefinitionList columns={2} aria-label={t("wifStatus")}>
      <DefinitionListItem
        surface
        label={t("wifScopes")}
        value={result.wif_effective_scopes?.join(" ") || "—"}
        valueDir="ltr"
        mono
        wrap
      />
      <DefinitionListItem
        surface
        label={t("wifExpiresAt")}
        value={
          <DateTimeValue
            value={result.wif_expires_at}
            locale={locale}
            fallback="—"
          />
        }
        valueDir="ltr"
      />
      <DefinitionListItem
        surface
        label={t("wifLastRefreshAt")}
        value={
          <DateTimeValue
            value={result.wif_last_refresh_at}
            locale={locale}
            fallback="—"
          />
        }
        valueDir="ltr"
      />
    </DefinitionList>
  );
}
