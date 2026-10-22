"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { Armchair } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

import type { SeatSummary } from "./seats/api";
import { useBillingFormat } from "./use-billing-format";

type SubscriptionOverviewProps = {
  summary: SeatSummary | null;
  initialError?: string | null;
};

/**
 * The tenant's subscription: the free plan, or the seat subscription that its
 * first paid seat started. Seats are bought and renewed on the seats page.
 */
export function SubscriptionOverview({ summary, initialError }: SubscriptionOverviewProps) {
  const t = useTranslations("billingPlan.subscription");
  const tIntervals = useTranslations("billingSeats.intervals");
  const format = useBillingFormat();

  if (!summary) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{initialError ?? t("loadError")}</AlertDescription>
      </Alert>
    );
  }

  const seatsRunning = summary.subscription_kind === "seats";
  const purchased = (type: "pro" | "max") => summary.paid.find((item) => item.type === type)?.purchased ?? 0;
  const hasDetails = seatsRunning || summary.free_limit !== null;

  // The free plan's next step is buying a first paid seat, so its action is the
  // primary one; a running subscription only needs a way to the seats page.
  return (
    <AppCard
      title={seatsRunning ? t("seats") : t("free")}
      description={seatsRunning ? t("seatsDescription") : t("freeDescription")}
      actions={
        <Button asChild variant={seatsRunning ? "outline" : "default"}>
          <Link href="/billing/seats">
            <Armchair data-icon="inline-start" />
            {seatsRunning ? t("manageSeats") : t("buySeats")}
          </Link>
        </Button>
      }
    >
      <div className="flex flex-col gap-4">
        {hasDetails ? (
          <DefinitionList columns={4}>
            {seatsRunning ? (
              <DefinitionListItem
                surface
                label={t("paidSeats")}
                value={<span className="tabular-nums">{t("paidSeatsValue", { pro: purchased("pro"), max: purchased("max") })}</span>}
              />
            ) : null}
            {summary.free_limit !== null ? (
              <DefinitionListItem
                surface
                label={t("freeSeats")}
                value={<span className="tabular-nums">{t("freeSeatsValue", { used: summary.free_used, total: summary.free_limit })}</span>}
              />
            ) : null}
            {seatsRunning ? (
              <>
                <DefinitionListItem surface label={t("interval")} value={tIntervals(summary.billing_interval === "yearly" ? "yearly" : "monthly")} />
                <DefinitionListItem surface label={t("periodEnd")} value={summary.period_end ? format.date(summary.period_end) : "—"} />
              </>
            ) : null}
          </DefinitionList>
        ) : null}

        {summary.renewal_due && summary.period_end ? (
          <Alert>
            <AlertDescription>{t("renewalDue", { date: format.date(summary.period_end) })}</AlertDescription>
          </Alert>
        ) : null}

        <p className="text-sm leading-6 text-muted-foreground">{seatsRunning ? t("noRenewal") : t("startHint")}</p>
      </div>
    </AppCard>
  );
}
