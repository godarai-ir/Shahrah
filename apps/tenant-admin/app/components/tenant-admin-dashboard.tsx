"use client";

import Link from "next/link";
import { type ComponentProps, useCallback, useMemo, useState } from "react";
import {
  ArrowUpRight,
  BookOpen,
  Gauge,
  type LucideIcon,
  RefreshCcw,
} from "lucide-react";
import { useTranslations } from "next-intl";

import { CompactValue, MetricCard } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCompactNumber } from "@/lib/compact-number";
import { cn } from "@/lib/utils";
import { DashboardAnalytics } from "./dashboard-analytics";

type TenantSubscription = {
  plan_slug?: string;
  plan_display_name?: string;
  status: string;
  billing_interval: string;
  current_period_start?: string;
  current_period_end?: string;
  trial_started_at?: string;
  trial_ends_at?: string;
};

type UsageCounter = {
  feature_key: string;
  used: number;
  limit?: number | null;
  window_end?: string;
};

type EffectiveEntitlements = {
  subscription?: TenantSubscription | null;
  features?: Record<string, unknown>;
  overrides?: Record<string, unknown>;
  usage_counters?: UsageCounter[];
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
        requestJSON<TenantSubscription>("/api/subscription"),
        requestJSON<EffectiveEntitlements>("/api/entitlements"),
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
    <div className="flex flex-col gap-6">
      <section className="px-1 py-1 sm:px-2 sm:py-2">
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(20rem,24rem)] xl:items-start">
          <div className="flex min-w-0 flex-col gap-5">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex flex-wrap items-center gap-2">
                  <Badge variant={planVisual.badgeVariant}>{planStatusLabel}</Badge>
                  <Badge variant="outline">{planIntervalLabel}</Badge>
                  {daysRemaining !== null ? (
                    <Badge variant="secondary">
                      {t("plan.daysRemainingBadge", { days: numberFormatter.format(daysRemaining) })}
                    </Badge>
                  ) : null}
              </div>

              <div className="flex flex-wrap items-center gap-2 sm:justify-end">
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={() => void loadDashboard()}
                  disabled={isLoading}
                  aria-label={isLoading ? t("refreshing") : common("refresh")}
                >
                  <RefreshCcw data-icon="inline-start" aria-hidden="true" />
                </Button>
                <Button type="button" asChild>
                  <Link href="/billing/plan">
                    {planActionLabel}
                    <ArrowUpRight data-icon="inline-end" aria-hidden="true" />
                  </Link>
                </Button>
              </div>
            </div>

            <div className="grid gap-4">
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
            </div>
          </div>

          <DashboardPanel className="border-info/30 bg-info/5">
            <CardHeader className="gap-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <span className="flex size-9 items-center justify-center rounded-xl bg-info/12 text-info">
                  <BookOpen className="size-4" aria-hidden="true" />
                </span>
                {t("quickStart.title")}
              </CardTitle>
              <CardDescription className="max-w-2xl leading-6">{t("quickStart.description")}</CardDescription>
            </CardHeader>
            <CardContent className="pt-0">
              <Button type="button" variant="outline" asChild>
                <Link href="/quick-start">
                  {t("quickStart.action")}
                  <ArrowUpRight data-icon="inline-end" aria-hidden="true" />
                </Link>
              </Button>
            </CardContent>
          </DashboardPanel>
        </div>
      </section>

      {banner ? (
        <Alert variant="destructive">
          <AlertTitle>{t("messages.errorTitle")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <DashboardAnalytics
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
      className="card-hover"
    />
  );
}

function DashboardPanel(props: ComponentProps<typeof Card>) {
  return (
    <Card
      {...props}
      className={cn(
        "overflow-hidden rounded-2xl border border-border/70 bg-card/95 shadow-sm backdrop-blur-sm",
        props.className,
      )}
    />
  );
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(path, {
    ...options,
    headers,
    cache: "no-store",
  });

  const text = await response.text();
  const data = parseResponsePayload(text);
  if (!response.ok) {
    throw new Error(extractMessageFromPayload(data) ?? `خطای درخواست (${response.status})`);
  }

  return data as T;
}

function parseResponsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessageFromPayload(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
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
