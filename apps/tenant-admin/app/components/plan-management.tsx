"use client";

import { useState, useTransition } from "react";
import { ArrowUpRight, Check, Loader2 } from "lucide-react";

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
        setError(payload?.error?.message ?? "Plan change failed.");
        return;
      }
      if (payload?.mode === "payment_required" && payload.payment_intent?.payment_url) {
        window.location.assign(payload.payment_intent.payment_url);
        return;
      }
      setMessage("Plan change saved.");
      window.location.reload();
    });
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="text-sm font-medium text-muted-foreground">Billing</p>
          <h1 className="text-2xl font-semibold tracking-normal text-foreground">Plan</h1>
        </div>
        {subscription ? (
          <Badge variant={subscription.status === "active" ? "success" : "warning"}>
            {subscription.status}
          </Badge>
        ) : null}
      </div>

      {error ? (
        <div className="rounded-lg border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      ) : null}
      {message ? (
        <div className="rounded-lg border border-success/30 bg-success/10 px-4 py-3 text-sm text-success">
          {message}
        </div>
      ) : null}

      <section className="grid gap-4 lg:grid-cols-2">
        <Card className="rounded-lg">
          <CardHeader>
            <CardTitle>Current subscription</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Metric label="Plan" value={subscription?.plan_display_name ?? subscription?.plan_slug ?? "Unavailable"} />
            <Metric label="Interval" value={subscription?.billing_interval ?? "Unavailable"} />
            <Metric label="Period end" value={formatDate(subscription?.current_period_end)} />
          </CardContent>
        </Card>
        <Card className="rounded-lg">
          <CardHeader>
            <CardTitle>Effective limits</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Metric label="Monthly gateway requests" value={formatUsage(requestUsage, features["gateway.requests.monthly"])} />
            <Metric label="Users" value={formatLimit(features["users.max"])} />
            <Metric label="VAT" value={formatEnabled(features["auth.vat.enabled"])} />
            <Metric label="Overrides" value={Object.keys(overrides).length ? String(Object.keys(overrides).length) : "None"} />
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
                  {isCurrent ? <Badge variant="secondary">Current</Badge> : null}
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2 text-sm">
                  <Feature label="Requests" value={formatLimit(plan.features?.["gateway.requests.monthly"])} />
                  <Feature label="Users" value={formatLimit(plan.features?.["users.max"])} />
                  <Feature label="VAT" value={formatEnabled(plan.features?.["auth.vat.enabled"])} />
                  <Feature label="Rate policies" value={formatEnabled(plan.features?.["rate_limit_rules.enabled"])} />
                  <Feature label="Budget policies" value={formatEnabled(plan.features?.["budget_limit_rules.enabled"])} />
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
                        loading={pending && target === key}
                        disabled={pending || activeInterval}
                        onClick={() => changePlan(plan.slug, price.billing_interval)}
                      >
                        <span>{intervalLabel(price.billing_interval)}</span>
                        <span className="inline-flex items-center gap-2 text-xs">
                          {formatPrice(price)}
                          {pending && target === key ? (
                            <Loader2 className="h-3 w-3 animate-spin" />
                          ) : activeInterval ? (
                            <Check className="h-3 w-3" />
                          ) : (
                            <ArrowUpRight className="h-3 w-3" />
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

function formatLimit(value: unknown): string {
  if (value === null || value === undefined || value === "custom") {
    return "Custom";
  }
  if (typeof value === "number") {
    return new Intl.NumberFormat("en-US").format(value);
  }
  if (typeof value === "string" && value.trim()) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? new Intl.NumberFormat("en-US").format(parsed) : value;
  }
  return "Unavailable";
}

function findUsageCounter(entitlements: EffectiveEntitlements | null, featureKey: string): UsageCounter | null {
  return entitlements?.usage_counters?.find((item) => item.feature_key === featureKey) ?? null;
}

function formatUsage(counter: UsageCounter | null, fallbackLimit: unknown): string {
  const used = counter ? new Intl.NumberFormat("en-US").format(counter.used) : "0";
  const limit = counter?.limit ?? fallbackLimit;
  return `${used} / ${formatLimit(limit)}`;
}

function formatEnabled(value: unknown): string {
  return value === true ? "Enabled" : "Disabled";
}

function formatDate(value?: string): string {
  if (!value) {
    return "Unavailable";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-US", { dateStyle: "medium" }).format(date);
}

function intervalLabel(interval: string): string {
  switch (interval) {
    case "monthly":
      return "Monthly";
    case "quarterly":
      return "Quarterly";
    case "yearly":
      return "Yearly";
    default:
      return interval;
  }
}

function formatPrice(price: PlanPrice): string {
  if (price.amount_minor === null || price.amount_minor === undefined) {
    return "Custom";
  }
  const formatted = new Intl.NumberFormat("en-US").format(price.amount_minor);
  const discount = price.discount_percent && price.discount_percent > 0 ? ` -${price.discount_percent}%` : "";
  return `${formatted} ${price.currency}${discount}`;
}
