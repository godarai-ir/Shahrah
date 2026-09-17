"use client";

import Link from "next/link";
import dynamic from "next/dynamic";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ArrowUpRight, CircleAlert, Rocket, TriangleAlert } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { DashboardAlertList, type DashboardAlertItem } from "@/components/app-ui/DashboardAlert";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";
import { cn } from "@/lib/utils";
import {
  getDashboardEntitlements,
  getDashboardSubscription,
  getDashboardTimeseries,
  type AnalyticsPoint,
} from "./api";
import type { EffectiveEntitlements, TenantSubscription } from "./contracts";
import {
  DAY_MS,
  DEFAULT_ANALYTICS_RANGE,
  HOUR_MS,
  forecastPeriodEndUsage,
  parseValidDate,
  readStoredAnalyticsRange,
  storeAnalyticsRange,
  summarizeDailyCost,
  summarizeErrorRate,
  summarizeLatencyP95,
  type AnalyticsRangeKey,
} from "./dashboard-metrics";
import { Sparkline } from "./Sparkline";

const DynamicDashboardAnalytics = dynamic(
  () => import("./DashboardAnalytics").then((module) => module.DashboardAnalytics),
  {
    ssr: false,
    loading: () => <DashboardAnalyticsPlaceholder />,
  },
);

type DashboardAnalyticsProps = {
  range: AnalyticsRangeKey;
  onRangeChange: (range: AnalyticsRangeKey) => void;
  refreshKey: number;
};

type TenantAdminDashboardProps = {
  initialSubscription: TenantSubscription | null;
  initialEntitlements: EffectiveEntitlements | null;
  initialError: string | null;
};

type SignalsState =
  | { status: "loading" }
  | { status: "failed" }
  | { status: "ready"; hourly: AnalyticsPoint[]; daily: AnalyticsPoint[] };

const REQUESTS_FEATURE_KEY = "gateway.requests.monthly";
const QUOTA_WARNING_RATIO = 0.85;
/** Daily history for the cost card (9 days) and the Theil–Sen forecast (14 complete days). */
const DAILY_HISTORY_DAYS = 15;
const ERROR_LOGS_HREF = "/logs?status=4xx,5xx";

