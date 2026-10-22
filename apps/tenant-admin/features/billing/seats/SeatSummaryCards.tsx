"use client";

import { useTranslations } from "next-intl";
import { CalendarClock, Crown, Sparkles, UserRound, UserX } from "lucide-react";

import { MetricCard } from "@/components/app-ui/MetricCard";

import { useBillingFormat } from "../use-billing-format";
import type { PaidSeatType, SeatSummary } from "./api";
import { freeSeatsLeft } from "./domain";

/** The seat overview: free seats left, paid seats in use, and the renewal date. */
export function SeatSummaryCards({ summary }: { summary: SeatSummary }) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const freeLeft = freeSeatsLeft(summary);
  const seatsRunning = summary.subscription_kind === "seats";

  // A paid type nobody bought says so instead of "0 of 0 · 0 available".
  const paidCard = (type: PaidSeatType) => {
    const item = summary.paid.find((paid) => paid.type === type);
    const purchased = item?.purchased ?? 0;

    return purchased === 0
      ? { value: format.count(0), description: t("summary.notBought") }
      : {
          value: t("summary.ofTotal", { used: format.count(item?.assigned ?? 0), total: format.count(purchased) }),
          description: t("summary.available", { count: item?.available ?? 0 }),
        };
  };
  const pro = paidCard("pro");
  const max = paidCard("max");

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-5">
      <MetricCard
        density="compact"
        icon={UserRound}
        label={t("summary.free")}
        value={
          summary.free_limit === null
            ? format.count(summary.free_used)
            : t("summary.ofTotal", { used: format.count(summary.free_used), total: format.count(summary.free_limit) })
        }
        description={freeLeft === null ? undefined : t("summary.freeLeft", { count: freeLeft })}
        tone={freeLeft === 0 ? "warning" : "default"}
      />
      <MetricCard density="compact" icon={Sparkles} label={t("summary.pro")} value={pro.value} description={pro.description} />
      <MetricCard density="compact" icon={Crown} label={t("summary.max")} value={max.value} description={max.description} />
      <MetricCard
        density="compact"
        icon={UserX}
        label={t("summary.unassigned")}
        value={format.count(summary.unassigned)}
        description={summary.unassigned > 0 ? t("summary.unassignedHint") : t("summary.allAssigned")}
        tone={summary.unassigned > 0 && summary.enforced ? "warning" : "muted"}
      />
      <MetricCard
        density="compact"
        icon={CalendarClock}
        label={t("summary.periodEnd")}
        value={seatsRunning && summary.period_end ? format.date(summary.period_end) : "—"}
        description={
          seatsRunning ? t(`intervals.${summary.billing_interval === "yearly" ? "yearly" : "monthly"}`) : t("summary.noSubscription")
        }
        tone="muted"
      />
    </div>
  );
}
