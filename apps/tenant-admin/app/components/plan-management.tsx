"use client";

import type { ReactNode } from "react";
import { useLocale, useTranslations } from "next-intl";

import { EntityIdentityHeader, StatusBadge } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type TenantSubscription = {
  plan_slug?: string;
  plan_display_name?: string;
  status: string;
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

type EntitlementRow = {
  key: string;
  title: string;
  value: unknown;
  usageCounter?: UsageCounter | null;
  overridden: boolean;
};

type PlanManagementProps = {
  subscription: TenantSubscription | null;
  entitlements: EffectiveEntitlements | null;
  initialError?: string | null;
};

const featureOrder = [
  "gateway.requests.monthly",
  "users.max",
  "teams.enabled",
  "cache.simple.enabled",
  "cache.simple.ttl_days",
  "cache.semantic.enabled",
  "routing.weighted.enabled",
  "routing.latency.enabled",
  "rate_limit_rules.enabled",
  "budget_limit_rules.enabled",
  "logs.retention_days",
  "traces.retention_days",
  "metadata.custom.enabled",
  "llm_pricing.custom.enabled",
  "analytics.enabled",
  "alerts.enabled",
  "otel_export.enabled",
  "prompt_templates.max",
  "prompt_versions.enabled",
  "guardrails.third_party.enabled",
  "guardrails.webhook.enabled",
  "auth.pat.enabled",
  "auth.vat.enabled",
  "rbac.enabled",
  "sso.enabled",
  "org_units.enabled",
  "admin_audit.enabled",
  "deployment.saas.enabled",
  "deployment.vpc.enabled",
  "support.tier",
  "support.sla",
] as const;

export default function PlanManagement({
  subscription,
  entitlements,
  initialError,
}: PlanManagementProps) {
  const t = useTranslations("billingPlan");
  const common = useTranslations("common");
  const locale = useLocale();

  const currentSlug = subscription?.plan_slug ?? "";
  const features = entitlements?.features ?? {};
  const overrides = entitlements?.overrides ?? {};
  const currentPlanTitle = subscription?.plan_display_name ?? currentSlug;
  const rows = buildEntitlementRows(features, overrides, entitlements?.usage_counters ?? [], t);

  return (
    <div className="flex flex-col gap-6">
      {initialError ? (
        <Alert variant="destructive">
          <AlertTitle>{t("messages.errorTitle")}</AlertTitle>
          <AlertDescription>{initialError}</AlertDescription>
        </Alert>
      ) : null}

      <section>
        <Card className="rounded-lg">
          <CardHeader>
            <EntityIdentityHeader
              title={currentPlanTitle || common("notSet")}
              identifier={currentSlug || undefined}
              description={t("currentPlan.description")}
              status={
                subscription ? (
                  <>
                    <Badge variant="secondary">{t("badges.current")}</Badge>
                    <Badge variant={subscription.status === "active" ? "success" : "warning"}>
                      {t(`status.${normalizeStatus(subscription.status)}`)}
                    </Badge>
                  </>
                ) : null
              }
            />
          </CardHeader>
          <CardContent>
            <Table>
              <TableCaption className="sr-only">{t("table.caption")}</TableCaption>
              <TableHeader>
                <TableRow className="bg-muted/40 hover:bg-muted/40">
                  <TableHead scope="col">{t("table.feature")}</TableHead>
                  <TableHead scope="col">{t("table.value")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.length ? (
                  rows.map((row) => (
                    <TableRow key={row.key}>
                      <TableCell className="min-w-52 font-medium">
                        <div className="flex min-w-0 flex-col gap-1">
                          <span>{row.title}</span>
                          {row.overridden ? (
                            <span className="text-xs font-normal text-muted-foreground">{t("badges.override")}</span>
                          ) : null}
                        </div>
                      </TableCell>
                      <TableCell>{formatEntitlementValue(row, locale, t, common)}</TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={2} className="py-8 text-center text-muted-foreground">
                      {t("table.empty")}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </section>
    </div>
  );
}

function formatLimit(value: unknown, locale: string, common: ReturnType<typeof useTranslations>, t: ReturnType<typeof useTranslations>): string {
  if (value === null || value === undefined || value === "custom") {
    return t("values.custom");
  }
  if (value === -1 || value === "-1") {
    return t("values.unlimited");
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

function buildEntitlementRows(
  features: Record<string, unknown>,
  overrides: Record<string, unknown>,
  usageCounters: UsageCounter[],
  t: ReturnType<typeof useTranslations>,
): EntitlementRow[] {
  const usageByKey = new Map(usageCounters.map((counter) => [counter.feature_key, counter]));
  const orderedKeys = [
    ...featureOrder.filter((key) => Object.prototype.hasOwnProperty.call(features, key)),
    ...Object.keys(features)
      .filter((key) => !(featureOrder as readonly string[]).includes(key))
      .sort((a, b) => a.localeCompare(b)),
  ];

  return orderedKeys.map((key) => ({
    key,
    title: getFeatureTitle(key, t),
    value: features[key],
    usageCounter: usageByKey.get(key),
    overridden: Object.prototype.hasOwnProperty.call(overrides, key),
  }));
}

function formatEntitlementValue(
  row: EntitlementRow,
  locale: string,
  t: ReturnType<typeof useTranslations>,
  common: ReturnType<typeof useTranslations>,
): ReactNode {
  if (typeof row.value === "boolean") {
    return (
      <StatusBadge status={row.value ? "active" : "inactive"}>
        {row.value ? t("values.enabled") : t("values.disabled")}
      </StatusBadge>
    );
  }
  if (row.key === "support.tier" || row.key === "support.sla") {
    return formatSupportValue(row.key, row.value, t, common);
  }
  if (row.usageCounter) {
    const used = formatCount(row.usageCounter.used, locale);
    const limit = row.usageCounter.limit ?? row.value;
    return t("values.usage", { used, limit: formatLimit(limit, locale, common, t) });
  }
  if (row.key.endsWith(".ttl_days") || row.key.endsWith(".retention_days")) {
    return formatDayLimit(row.value, locale, common, t);
  }
  return formatLimit(row.value, locale, common, t);
}

function formatDayLimit(
  value: unknown,
  locale: string,
  common: ReturnType<typeof useTranslations>,
  t: ReturnType<typeof useTranslations>,
): string {
  if (value === null || value === undefined || value === "custom") {
    return t("values.custom");
  }
  if (value === -1 || value === "-1") {
    return t("values.unlimited");
  }
  return t("values.days", { count: formatLimit(value, locale, common, t) });
}

function formatSupportValue(
  key: string,
  value: unknown,
  t: ReturnType<typeof useTranslations>,
  common: ReturnType<typeof useTranslations>,
): string {
  if (value === null || value === undefined || value === "") {
    return key === "support.sla" ? t("values.noSla") : common("notSet");
  }
  if (typeof value !== "string") {
    return String(value);
  }
  switch (value) {
    case "chatbot":
      return t("support.chatbot");
    case "product":
      return t("support.product");
    case "priority":
      return t("support.priority");
    case "standard":
      return t("support.standardSla");
    case "enterprise":
      return t("support.enterpriseSla");
    default:
      return value;
  }
}

function getFeatureTitle(key: string, t: ReturnType<typeof useTranslations>): string {
  switch (key) {
    case "gateway.requests.monthly":
      return t("features.gatewayRequestsMonthly");
    case "users.max":
      return t("features.usersMax");
    case "teams.enabled":
      return t("features.teamsEnabled");
    case "cache.simple.enabled":
      return t("features.cacheSimpleEnabled");
    case "cache.simple.ttl_days":
      return t("features.cacheSimpleTtlDays");
    case "cache.semantic.enabled":
      return t("features.cacheSemanticEnabled");
    case "routing.weighted.enabled":
      return t("features.routingWeightedEnabled");
    case "routing.latency.enabled":
      return t("features.routingLatencyEnabled");
    case "rate_limit_rules.enabled":
      return t("features.rateLimitRulesEnabled");
    case "budget_limit_rules.enabled":
      return t("features.budgetLimitRulesEnabled");
    case "logs.retention_days":
      return t("features.logsRetentionDays");
    case "traces.retention_days":
      return t("features.tracesRetentionDays");
    case "metadata.custom.enabled":
      return t("features.metadataCustomEnabled");
    case "llm_pricing.custom.enabled":
      return t("features.llmPricingCustomEnabled");
    case "analytics.enabled":
      return t("features.analyticsEnabled");
    case "alerts.enabled":
      return t("features.alertsEnabled");
    case "otel_export.enabled":
      return t("features.otelExportEnabled");
    case "prompt_templates.max":
      return t("features.promptTemplatesMax");
    case "prompt_versions.enabled":
      return t("features.promptVersionsEnabled");
    case "guardrails.third_party.enabled":
      return t("features.guardrailsThirdPartyEnabled");
    case "guardrails.webhook.enabled":
      return t("features.guardrailsWebhookEnabled");
    case "auth.pat.enabled":
      return t("features.authPatEnabled");
    case "auth.vat.enabled":
      return t("features.authVatEnabled");
    case "rbac.enabled":
      return t("features.rbacEnabled");
    case "sso.enabled":
      return t("features.ssoEnabled");
    case "org_units.enabled":
      return t("features.orgUnitsEnabled");
    case "admin_audit.enabled":
      return t("features.adminAuditEnabled");
    case "deployment.saas.enabled":
      return t("features.deploymentSaasEnabled");
    case "deployment.vpc.enabled":
      return t("features.deploymentVpcEnabled");
    case "support.tier":
      return t("features.supportTier");
    case "support.sla":
      return t("features.supportSla");
    default:
      return key;
  }
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
