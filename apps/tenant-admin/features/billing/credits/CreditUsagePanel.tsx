"use client";

import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { ModelNameValue } from "@/components/app-ui/ModelNameValue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useNumberFormatter } from "@/hooks/use-locale-format";

import type { CreditUsageItem, CreditUsageSummary } from "../api";
import { useBillingFormat } from "../use-billing-format";
import { CREDIT_USAGE_PRESETS, getCreditUsageSummary, type CreditUsagePreset } from "./api";
import { creditShare } from "./domain";

const GROUPS = ["model", "team", "user", "product"] as const;
type UsageGroup = (typeof GROUPS)[number];

const KNOWN_PRODUCTS = new Set(["gateway", "librechat", "guardrail", "semantic_cache", "embeddings", "ocr", "playground", "oidc_app"]);

type CreditUsagePanelProps = {
  /** The current month's summary, loaded with the page. */
  initialSummary: CreditUsageSummary | null;
  /** The public pricing changelog, where credit price and rate-card changes are announced. */
  changelogHref?: string;
};

/** Credit consumption over a report window, broken down by model, team, user, and product. */
export function CreditUsagePanel({ initialSummary, changelogHref }: CreditUsagePanelProps) {
  const t = useTranslations("billing");
  const format = useBillingFormat();
  const percent = useNumberFormatter({ style: "percent", maximumFractionDigits: 0 });
  const [preset, setPreset] = useState<CreditUsagePreset>("current_month");
  const [summary, setSummary] = useState<CreditUsageSummary | null>(initialSummary);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(initialSummary ? null : t("credits.usage.loadError"));
  // The period whose summary is shown, so a re-run effect does not refetch it.
  const loadedPreset = useRef<CreditUsagePreset | null>(initialSummary ? "current_month" : null);

  useEffect(() => {
    if (preset === loadedPreset.current) return;
    const controller = new AbortController();
    setLoading(true);
    setError(null);
    getCreditUsageSummary(preset, controller.signal)
      .then((next) => {
        loadedPreset.current = preset;
        setSummary(next);
      })
      .catch((err: unknown) => {
        if (!controller.signal.aborted) setError(err instanceof Error ? err.message : t("unknownError"));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [preset, t]);

  const columns = useMemo(() => {
    const nameCell = (group: UsageGroup, item: CreditUsageItem): ReactNode => {
      if (group === "model") return <ModelNameValue value={item.key} />;
      if (group === "product") {
        return t(KNOWN_PRODUCTS.has(item.key) ? `credits.usage.products.${item.key}` : "credits.usage.products.other");
      }
      const fallback = group === "team" ? t("credits.usage.noTeam") : t("credits.usage.noUser");
      return <span className="font-medium text-foreground" dir="auto">{item.label || item.key || fallback}</span>;
    };
    const total = summary?.total_credits ?? "0";

    return Object.fromEntries(GROUPS.map((group) => [group, [
      { id: "name", header: t(`credits.usage.columns.${group}`), cell: ({ row }) => nameCell(group, row.original) },
      {
        accessorKey: "credits",
        header: t("credits.usage.columns.credits"),
        cell: ({ row }) => <span className="tabular-nums">{format.credits(row.original.credits)}</span>,
      },
      {
        id: "share",
        header: t("credits.usage.columns.share"),
        cell: ({ row }) => {
          const share = creditShare(row.original.credits, total);
          return (
            <div className="flex min-w-24 items-center gap-2">
              <Progress value={share * 100} aria-label={t("credits.usage.columns.share")} className="h-1.5" />
              <span className="text-xs tabular-nums text-muted-foreground">{percent.format(share)}</span>
            </div>
          );
        },
      },
      {
        accessorKey: "request_count",
        header: t("credits.usage.columns.requests"),
        cell: ({ row }) => <span className="tabular-nums">{format.count(row.original.request_count)}</span>,
      },
    ] satisfies ColumnDef<CreditUsageItem>[]])) as Record<UsageGroup, ColumnDef<CreditUsageItem>[]>;
  }, [format, percent, summary?.total_credits, t]);

  const rows: Record<UsageGroup, CreditUsageItem[]> = {
    model: summary?.by_model ?? [],
    team: summary?.by_team ?? [],
    user: summary?.by_user ?? [],
    product: summary?.by_product ?? [],
  };

  return (
    <AppCard
      title={t("credits.usage.title")}
      description={summary?.available === false ? t("credits.usage.unavailable") : t("credits.usage.description")}
      actions={
        <AppSelectField
          ariaLabel={t("credits.usage.preset")}
          size="sm"
          className="min-w-44"
          value={preset}
          disabled={loading}
          onValueChange={(value) => setPreset(value as CreditUsagePreset)}
          options={CREDIT_USAGE_PRESETS.map((value) => ({ value, label: t(`credits.usage.presets.${value}`) }))}
        />
      }
      footer={changelogHref ? (
        <Button asChild variant="link" className="h-auto px-0">
          <a href={changelogHref} target="_blank" rel="noreferrer">{t("credits.usage.changelog")}</a>
        </Button>
      ) : undefined}
    >
      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : summary && summary.request_count === 0 && !loading ? (
        <EmptyState variant="inline" density="compact" title={t("credits.usage.emptyTitle")} description={t("credits.usage.emptyDescription")} />
      ) : (
        <Tabs defaultValue="model" className="flex flex-col gap-4">
          <TabsList>
            {GROUPS.map((group) => <TabsTrigger key={group} value={group}>{t(`credits.usage.tabs.${group}`)}</TabsTrigger>)}
          </TabsList>
          {GROUPS.map((group) => (
            <TabsContent key={group} value={group}>
              <DataTable
                tableId={`billing-credit-usage-${group}`}
                columns={columns[group]}
                data={rows[group]}
                state={loading ? "loading" : "ready"}
                emptyLabel={t("credits.usage.emptyGroup")}
              />
            </TabsContent>
          ))}
        </Tabs>
      )}
    </AppCard>
  );
}
