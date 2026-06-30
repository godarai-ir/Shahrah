"use client";

import { useState, useTransition } from "react";
import { useLocale, useTranslations } from "next-intl";
import { ArrowUpRight, Check, Loader2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type PlanPrice = {
  billing_interval: string;
  currency: string;
  amount_minor?: number | null;
  discount_percent?: number;
};

type CommercialPlan = {
  slug: string;
  display_name: string;
  description?: string;
  prices?: PlanPrice[];
  features?: Record<string, unknown>;
};

type TenantSubscription = {
  plan_slug?: string;
  plan_display_name?: string;
  status: string;
  billing_interval: string;
  current_period_end?: string;
	trial_ends_at?: string;
};

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
  overrides?: Record<string, unknown>;
  usage_counters?: UsageCounter[];
};

type UsageCounter = {
  feature_key: string;
  used: number;
  limit?: number | null;
  window_end?: string;
};

type PlanManagementProps = {
  plans: CommercialPlan[];
  subscription: TenantSubscription | null;
  entitlements: EffectiveEntitlements | null;
  initialError?: string | null;
};

export default function PlanManagement({
  plans,
  subscription,
  entitlements,
  initialError,
}: PlanManagementProps) {
  const t = useTranslations("billingPlan");
  const common = useTranslations("common");
  const locale = useLocale();
  const [pending, startTransition] = useTransition();
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(initialError ?? null);
  const [target, setTarget] = useState("");

  const currentSlug = subscription?.plan_slug ?? "";
  const features = entitlements?.features ?? {};
  const overrides = entitlements?.overrides ?? {};
  const requestUsage = findUsageCounter(entitlements, "gateway.requests.monthly");

  function changePlan(planSlug: string, billingInterval: string) {
    const nextTarget = `${planSlug}:${billingInterval}`;
    setTarget(nextTarget);
    setError(null);
    setMessage(null);
    startTransition(async () => {
      const response = await fetch("/api/subscription/change", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ plan_slug: planSlug, billing_interval: billingInterval }),
      });
      const payload = (await response.json().catch(() => null)) as {
        error?: { message?: string };
        mode?: string;
        payment_intent?: { payment_url?: string };
      } | null;
      if (!response.ok) {
        setError(payload?.error?.message ?? t("messages.changeFailed"));
        return;
      }
      if (payload?.mode === "payment_required" && payload.payment_intent?.payment_url) {
        window.location.assign(payload.payment_intent.payment_url);
        return;
      }
      setMessage(t("messages.changeSaved"));
      window.location.reload();
    });
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="text-sm font-medium text-muted-foreground">{t("eyebrow")}</p>
          <h1 className="text-2xl font-semibold tracking-normal text-foreground">{t("title")}</h1>
        </div>
        {subscription ? (
          <Badge variant={subscription.status === "active" ? "success" : "warning"}>
            {t(`status.${normalizeStatus(subscription.status)}`)}
          </Badge>
        ) : null}
      </div>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("messages.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {message ? (
        <Alert variant="success">
          <AlertTitle>{t("messages.successTitle")}</AlertTitle>
          <AlertDescription>{message}</AlertDescription>
        </Alert>
      ) : null}
      {subscription?.status === "trialing" ? (
        <Alert>
          <AlertTitle>{t("trial.title")}</AlertTitle>
          <AlertDescription>
            {t("trial.description", {
              plan: subscription.plan_display_name ?? subscription.plan_slug ?? common("notSet"),
              date: formatDate(subscription.trial_ends_at, locale),
            })}
          </AlertDescription>
        </Alert>
      ) : null}

      <section className="grid gap-4 lg:grid-cols-2">
        <Card className="rounded-lg">
          <CardHeader>
            <CardTitle>{t("currentSubscription.title")}</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Metric label={t("currentSubscription.plan")} value={subscription?.plan_display_name ?? subscription?.plan_slug ?? common("notSet")} />
            <Metric label={t("currentSubscription.interval")} value={intervalLabel(subscription?.billing_interval, t)} />
            <Metric label={t("currentSubscription.periodEnd")} value={formatDate(subscription?.current_period_end, locale, common("notSet"))} />
            {subscription?.status === "trialing" ? (
              <Metric label={t("currentSubscription.trialEnds")} value={formatDate(subscription.trial_ends_at, locale, common("notSet"))} />
            ) : null}
          </CardContent>
        </Card>
        <Card className="rounded-lg">
          <CardHeader>
            <CardTitle>{t("limits.title")}</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Metric label={t("limits.requests")} value={formatUsage(requestUsage, features["gateway.requests.monthly"], locale, t, common)} />
            <Metric label={t("limits.users")} value={formatLimit(features["users.max"], locale, common, t)} />
            <Metric label={t("limits.vat")} value={formatEnabled(features["auth.vat.enabled"], t)} />
            <Metric label={t("limits.overrides")} value={Object.keys(overrides).length ? formatCount(Object.keys(overrides).length, locale) : t("values.none")} />
          </CardContent>
        </Card>
      </section>

      <section className="grid gap-4 lg:grid-cols-2 xl:grid-cols-4">
        {plans.map((plan) => {
          const isCurrent = plan.slug === currentSlug;
          const prices = orderedPrices(plan.prices ?? []);
          return (
            <Card key={plan.slug} className="rounded-lg">
              <CardHeader className="gap-3">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <CardTitle>{plan.display_name}</CardTitle>
                    <p className="mt-2 text-sm text-muted-foreground">{plan.description ?? plan.slug}</p>
                  </div>
                  {isCurrent ? <Badge variant="secondary">{t("badges.current")}</Badge> : null}
                </div>
              </CardHeader>
              <CardContent className="flex flex-col gap-4">
                <div className="flex flex-col gap-2 text-sm">
                  <Feature label={t("features.requests")} value={formatLimit(plan.features?.["gateway.requests.monthly"], locale, common, t)} />
                  <Feature label={t("features.users")} value={formatLimit(plan.features?.["users.max"], locale, common, t)} />
                  <Feature label={t("features.vat")} value={formatEnabled(plan.features?.["auth.vat.enabled"], t)} />
                  <Feature label={t("features.ratePolicies")} value={formatEnabled(plan.features?.["rate_limit_rules.enabled"], t)} />
                  <Feature label={t("features.budgetPolicies")} value={formatEnabled(plan.features?.["budget_limit_rules.enabled"], t)} />
                </div>
                <div className="grid gap-2">
                  {prices.map((price) => {
                    const key = `${plan.slug}:${price.billing_interval}`;
                    const activeInterval = isCurrent && price.billing_interval === subscription?.billing_interval;
                    return (
                      <Button
                        key={price.billing_interval}
                        type="button"
                        variant={activeInterval ? "secondary" : "outline"}
                        className="justify-between rounded-lg"
                        disabled={pending || activeInterval}
                        onClick={() => changePlan(plan.slug, price.billing_interval)}
                      >
                        <span>{intervalLabel(price.billing_interval, t)}</span>
                        <span className="inline-flex items-center gap-2 text-xs">
                          {formatPrice(price, locale, t)}
                          {pending && target === key ? (
                            <Loader2 className="size-3 animate-spin" />
                          ) : activeInterval ? (
                            <Check className="size-3" />
                          ) : (
                            <ArrowUpRight className="size-3" />
                          )}
                        </span>
                      </Button>
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </section>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border border-border px-3 py-2">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}

function Feature({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}

function orderedPrices(prices: PlanPrice[]) {
  const order = new Map([
    ["monthly", 0],
    ["quarterly", 1],
    ["yearly", 2],
  ]);
  return [...prices].sort((a, b) => (order.get(a.billing_interval) ?? 99) - (order.get(b.billing_interval) ?? 99));
}

function formatLimit(
  value: unknown,
  locale: string,
  common: ReturnType<typeof useTranslations>,
  t: ReturnType<typeof useTranslations>,
): string {
  if (value === null || value === undefined || value === "custom") {
    return t("values.custom");
  }
  if (typeof value === "number") {
    return formatCount(value, locale);
  }
  if (typeof value === "string" && value.trim()) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? formatCount(parsed, locale) : value;
  }
  return common("notSet");
}

function findUsageCounter(entitlements: EffectiveEntitlements | null, featureKey: string): UsageCounter | null {
  return entitlements?.usage_counters?.find((item) => item.feature_key === featureKey) ?? null;
}

function formatUsage(
  counter: UsageCounter | null,
  fallbackLimit: unknown,
  locale: string,
  t: ReturnType<typeof useTranslations>,
  common: ReturnType<typeof useTranslations>,
): string {
  const used = counter ? formatCount(counter.used, locale) : formatCount(0, locale);
  const limit = counter?.limit ?? fallbackLimit;
  return t("values.usage", { used, limit: formatLimit(limit, locale, common, t) });
}

function formatEnabled(value: unknown, t: ReturnType<typeof useTranslations>): string {
  return value === true ? t("values.enabled") : t("values.disabled");
}

function formatDate(value: string | undefined, locale: string, fallback?: string): string {
  if (!value) {
    return fallback ?? "—";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(locale, { dateStyle: "medium" }).format(date);
}

function intervalLabel(interval: string | undefined, t: ReturnType<typeof useTranslations>): string {
  switch (interval) {
    case "monthly":
      return t("intervals.monthly");
    case "quarterly":
      return t("intervals.quarterly");
    case "yearly":
      return t("intervals.yearly");
    default:
      return interval || t("values.unknown");
  }
}

function formatPrice(price: PlanPrice, locale: string, t: ReturnType<typeof useTranslations>): string {
  if (price.amount_minor === null || price.amount_minor === undefined) {
    return t("values.custom");
  }
  const formatted = formatCount(price.amount_minor, locale);
  if (price.discount_percent && price.discount_percent > 0) {
    return t("values.priceWithDiscount", {
      amount: formatted,
      currency: price.currency,
      discount: formatCount(price.discount_percent, locale),
    });
  }
  return t("values.price", { amount: formatted, currency: price.currency });
}

function normalizeStatus(status: string): "active" | "trialing" | "past_due" | "other" {
  switch (status.trim().toLowerCase()) {
    case "active":
      return "active";
    case "trialing":
      return "trialing";
    case "past_due":
      return "past_due";
    default:
      return "other";
  }
}

function formatCount(value: number, locale: string): string {
  return new Intl.NumberFormat(locale).format(value);
}
