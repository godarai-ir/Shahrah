"use client";

import { useTranslations } from "next-intl";
import Link from "next/link";

import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { SettingsCard } from "@/components/app-ui/SettingsCard";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Button } from "@/components/ui/button";

import { useBillingFormat } from "../use-billing-format";
import type { CreditSettings } from "./api";

type CreditSettingsCardProps = {
  settings: CreditSettings;
};

/**
 * Shows whether the workspace may run into debt after the plan quota and
 * add-on packs run out. Overage extends credit, so only support can change it.
 */
export function CreditSettingsCard({ settings }: CreditSettingsCardProps) {
  const t = useTranslations("billing");
  const format = useBillingFormat();

  return (
    <SettingsCard
      title={t("credits.settings.title")}
      description={t("credits.settings.description")}
      actions={
        <Button asChild variant="outline" size="sm">
          <Link href="/billing/credits">{t("credits.settings.buyPacks")}</Link>
        </Button>
      }
    >
      <div className="flex flex-col gap-4">
        <DefinitionList columns={2}>
          <DefinitionListItem
            label={t("credits.settings.overage")}
            value={
              <StatusBadge status={settings.overage_enabled ? "active" : "inactive"}>
                {settings.overage_enabled ? t("credits.settings.enabled") : t("credits.settings.disabled")}
              </StatusBadge>
            }
          />
          {settings.overage_enabled ? (
            <DefinitionListItem
              label={t("credits.settings.cap")}
              value={settings.overage_cap_credits ? format.credits(settings.overage_cap_credits) : t("credits.settings.noCap")}
            />
          ) : null}
        </DefinitionList>
        <p className="text-sm text-muted-foreground">
          {settings.overage_enabled ? t("credits.settings.overageDescription") : t("credits.settings.disabledDescription")}
        </p>
        <p className="text-sm text-muted-foreground">{t("credits.settings.managedBySupport")}</p>
      </div>
    </SettingsCard>
  );
}