export default function TenantAdminDashboard(props: TenantAdminDashboardProps) {
  const t = useTranslations("dashboard");
  const common = useTranslations("common");
  const locale = useLocale();

  const [subscription, setSubscription] = useState<TenantSubscription | null>(
    props.initialSubscription ?? props.initialEntitlements?.subscription ?? null,
  );
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements | null>(props.initialEntitlements);
  const [loadError, setLoadError] = useState<string | null>(props.initialError);
  const [isLoading, setIsLoading] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const [analyticsRange, setAnalyticsRange] = useState<AnalyticsRangeKey>(DEFAULT_ANALYTICS_RANGE);
  const signals = useDashboardSignals(refreshKey);

  useEffect(() => {
    // Read after mount so server and client markup match.
    const stored = readStoredAnalyticsRange();
    if (stored !== DEFAULT_ANALYTICS_RANGE) queueMicrotask(() => setAnalyticsRange(stored));
  }, []);

  const handleRangeChange = useCallback((range: AnalyticsRangeKey) => {
    setAnalyticsRange(range);
    storeAnalyticsRange(range);
  }, []);

  const planStatus = subscription?.status?.trim().toLowerCase() ?? "";
  const planName =
    subscription?.plan_display_name?.trim() ||
    subscription?.plan_slug?.trim() ||
    t("plan.currentPlan");
  const planStatusLabel = t(`plan.status.${normalizePlanStatus(planStatus)}`);
  const planDeadlineISO =
    planStatus === "trialing"
      ? subscription?.trial_ends_at?.trim() || subscription?.current_period_end?.trim() || ""
      : subscription?.current_period_end?.trim() || "";
  const daysRemaining = useMemo(() => calculateDaysRemaining(planDeadlineISO), [planDeadlineISO]);
  const planMonths = billingIntervalToMonths(subscription?.billing_interval);
  const requestsPerMonthLimit = useMemo(
    () => coerceNumericEntitlement(entitlements?.features?.[REQUESTS_FEATURE_KEY]),
    [entitlements?.features],
  );
  const currentRequestCounter = useMemo(
    () => entitlements?.usage_counters?.find((item) => item.feature_key === REQUESTS_FEATURE_KEY) ?? null,
    [entitlements?.usage_counters],
  );
  const requestLimit =
    typeof currentRequestCounter?.limit === "number"
      ? currentRequestCounter.limit
      : requestsPerMonthLimit === null
        ? null
        : requestsPerMonthLimit * planMonths;
  const requestUsage = currentRequestCounter?.used ?? null;
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
      setSubscription(subscriptionData ?? entitlementsData.subscription ?? null);
      setEntitlements(entitlementsData);
      setLoadError(null);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : common("requestError"));
    } finally {
      setIsLoading(false);
      setRefreshKey((current) => current + 1);
    }
  }, [common]);

  const errorRate = useMemo(
    () => (signals.status === "ready" ? summarizeErrorRate(signals.hourly) : null),
    [signals],
  );
  const latency = useMemo(
    () => (signals.status === "ready" ? summarizeLatencyP95(signals.hourly) : null),
    [signals],
  );
  const dailyCost = useMemo(
    () => (signals.status === "ready" ? summarizeDailyCost(signals.daily) : null),
    [signals],
  );
  const projectedUsage = useMemo(
    () =>
      signals.status === "ready"
        ? forecastPeriodEndUsage({
            dailyPoints: signals.daily,
            currentUsage: requestUsage,
            periodEnd: parseValidDate(subscription?.current_period_end),
          })
        : null,
    [requestUsage, signals, subscription?.current_period_end],
  );

  const percentFormatter = useMemo(
    () => new Intl.NumberFormat(locale, { minimumFractionDigits: 1, maximumFractionDigits: 1 }),
    [locale],
  );
  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);

  const alerts = useMemo<DashboardAlertItem[]>(() => {
    const items: DashboardAlertItem[] = [];
    if (loadError) {
      items.push({ id: "load-failed", tone: "error", icon: CircleAlert, title: t("alerts.loadFailed.title"), detail: loadError });
    }
    if (errorRate?.spike) {
      items.push({
        id: "error-rate-spike",
        tone: "error",
        icon: TriangleAlert,
        title: t("alerts.errorRateSpike.title", { rate: percentFormatter.format(errorRate.spike.lastRate * 100) }),
        detail: t("alerts.errorRateSpike.detail", { baseline: percentFormatter.format(errorRate.spike.baseline * 100) }),
        action: { label: t("alerts.errorRateSpike.action"), href: ERROR_LOGS_HREF },
      });
    }
    return items;
  }, [errorRate, loadError, percentFormatter, t]);

  const quotaRatio = requestLimit && requestUsage !== null ? clampRatio(requestUsage / requestLimit) : null;
  const signalsLoading = signals.status === "loading";
  const kpiCount = (signalsLoading || errorRate ? 1 : 0) + (latency ? 1 : 0) + (signalsLoading || dailyCost ? 1 : 0) + 1;

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <PageHeader
        title={t("title")}
        description={t("subtitle")}
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
      />

      <DashboardAlertList items={alerts} moreLabel={(count) => t("alerts.more", { count })} />

      <section
        aria-label={t("metrics.sectionLabel")}
        className={cn("grid gap-4 sm:grid-cols-2", kpiCount >= 4 ? "xl:grid-cols-4" : "xl:grid-cols-3")}
      >
        {signalsLoading ? <MetricCardSkeleton /> : null}
        {errorRate ? (
          <MetricCard
            label={t("metrics.errorRate.label")}
            value={<span dir="ltr" className="tabular-nums">{percentFormatter.format(errorRate.rate * 100)}%</span>}
            delta={
              errorRate.previousRate === null
                ? undefined
                : buildDelta({
                    change: (errorRate.rate - errorRate.previousRate) * 100,
                    epsilon: 0.05,
                    upIsGood: false,
                    label: (change) =>
                      t("metrics.errorRate.delta", { value: `${change > 0 ? "+" : ""}${percentFormatter.format(change)}` }),
                  })
            }
            sparkline={<Sparkline values={errorRate.sparkline} stroke="var(--destructive)" />}
            tone="muted"
          />
        ) : null}

        {latency ? (
          <MetricCard
            label={t("metrics.latencyP95.label")}
            value={<span dir="ltr" className="tabular-nums">{percentFormatter.format(latency.valueMs / 1_000)}s</span>}
            delta={
              latency.previousMs
                ? buildDelta({
                    change: ((latency.valueMs - latency.previousMs) / latency.previousMs) * 100,
                    epsilon: 0.5,
                    upIsGood: false,
                    label: (change) => `${change > 0 ? "+" : ""}${percentFormatter.format(change)}%`,
                  })
                : undefined
            }
            sparkline={<Sparkline values={latency.sparkline} stroke="var(--chart-2)" />}
            tone="muted"
          />
        ) : null}

        {signalsLoading ? <MetricCardSkeleton /> : null}
        {dailyCost ? (
          <MetricCard
            label={t("metrics.costToday.label")}
            value={<span dir="ltr" className="tabular-nums">{compactCurrency(dailyCost.today)}</span>}
            delta={{
              direction:
                dailyCost.today > dailyCost.average7d ? "up" : dailyCost.today < dailyCost.average7d ? "down" : "flat",
              tone: "muted",
              label: (
                <>
                  {t("metrics.costToday.average")} <span dir="ltr">{compactCurrency(dailyCost.average7d)}</span>
                </>
              ),
            }}
            sparkline={<Sparkline values={dailyCost.sparkline} stroke="var(--chart-3)" variant="bars" />}
            tone="muted"
          />
        ) : null}

        <MetricCard
          label={t("metrics.quota.label")}
          value={
            <span dir="ltr" className="tabular-nums">
              {quotaRatio === null ? compactNumber(requestUsage ?? 0) : `${percentFormatter.format(quotaRatio * 100)}%`}
            </span>
          }
          delta={
            daysRemaining === null
              ? undefined
              : {
                  tone: daysRemaining <= 3 ? "negative" : "muted",
                  label: t("metrics.quota.daysRemaining", { days: numberFormatter.format(daysRemaining) }),
                }
          }
          sparkline={
            <div className="flex flex-col gap-1.5">
              {quotaRatio !== null ? (
                <div
                  className="h-2 overflow-hidden rounded-full bg-muted"
                  role="progressbar"
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={Math.round(quotaRatio * 100)}
                  aria-label={t("metrics.quota.label")}
                >
                  <div
                    className={cn("h-full rounded-full", quotaRatio >= QUOTA_WARNING_RATIO ? "bg-warning" : "bg-chart-2")}
                    style={{ width: `${quotaRatio * 100}%` }}
                  />
                </div>
              ) : null}
              <p className="text-[11px] leading-5 text-muted-foreground">
                {t("metrics.quota.progressHint", {
                  used: compactNumber(requestUsage ?? 0),
                  limit: requestLimit === null ? "∞" : compactNumber(requestLimit),
                })}
                {projectedUsage !== null ? (
                  <> · {t("metrics.quota.forecast", { value: compactNumber(projectedUsage) })}</>
                ) : null}
              </p>
            </div>
          }
          description={`${planName} · ${planStatusLabel}`}
          tone={quotaRatio !== null && quotaRatio >= QUOTA_WARNING_RATIO ? "warning" : "muted"}
        />
      </section>

      <OnboardingBar
        title={t("onboarding.title")}
        description={t("onboarding.description")}
        continueLabel={t("onboarding.continue")}
      />

      <DeferredDashboardAnalytics range={analyticsRange} onRangeChange={handleRangeChange} refreshKey={refreshKey} />
    </div>
  );
}

