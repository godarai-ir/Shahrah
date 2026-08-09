"use client";

import Link from "next/link";
import dynamic from "next/dynamic";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ArrowUpRight,
  BookOpen,
  Gauge,
  type LucideIcon,
  Users,
} from "lucide-react";
import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatCompactNumber } from "@/lib/compact-number";
import { getDashboardEntitlements, getDashboardSubscription } from "./api";
import type { EffectiveEntitlements, TenantSubscription } from "./contracts";

const DynamicDashboardAnalytics = dynamic(
  () => import("./DashboardAnalytics").then((module) => module.DashboardAnalytics),
  {
    ssr: false,
    loading: () => <DashboardAnalyticsPlaceholder />,
  },
);

type DashboardAnalyticsProps = {
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  currentUsage: number | null;
  requestLimit: number | null;
  refreshKey: number;
};

type Banner = {
  kind: "error";
  message: string;
};

type TenantAdminDashboardProps = {
  initialSubscription: TenantSubscription | null;
  initialEntitlements: EffectiveEntitlements | null;
  initialUsersTotal: number | null;
  initialError: string | null;
};

const numberFormatter = new Intl.NumberFormat("fa-IR");
const REQUESTS_FEATURE_KEY = "gateway.requests.monthly";

export default function TenantAdminDashboard(props: TenantAdminDashboardProps) {
  const t = useTranslations("dashboard");
  const common = useTranslations("common");

  const [subscription, setSubscription] = useState<TenantSubscription | null>(
    props.initialSubscription ?? props.initialEntitlements?.subscription ?? null,
  );
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements | null>(props.initialEntitlements);
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError
      ? {
          kind: "error",
          message: props.initialError,
        }
      : null,
  );
  const [isLoading, setIsLoading] = useState(false);
  const [analyticsRefreshKey, setAnalyticsRefreshKey] = useState(0);

  const planStatus = subscription?.status?.trim().toLowerCase() ?? "";
  const planName =
    subscription?.plan_display_name?.trim() ||
    subscription?.plan_slug?.trim() ||
    t("plan.currentPlan");
  const planVisual = useMemo(() => getPlanVisual(planStatus), [planStatus]);
  const planStatusLabel = t(`plan.status.${planVisual.statusKey}`);
  const planIntervalLabel = t(`plan.interval.${normalizeBillingInterval(subscription?.billing_interval)}`);
  const planDeadlineISO =
    planStatus === "trialing"
      ? subscription?.trial_ends_at?.trim() || subscription?.current_period_end?.trim() || ""
      : subscription?.current_period_end?.trim() || "";
  const daysRemaining = useMemo(() => calculateDaysRemaining(planDeadlineISO), [planDeadlineISO]);
  const planMonths = useMemo(() => billingIntervalToMonths(subscription?.billing_interval), [subscription?.billing_interval]);
  const requestsPerMonthLimit = useMemo(
    () => coerceNumericEntitlement(entitlements?.features?.[REQUESTS_FEATURE_KEY]),
    [entitlements?.features],
  );
  const currentRequestCounter = useMemo(
    () => entitlements?.usage_counters?.find((item) => item.feature_key === REQUESTS_FEATURE_KEY) ?? null,
    [entitlements?.usage_counters],
  );
  const totalPlanRequestLimit =
    typeof currentRequestCounter?.limit === "number"
      ? currentRequestCounter.limit
      : requestsPerMonthLimit === null
        ? null
        : requestsPerMonthLimit * planMonths;
  const requestUsageValue = currentRequestCounter?.used ?? null;
  const requestUsageRatio =
    totalPlanRequestLimit && typeof requestUsageValue === "number"
      ? clampUsageRatio(requestUsageValue / totalPlanRequestLimit)
      : null;
  const planActionLabel =
    planStatus === "trialing"
      ? t("plan.actions.upgrade")
      : daysRemaining !== null && daysRemaining <= 10
        ? t("plan.actions.renew")
        : t("plan.actions.manage");

  const loadDashboard = useCallback(async () => {
    setIsLoading(true);
    try {
      const [subscriptionData, entitlementsData] = await Promise.all([
        getDashboardSubscription(),
        getDashboardEntitlements(),
      ]);

      const nextSubscription = subscriptionData ?? entitlementsData.subscription ?? null;

      setSubscription(nextSubscription);
      setEntitlements(entitlementsData);
      setBanner(null);
      setAnalyticsRefreshKey((current) => current + 1);
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, common("requestError")),
      });
    } finally {
      setIsLoading(false);
    }
  }, [common]);

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <AppCard
        title={t("overviewEyebrow")}
        description={planName}
        actions={
          <>
            <RefreshButton
              onClick={() => void loadDashboard()}
              isLoading={isLoading}
              label={common("refresh")}
              loadingLabel={t("refreshing")}
            />
            <Button type="button" asChild>
              <Link href="/billing/plan">
                {planActionLabel}
                <ArrowUpRight data-icon="inline-end" aria-hidden="true" />
              </Link>
            </Button>
          </>
        }
        className="theme-accent-line overflow-hidden"
        headerClassName="bg-neutral-soft/45"
        contentClassName="pt-3 md:pt-3"
      >
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={planVisual.badgeVariant}>{planStatusLabel}</Badge>
          <Badge variant="outline">{planIntervalLabel}</Badge>
          {daysRemaining !== null ? (
            <Badge variant="secondary">
              {t("plan.daysRemainingBadge", { days: numberFormatter.format(daysRemaining) })}
            </Badge>
          ) : null}
        </div>
      </AppCard>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <UsageSummaryCard
          icon={Gauge}
          title={t("plan.requestLimit")}
          used={requestUsageValue}
          limit={totalPlanRequestLimit}
          subtitle={t("plan.requestLimitHint", {
            monthly: formatUsagePart(requestsPerMonthLimit),
            months: numberFormatter.format(planMonths),
          })}
          tone={requestUsageRatio !== null && requestUsageRatio >= 0.85 ? "warning" : "success"}
        />

        <MetricCard
          label={t("workspaceUsers")}
          value={
            props.initialUsersTotal === null
              ? "-"
              : numberFormatter.format(props.initialUsersTotal)
          }
          description={t("workspaceUsersHint")}
          icon={Users}
          tone="default"
        />

        <AppCard
          title={
            <span className="flex items-center gap-2">
              <span className="flex size-8 items-center justify-center rounded-md border border-info-soft-border bg-info-soft text-info-soft-foreground">
                <BookOpen className="size-4" aria-hidden="true" />
              </span>
              {t("quickStart.title")}
            </span>
          }
          description={t("quickStart.description")}
          actions={
            <Button type="button" variant="outline" asChild>
              <Link href="/quick-start">
                {t("quickStart.action")}
                <ArrowUpRight data-icon="inline-end" aria-hidden="true" />
              </Link>
            </Button>
          }
          className="border-info-soft-border bg-info-soft/35 md:col-span-2 xl:col-span-1"
          headerClassName="h-full sm:items-center"
        />
      </div>

      {banner ? (
        <Alert variant="destructive">
          <AlertTitle>{t("messages.errorTitle")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <DeferredDashboardAnalytics
        currentPeriodStart={subscription?.current_period_start}
        currentPeriodEnd={subscription?.current_period_end}
        currentUsage={requestUsageValue}
        requestLimit={totalPlanRequestLimit}
        refreshKey={analyticsRefreshKey}
      />
    </div>
  );
}

function UsageSummaryCard(props: {
  title: string;
  used: number | null;
  limit: number | null;
  subtitle: string;
  icon: LucideIcon;
  tone: "success" | "warning" | "muted";
}) {
  return (
    <MetricCard
      label={props.title}
      value={<UsageLimitValue used={props.used} limit={props.limit} />}
      description={props.subtitle}
      icon={props.icon}
      tone={props.tone}
    />
  );
}

function DashboardAnalyticsPlaceholder() {
  return (
    <section className="grid min-h-96 gap-4" aria-hidden="true">
      <div className="min-h-72 animate-pulse rounded-xl border bg-muted/35" />
      <div className="grid gap-4 xl:grid-cols-2">
        <div className="min-h-64 animate-pulse rounded-xl border bg-muted/35" />
        <div className="min-h-64 animate-pulse rounded-xl border bg-muted/35" />
      </div>
    </section>
  );
}

function DeferredDashboardAnalytics(props: DashboardAnalyticsProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [shouldLoad, setShouldLoad] = useState(false);

  useEffect(() => {
    const target = containerRef.current;
    if (!target || typeof IntersectionObserver === "undefined") {
      let disposed = false;
      queueMicrotask(() => {
        if (!disposed) setShouldLoad(true);
      });
      return () => {
        disposed = true;
      };
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry?.isIntersecting) return;
        setShouldLoad(true);
        observer.disconnect();
      },
      { rootMargin: "400px 0px" },
    );
    observer.observe(target);

    return () => observer.disconnect();
  }, []);

  return (
    <div ref={containerRef}>
      {shouldLoad ? <DynamicDashboardAnalytics {...props} /> : <DashboardAnalyticsPlaceholder />}
    </div>
  );
}

