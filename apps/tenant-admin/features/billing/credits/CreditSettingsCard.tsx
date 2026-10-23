"use client";

import { useTranslations } from "next-intl";
import { LifeBuoy, Package } from "lucide-react";
import Link from "next/link";

import { AppCard } from "@/components/app-ui/AppCard";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
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
  const enabled = settings.overage_enabled;

  return (
    <AppCard
      title={
        <span className="flex flex-wrap items-center gap-2">
          {t("credits.settings.title")}
          <StatusBadge status={enabled ? "active" : "inactive"}>
            {enabled ? t("credits.settings.enabled") : t("credits.settings.disabled")}
          </StatusBadge>
        </span>
      }
      description={enabled ? t("credits.settings.overageDescription") : t("credits.settings.disabledDescription")}
      actions={
        <Button asChild variant="outline">
          <Link href="/billing/credits">
            <Package data-icon="inline-start" />
            {t("credits.settings.buyPacks")}
          </Link>
        </Button>
      }
    >
      <div className="flex flex-col gap-4">
        {enabled ? (
          <DefinitionList columns={4}>
            <DefinitionListItem
              surface
              label={t("credits.settings.cap")}
              value={settings.overage_cap_credits ? format.credits(settings.overage_cap_credits) : t("credits.settings.noCap")}
            />
          </DefinitionList>
        ) : null}
        <div className="flex flex-col items-start gap-2 border-t border-border pt-4 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm leading-6 text-muted-foreground">{t("credits.settings.managedBySupport")}</p>
          <Button asChild variant="ghost" size="sm" className="shrink-0">
            <Link href="/support/tickets/new">
              <LifeBuoy data-icon="inline-start" />
              {t("credits.settings.contactSupport")}
            </Link>
          </Button>
        </div>
      </div>
    </AppCard>
  );
}
