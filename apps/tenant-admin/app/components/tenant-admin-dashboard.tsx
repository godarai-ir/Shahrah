"use client";

import Link from "next/link";
import { type ComponentProps, useCallback, useMemo, useState } from "react";
import {
  AlertTriangle,
  ArrowUpRight,
  BookOpen,
  Clock3,
  Crown,
  FlaskConical,
  Gauge,
  type LucideIcon,
  RefreshCcw,
  ShieldCheck,
  Users,
} from "lucide-react";
import { useTranslations } from "next-intl";

import { MetricCard } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

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

type TenantAdminUserListResponse = {
  total?: number;
};

type MetricsQueryResponse = {
  data?: Array<Record<string, unknown>>;
};

type Banner = {
  kind: "error";
  message: string;
};

type TenantAdminDashboardProps = {
  initialSubscription: TenantSubscription | null;
  initialEntitlements: EffectiveEntitlements | null;
  initialUsersTotal: number | null;
  initialPlanRequestsUsed: number | null;
  initialError: string | null;
};

const numberFormatter = new Intl.NumberFormat("fa-IR");
const REQUESTS_FEATURE_KEY = "gateway.requests.monthly";
const USERS_FEATURE_KEY = "users.max";

export default function TenantAdminDashboard(props: TenantAdminDashboardProps) {
  const t = useTranslations("dashboard");
  const common = useTranslations("common");

  const [subscription, setSubscription] = useState<TenantSubscription | null>(
    props.initialSubscription ?? props.initialEntitlements?.subscription ?? null,
  );
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements | null>(props.initialEntitlements);
  const [usersTotal, setUsersTotal] = useState<number | null>(props.initialUsersTotal);
  const [planRequestsUsed, setPlanRequestsUsed] = useState<number | null>(props.initialPlanRequestsUsed);
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError
      ? {
          kind: "error",
          message: props.initialError,
        }
      : null,
  );
  const [isLoading, setIsLoading] = useState(false);

  const planStatus = subscription?.status?.trim().toLowerCase() ?? "";
  const planVisual = useMemo(() => getPlanVisual(planStatus), [planStatus]);
  const PlanStatusIcon = planVisual.icon;
  const planStatusLabel = t(`plan.status.${planVisual.statusKey}`);
  const planDisplayName = subscription?.plan_display_name?.trim() || subscription?.plan_slug?.trim() || common("notSet");
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
  const usersLimit = useMemo(
    () => coerceNumericEntitlement(entitlements?.features?.[USERS_FEATURE_KEY]),
    [entitlements?.features],
  );
  const currentRequestCounter = useMemo(
    () => entitlements?.usage_counters?.find((item) => item.feature_key === REQUESTS_FEATURE_KEY) ?? null,
    [entitlements?.usage_counters],
  );
  const totalPlanRequestLimit = requestsPerMonthLimit === null ? null : requestsPerMonthLimit * planMonths;
  const requestUsageValue =
    typeof planRequestsUsed === "number"
      ? planRequestsUsed
      : planMonths === 1
        ? currentRequestCounter?.used ?? null
        : null;
  const requestUsageRatio =
    totalPlanRequestLimit && typeof requestUsageValue === "number"
      ? clampUsageRatio(requestUsageValue / totalPlanRequestLimit)
      : null;
  const usersUsageRatio =
    usersLimit && typeof usersTotal === "number"
      ? clampUsageRatio(usersTotal / usersLimit)
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
      const [subscriptionData, entitlementsData, usersData] = await Promise.all([
        requestJSON<TenantSubscription>("/api/subscription"),
        requestJSON<EffectiveEntitlements>("/api/entitlements"),
        requestJSON<TenantAdminUserListResponse>("/api/users?limit=1&offset=0"),
      ]);

      const nextSubscription = subscriptionData ?? entitlementsData.subscription ?? null;
      const nextPlanRequestsUsed = await loadPlanRequestUsage(nextSubscription, planRequestsUsed);

      setSubscription(nextSubscription);
      setEntitlements(entitlementsData);
      setUsersTotal(usersData.total ?? 0);
      setPlanRequestsUsed(nextPlanRequestsUsed);
      setBanner(null);
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, common("requestError")),
      });
    } finally {
      setIsLoading(false);
    }
  }, [common, planRequestsUsed]);

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

            <div className="grid gap-4 sm:grid-cols-2 2xl:grid-cols-4">
              <MetricCard
                icon={PlanStatusIcon}
                label={t("plan.currentPlan")}
                value={planDisplayName}
                description={planStatus === "trialing" ? t("plan.trialSubtitle") : planIntervalLabel}
                tone={planVisual.tone}
                className="card-hover"
                valueClassName="break-words"
              />
              <MetricCard
                icon={Clock3}
                label={t("plan.daysRemaining")}
                value={daysRemaining === null ? common("notSet") : t("plan.daysValue", { days: numberFormatter.format(daysRemaining) })}
                description={planDeadlineISO ? formatDateTime(planDeadlineISO) : common("notSet")}
                tone={daysRemaining !== null && daysRemaining <= 10 ? "warning" : "success"}
                className="card-hover"
              />
              <UsageSummaryCard
                icon={Users}
                title={t("plan.userLimit")}
                used={usersTotal}
                limit={usersLimit}
                subtitle={t("plan.userLimitHint")}
                tone={usersUsageRatio !== null && usersUsageRatio >= 1 ? "warning" : "success"}
              />
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
      value={formatUsageValue(props.used, props.limit)}
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