function normalizeError(error: unknown, fallback: string): string {
  if (error instanceof Error) return error.message;
  return fallback;
}

function calculateDaysRemaining(deadlineISO?: string): number | null {
  if (!deadlineISO) return null;
  const deadline = Date.parse(deadlineISO);
  if (Number.isNaN(deadline)) return null;
  return Math.max(0, Math.ceil((deadline - Date.now()) / (24 * 60 * 60 * 1000)));
}

function billingIntervalToMonths(interval?: string): number {
  switch ((interval ?? "").trim().toLowerCase()) {
    case "quarterly":
      return 3;
    case "yearly":
      return 12;
    case "monthly":
    default:
      return 1;
  }
}

function normalizeBillingInterval(interval?: string): "monthly" | "quarterly" | "yearly" {
  const normalized = (interval ?? "").trim().toLowerCase();
  if (normalized === "quarterly" || normalized === "yearly") return normalized;
  return "monthly";
}

function coerceNumericEntitlement(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}

function formatUsagePart(value: number | null): string {
  return value === null ? "-" : numberFormatter.format(Math.round(value));
}

function UsageLimitValue({ used, limit }: { used: number | null; limit: number | null }) {
  return (
    <span className="inline-flex items-center gap-1 tabular-nums" dir="ltr">
      {limit === null ? (
        <span>-</span>
      ) : (
        <CompactValue
          display={formatCompactNumber(limit, { locale: "en-US", maximumFractionDigits: 0 })}
          dir="ltr"
        />
      )}
      <span>/</span>
      {used === null ? (
        <span>-</span>
      ) : (
        <CompactValue
          display={formatCompactNumber(used, { locale: "en-US", maximumFractionDigits: 0 })}
          dir="ltr"
        />
      )}
    </span>
  );
}

function clampUsageRatio(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.max(0, Math.min(1, value));
}

function getPlanVisual(status: string): {
  badgeVariant: "success" | "warning" | "outline" | "destructive";
  statusKey: "active" | "trialing" | "past_due" | "inactive";
} {
  switch (status) {
    case "trialing":
      return { badgeVariant: "warning", statusKey: "trialing" };
    case "active":
      return { badgeVariant: "success", statusKey: "active" };
    case "past_due":
      return { badgeVariant: "destructive", statusKey: "past_due" };
    default:
      return { badgeVariant: "outline", statusKey: "inactive" };
  }
}
