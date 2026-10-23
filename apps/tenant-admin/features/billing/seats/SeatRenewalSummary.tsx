"use client";

import { useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { CalendarClock } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { Button } from "@/components/ui/button";

import { useBillingFormat } from "../use-billing-format";
import type { SeatSummary } from "./api";
import { renewalWindowOpen } from "./domain";

/**
 * Where the seats page points at renewal: whether it is open, how many paid
 * seats of each type are kept, and the way to the renewal page, where renewing
 * is done. Nothing renews or is charged from here.
 */
export function SeatRenewalSummary({ summary }: { summary: SeatSummary }) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const [now] = useState(() => Date.now());

  const held = summary.paid.filter((item) => item.purchased > 0);
  if (summary.subscription_kind !== "seats" || !summary.period_end || held.length === 0) return null;

  const open = renewalWindowOpen(summary, now);
  const pending = open && Boolean(summary.renewal_due);

  const description = pending
    ? t("renewal.open", { date: format.date(summary.period_end) })
    : open
      ? t("renewal.done")
      : t("renewal.closed", { date: summary.renewal_opens_at ? format.date(summary.renewal_opens_at) : format.date(summary.period_end) });

  return (
    <AppCard
      title={t("renewal.title")}
      description={description}
      actions={
        <Button asChild variant={pending ? "default" : "outline"}>
          <Link href="/billing/seats/renew">
            <CalendarClock data-icon="inline-start" />
            {pending ? t("renewal.action") : t("renewal.view")}
          </Link>
        </Button>
      }
    >
      <DefinitionList columns={4}>
        {held.map((item) => (
          <DefinitionListItem
            key={item.type}
            surface
            label={t("renewal.seatLabel", { seat: t(`seats.${item.type}`) })}
            value={<span className="tabular-nums">{t("renewal.held", { held: item.purchased, assigned: item.assigned })}</span>}
          />
        ))}
        <DefinitionListItem surface label={t("renewal.periodEnd")} value={format.date(summary.period_end)} />
      </DefinitionList>
    </AppCard>
  );
}