function useDashboardSignals(refreshKey: number): SignalsState {
  const [state, setState] = useState<SignalsState>({ status: "loading" });

  useEffect(() => {
    const controller = new AbortController();
    const now = Date.now();
    const to = new Date(now).toISOString();
    // Two full days of hourly buckets: the current 24h window plus the one it is compared with.
    const hourlyFrom = new Date(Math.floor(now / HOUR_MS) * HOUR_MS - 49 * HOUR_MS).toISOString();
    const dailyFrom = new Date(Math.floor(now / DAY_MS) * DAY_MS - DAILY_HISTORY_DAYS * DAY_MS).toISOString();

    void Promise.all([
      getDashboardTimeseries({ from: hourlyFrom, to, intervalSeconds: 3_600 }, controller.signal),
      getDashboardTimeseries({ from: dailyFrom, to, intervalSeconds: 86_400 }, controller.signal),
    ])
      .then(([hourly, daily]) => setState({ status: "ready", hourly, daily }))
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === "AbortError") return;
        setState({ status: "failed" });
      });

    return () => controller.abort();
  }, [refreshKey]);

  return state;
}

/**
 * Onboarding strip. `features/quick-start` derives step status from server-side
 * resource counts with no client hook, so steps are out of scope for this phase
 * and the strip links to the full quick start instead.
 */
