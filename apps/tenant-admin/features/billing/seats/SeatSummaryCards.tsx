"use client";

import { useTranslations } from "next-intl";
import { CalendarClock, Crown, Sparkles, UserRound, UserX } from "lucide-react";

import { MetricCard } from "@/components/app-ui/MetricCard";

import { useBillingFormat } from "../use-billing-format";
import type { SeatSummary } from "./api";
import { freeSeatsLeft } from "./domain";

/** The seat overview: free seats left, paid seats in use, and the renewal date. */
export function SeatSummaryCards({ summary }: { summary: SeatSummary }) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const pro = summary.paid.find((item) => item.type === "pro");
  const max = summary.paid.find((item) => item.type === "max");
  const freeLeft = freeSeatsLeft(summary);

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
        description={freeLeft === null ? t("summary.freeUnlimited") : t("summary.freeLeft", { count: freeLeft })}
        tone={freeLeft === 0 ? "warning" : "default"}
      />
      <MetricCard
        density="compact"
        icon={Sparkles}
        label={t("summary.pro")}
        value={t("summary.ofTotal", { used: format.count(pro?.assigned ?? 0), total: format.count(pro?.purchased ?? 0) })}
        description={t("summary.available", { count: pro?.available ?? 0 })}
      />
      <MetricCard
        density="compact"
        icon={Crown}
        label={t("summary.max")}
        value={t("summary.ofTotal", { used: format.count(max?.assigned ?? 0), total: format.count(max?.purchased ?? 0) })}
        description={t("summary.available", { count: max?.available ?? 0 })}
      />
      <MetricCard
        density="compact"
        icon={UserX}
        label={t("summary.unassigned")}
        value={format.count(summary.unassigned)}
        description={t("summary.unassignedHint")}
        tone={summary.unassigned > 0 && summary.enforced ? "warning" : "muted"}
      />
      <MetricCard
        density="compact"
        icon={CalendarClock}
        label={t("summary.periodEnd")}
        value={summary.period_end ? format.date(summary.period_end) : "—"}
        description={summary.billing_interval ? t(`intervals.${summary.billing_interval === "yearly" ? "yearly" : "monthly"}`) : t("summary.noSubscription")}
        tone="muted"
      />
    </div>
  );
}