function formatDateTime(value: string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return value;
  return new Intl.DateTimeFormat("fa-IR", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(timestamp));
}

async function loadPlanRequestUsage(
  subscription: TenantSubscription | null,
  fallback: number | null,
): Promise<number | null> {
  const planPeriodStart = getPlanPeriodStart(subscription);
  if (!planPeriodStart) return null;

  try {
    const response = await requestJSON<MetricsQueryResponse>("/api/metrics/query", {
      method: "POST",
      body: JSON.stringify({
        startTs: planPeriodStart,
        endTs: new Date().toISOString(),
        datasource: "modelMetrics",
        type: "distribution",
      }),
    });

    return extractMetricsCount(response) ?? fallback;
  } catch {
    return fallback;
  }
}

function getPlanPeriodStart(subscription: TenantSubscription | null): string {
  return subscription?.current_period_start?.trim() || subscription?.trial_started_at?.trim() || "";
}

function extractMetricsCount(response: MetricsQueryResponse): number | null {
  const row = response.data?.[0] ?? null;
  if (!row || typeof row !== "object") return null;

  const rawCount = row.countAll;
  if (typeof rawCount === "number" && Number.isFinite(rawCount)) return rawCount;
  if (typeof rawCount === "string") {
    const parsed = Number(rawCount);
    return Number.isFinite(parsed) ? parsed : null;
  }

  return null;
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

function formatUsageValue(used: number | null, limit: number | null): string {
  return `${formatUsagePart(used)} / ${formatUsagePart(limit)}`;
}

function clampUsageRatio(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.max(0, Math.min(1, value));
}

function getPlanVisual(status: string): {
  icon: LucideIcon;
  tone: "success" | "warning" | "muted";
  badgeVariant: "success" | "warning" | "outline" | "destructive";
  statusKey: "active" | "trialing" | "past_due" | "inactive";
} {
  switch (status) {
    case "trialing":
      return { icon: FlaskConical, tone: "warning", badgeVariant: "warning", statusKey: "trialing" };
    case "active":
      return { icon: Crown, tone: "success", badgeVariant: "success", statusKey: "active" };
    case "past_due":
      return { icon: AlertTriangle, tone: "warning", badgeVariant: "destructive", statusKey: "past_due" };
    default:
      return { icon: ShieldCheck, tone: "muted", badgeVariant: "outline", statusKey: "inactive" };
  }
}