function OnboardingBar({ title, description, continueLabel }: { title: string; description: string; continueLabel: string }) {
  return (
    <section
      data-slot="dashboard-onboarding"
      className="flex flex-col gap-3 rounded-xl border border-info-soft-border bg-info-soft/35 px-4 py-3 sm:flex-row sm:items-center"
    >
      <span className="flex size-8 shrink-0 items-center justify-center rounded-md border border-info-soft-border bg-info-soft text-info-soft-foreground">
        <Rocket className="size-4" aria-hidden="true" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold text-foreground">{title}</p>
        <p className="text-xs leading-6 text-muted-foreground">{description}</p>
      </div>
      <Button type="button" variant="outline" size="sm" asChild>
        <Link href="/quick-start">
          {continueLabel}
          <ArrowUpRight data-icon="inline-end" aria-hidden="true" />
        </Link>
      </Button>
    </section>
  );
}

function buildDelta(input: {
  change: number;
  epsilon: number;
  upIsGood: boolean;
  label: (change: number) => string;
}) {
  const direction = Math.abs(input.change) < input.epsilon ? "flat" : input.change > 0 ? "up" : "down";
  const tone =
    direction === "flat" ? "muted" : (direction === "up") === input.upIsGood ? "positive" : "negative";
  return { direction, tone, label: input.label(input.change) } as const;
}

function MetricCardSkeleton() {
  return <Skeleton className="h-36 rounded-xl" aria-hidden="true" />;
}

function DashboardAnalyticsPlaceholder() {
  return (
    <section className="grid min-h-72 gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]" aria-hidden="true">
      <div className="min-h-72 animate-pulse rounded-xl border bg-muted/35" />
      <div className="min-h-72 animate-pulse rounded-xl border bg-muted/35" />
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

function calculateDaysRemaining(deadlineISO?: string): number | null {
  if (!deadlineISO) return null;
  const deadline = Date.parse(deadlineISO);
  if (Number.isNaN(deadline)) return null;
  return Math.max(0, Math.ceil((deadline - Date.now()) / DAY_MS));
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

function coerceNumericEntitlement(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}

function clampRatio(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.max(0, Math.min(1, value));
}

function normalizePlanStatus(status: string): "active" | "trialing" | "past_due" | "inactive" {
  return status === "active" || status === "trialing" || status === "past_due" ? status : "inactive";
}

function compactNumber(value: number) {
  return formatCompactNumber(value, { locale: "en-US", compactThreshold: 1_000, maximumFractionDigits: 1 }).value;
}

function compactCurrency(value: number) {
  return formatCompactCurrency(value, {
    locale: "en-US",
    currency: "USD",
    compactThreshold: 1_000,
    maximumFractionDigits: 2,
  }).value;
}
