"use client";

import { BookOpen } from "lucide-react";
import type { ReactNode } from "react";
import { useLocale, useTranslations } from "next-intl";

import { EntityIdentityHeader } from "@/components/app-ui/EntityIdentityHeader";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
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
};

type EntitlementRow = {
  key: string;
  title: string;
  value: unknown;
  overridden: boolean;
  docsHref?: string | null;
};

type PlanManagementProps = {
  subscription: TenantSubscription | null;
  entitlements: EffectiveEntitlements | null;
  docsBaseHref: string;
  initialError?: string | null;
};

const featureOrder = [
  "credits.monthly_grant",
  "credits.trial_grant",
  "credits.trial_days",
  "credits.overage_allowed",
  "credits.rollover_allowed",
  "credits.bonus_percent",
  "credits.byok_fee_percent",
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

const docsPathByFeature: Record<string, string> = {
  "users.max": "/guides/users-access-overview",
  "teams.enabled": "/guides/team-management",
  "cache.simple.enabled": "/guides/caching",
  "cache.simple.ttl_days": "/guides/caching",
  "cache.semantic.enabled": "/guides/caching",
  "routing.weighted.enabled": "/virtual-model",
  "routing.latency.enabled": "/virtual-model",
  "rate_limit_rules.enabled": "/guides/rate-limiting",
  "budget_limit_rules.enabled": "/guides/budget-limiting",
  "credits.monthly_grant": "/financial/rate-card",
  "credits.byok_fee_percent": "/financial/rate-card",
  "logs.retention_days": "/observability-logs",
  "traces.retention_days": "/observability-traces",
  "metadata.custom.enabled": "/observability-logs",
  "llm_pricing.custom.enabled": "/financial/overview",
  "analytics.enabled": "/observability-model-metrics",
  "alerts.enabled": "/observability-overview",
  "otel_export.enabled": "/observability-opentelemetry",
  "prompt_templates.max": "/prompt-registry",
  "prompt_versions.enabled": "/prompt-registry",
  "guardrails.third_party.enabled": "/guardrails",
  "guardrails.webhook.enabled": "/guardrails",
  "auth.pat.enabled": "/guides/api-keys",
  "auth.vat.enabled": "/guides/virtual-account-management",
  "rbac.enabled": "/guides/roles-and-permissions",
  "org_units.enabled": "/guides/org-unit-management",
};

export default function PlanManagement({
  subscription,
  entitlements,
  docsBaseHref,
  initialError,
}: PlanManagementProps) {
  const t = useTranslations("billingPlan");
  const common = useTranslations("common");
  const locale = useLocale();

  const currentSlug = subscription?.plan_slug ?? "";
  const features = entitlements?.features ?? {};
  const overrides = entitlements?.overrides ?? {};
  const currentPlanTitle = subscription?.plan_display_name ?? currentSlug;
  const rows = buildEntitlementRows(features, overrides, t, docsBaseHref);

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
                          <FeatureTitle row={row} t={t} />
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
  t: ReturnType<typeof useTranslations>,
  docsBaseHref: string,
): EntitlementRow[] {
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
    overridden: Object.prototype.hasOwnProperty.call(overrides, key),
    docsHref: getFeatureDocsHref(key, docsBaseHref),
  }));
}

function FeatureTitle({
  row,
  t,
}: {
  row: EntitlementRow;
  t: ReturnType<typeof useTranslations>;
}) {
  const docsLabel = row.docsHref ? t("table.openDocs", { feature: row.title }) : "";

  return (
    <span className="inline-flex max-w-full items-center gap-1.5 align-middle">
      <span className="min-w-0">{row.title}</span>
      {row.docsHref ? (
        <a
          href={row.docsHref}
          target="_blank"
          rel="noopener noreferrer"
          aria-label={docsLabel}
          title={docsLabel}
          className="inline-flex shrink-0 rounded-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        >
          <BookOpen className="size-3.5" aria-hidden="true" />
        </a>
      ) : null}
    </span>
  );
}

function getFeatureDocsHref(key: string, docsBaseHref: string): string | null {
  const docsPath = docsPathByFeature[key];

  return docsPath ? `${docsBaseHref.replace(/\/$/, "")}${docsPath}` : null;
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
  if (row.key === "credits.monthly_grant" || row.key === "credits.trial_grant") {
    return t("values.credits", { count: formatLimit(row.value, locale, common, t) });
  }
  if (row.key === "credits.bonus_percent" || row.key === "credits.byok_fee_percent") {
    return t("values.percent", { value: formatLimit(row.value, locale, common, t) });
  }
  if (row.key.endsWith(".ttl_days") || row.key.endsWith(".retention_days") || row.key === "credits.trial_days") {
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

const creditFeatureTitleKeys: Record<string, string> = {
  "credits.monthly_grant": "features.creditsMonthlyGrant",
  "credits.trial_grant": "features.creditsTrialGrant",
  "credits.trial_days": "features.creditsTrialDays",
  "credits.overage_allowed": "features.creditsOverageAllowed",
  "credits.rollover_allowed": "features.creditsRolloverAllowed",
  "credits.bonus_percent": "features.creditsBonusPercent",
  "credits.byok_fee_percent": "features.creditsByokFeePercent",
};

function getFeatureTitle(key: string, t: ReturnType<typeof useTranslations>): string {
  if (key in creditFeatureTitleKeys) return t(creditFeatureTitleKeys[key]);
  switch (key) {
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
