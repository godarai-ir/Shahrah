"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";
import { CalendarClock, Coins, PiggyBank, TrendingDown } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { DataTable } from "@/components/ui/data-table";
import { useAppLocale } from "@/hooks/use-locale-format";

import type { CreditUsageSummary } from "../api";
import { useBillingFormat } from "../use-billing-format";
import type { CreditBalance, CreditLot } from "./api";
import { FORECAST_WARNING_DAYS, forecastDaysLeft, netCacheSavedCredits } from "./domain";

const KNOWN_LOT_SOURCES = new Set(["plan", "trial", "promo", "purchase_addon", "migration", "compensation", "adjust", "overage"]);

type CreditOverviewProps = {
  balance: CreditBalance | null;
  /** Usage over the current month, for the consumption and cache metrics. */
  usage: CreditUsageSummary | null;
};

/** Balance, this month's consumption, cache savings, the depletion forecast, and the spendable lots. */
export function CreditOverview({ balance, usage }: CreditOverviewProps) {
  const t = useTranslations("billing");
  const format = useBillingFormat();
  const locale = useAppLocale();
  const [now] = useState(() => new Date());

  const lotColumns = useMemo<ColumnDef<CreditLot>[]>(
    () => [
      {
        accessorKey: "source",
        header: t("credits.lots.columns.source"),
        cell: ({ row }) => t(KNOWN_LOT_SOURCES.has(row.original.source) ? `credits.lots.sources.${row.original.source}` : "credits.lots.sources.other"),
      },
      {
        accessorKey: "remaining_credits",
        header: t("credits.lots.columns.remaining"),
        cell: ({ row }) => <span className="tabular-nums">{format.credits(row.original.remaining_credits)}</span>,
      },
      {
        accessorKey: "amount_credits",
        header: t("credits.lots.columns.amount"),
        cell: ({ row }) => <span className="tabular-nums text-muted-foreground">{format.credits(row.original.amount_credits)}</span>,
      },
      {
        accessorKey: "expires_at",
        header: t("credits.lots.columns.expires"),
        cell: ({ row }) => row.original.expires_at
          ? <DateTimeValue value={row.original.expires_at} locale={locale} format={{ dateStyle: "medium" }} />
          : <span className="text-muted-foreground">{t("credits.lots.noExpiry")}</span>,
      },
    ],
    [format, locale, t],
  );

  if (!balance) {
    return (
      <Alert variant="destructive">
        <AlertTitle>{t("errorTitle")}</AlertTitle>
        <AlertDescription>{t("credits.overview.loadError")}</AlertDescription>
      </Alert>
    );
  }

  const debt = Number(balance.debt_credits) || 0;
  const forecast = balance.forecast;
  const daysLeft = forecastDaysLeft(forecast, now);
  const forecastValue = daysLeft === null
    ? t("credits.forecast.noUsage")
    : daysLeft === 0 ? t("credits.forecast.today") : t("credits.forecast.days", { days: daysLeft });
  const forecastDescription = forecast.limited_by === "expiry"
    ? t("credits.forecast.expiryDescription")
    : forecast.status === "no_usage"
      ? t("credits.forecast.noUsageDescription", { days: forecast.window_days })
      : t("credits.forecast.basis", { amount: format.credits(forecast.daily_average_credits), days: forecast.window_days });

  return (
    <div className="flex flex-col gap-6">
      {debt > 0 ? (
        <Alert variant="destructive">
          <AlertTitle>{t("credits.overview.debtTitle")}</AlertTitle>
          <AlertDescription>{t("credits.overview.debtDescription", { amount: format.credits(balance.debt_credits) })}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4 md:gap-4">
        <MetricCard
          icon={Coins}
          label={t("credits.overview.available")}
          value={format.credits(balance.available_credits)}
          description={t("credits.overview.availableDescription", { count: balance.lots.length })}
        />
        <MetricCard
          icon={TrendingDown}
          label={t("credits.overview.consumed")}
          value={usage?.available ? format.credits(usage.total_credits) : t("credits.overview.unavailable")}
          description={usage?.available ? t("credits.overview.requestCount", { count: usage.request_count }) : undefined}
        />
        <MetricCard
          icon={PiggyBank}
          label={t("credits.overview.cacheSaved")}
          value={usage?.available ? format.credits(netCacheSavedCredits(usage)) : t("credits.overview.unavailable")}
          description={usage?.available ? t("credits.overview.cacheSavedGross", { amount: format.credits(usage.cache_saved_credits) }) : undefined}
        />
        <MetricCard
          icon={CalendarClock}
          tone={daysLeft !== null && daysLeft <= FORECAST_WARNING_DAYS ? "warning" : "default"}
          label={t("credits.forecast.title")}
          value={forecastValue}
          description={forecastDescription}
          trend={forecast.status === "uncertain" && forecast.limited_by !== "expiry" ? <StatusBadge status="warning">{t("credits.forecast.uncertain")}</StatusBadge> : undefined}
        />
      </div>

      <AppCard title={t("credits.lots.title")} description={t("credits.lots.description")}>
        <DataTable tableId="billing-credit-lots" columns={lotColumns} data={balance.lots} emptyLabel={t("credits.lots.empty")} />
      </AppCard>
    </div>
  );
}
