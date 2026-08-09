import Link from "next/link";
import type { ReactNode } from "react";
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  BarChart3,
  Clock,
  Coins,
  Database,
  FileJson,
  Gauge,
  Hash,
  Layers,
  RefreshCw,
  Route,
  Server,
  Timer,
  type LucideIcon,
} from "lucide-react";
import { getLocale, getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { DefinitionList, DefinitionListItem as CompactField } from "@/components/app-ui/DefinitionList";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { SectionHeader } from "@/components/app-ui/SectionHeader";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { formatCompactCurrency, formatCompactIrr, formatCompactNumber } from "@/lib/compact-number";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { cn } from "@/lib/utils";
import {
  parseTraceAttempts,
  parseTraceDetail,
  parseTraceSpans,
  type TraceDetailV2Response,
  type TraceSpan,
  type RequestTraceV2,
} from "@/features/traces/detail-contracts";

type Params = { params: Promise<{ traceID: string }> };
type MetadataSource = "span" | "usage";
type MetadataRow = { source: MetadataSource; key: string; value: string };
type BadgeVariant = "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "info";

export const dynamic = "force-dynamic";

export default async function TraceDetailPage({ params }: Params) {
  const t = await getTranslations("traces");
  const locale = await getLocale();
  const auth = await requireAuth("/login?next=/traces");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);
  const { traceID } = await params;
  const normalized = traceID.trim();

  const [traceResult, timelineResult, attemptsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/traces/${encodeURIComponent(normalized)}`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/traces/${encodeURIComponent(normalized)}/timeline`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/traces/${encodeURIComponent(normalized)}/attempts`,
      userID: auth.user.id,
    }),
  ]);

  const detail = traceResult.ok ? parseTraceDetail(traceResult.payload) : null;
  const spans = timelineResult.ok ? parseTraceSpans(timelineResult.payload) : [];
  const attempts = attemptsResult.ok ? parseTraceAttempts(attemptsResult.payload) : [];
  const sortedSpans = sortSpans(spans);
  const waterfallRows = buildWaterfallRows(sortedSpans);
  const hasNonZeroSpanOffset = waterfallRows.some((row) => row.offsetMS > 0);
  const timelineTotalMS = waterfallRows[0]?.totalMS ?? 0;
  const rootSpan = sortedSpans.find((span) => span.name === "gateway.request" && span.kind === "server");
  const providerSpan = sortedSpans.find((span) => span.name === "provider.call");

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  const errorRows = buildErrorRows(detail, sortedSpans, t, locale);
  const emptyLabel = t("detail.values.empty");
  const msLabel = t("detail.units.ms");
  const metadataRows = buildMetadataRows(rootSpan?.attributes, detail?.usage?.metadata).filter((row) => shouldShowMetadataRow(row, detail));
  const requestID = distinctValue(detail?.trace.request_id, detail?.trace.trace_id);
  const providerModel = distinctValue(detail?.metric?.provider_model_name, detail?.trace.model);
  const metricErrorType = detail?.metric?.error_type;
  const usageLatency = detail?.metric?.latency_ms ?? detail?.usage?.response_time_ms;
  const showUsageLatency = typeof usageLatency === "number" && !isNearNumber(usageLatency, providerSpan?.duration_ms, 25);
  const showAttemptRoutingFields = attempts.length > 1 || attempts.some((attempt) => hasText(attempt.retry_reason) || hasText(attempt.fallback_reason));
  const showAttemptFallbackField = attempts.some((attempt) => hasText(attempt.fallback_reason));
  const metadataSourceLabels: Record<MetadataSource, string> = {
    span: t("detail.metadata.sources.span"),
    usage: t("detail.metadata.sources.usage"),
  };

  return (
    <AdminShell messageNamespaces={["traces","settings"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("detail.title"), description: normalized || emptyLabel }}
    >
      <PageContainer spacing="compact" padding="none" className="md:gap-6">
        <PageHeader
          eyebrow={t("detail.kicker")}
          title={t("detail.title")}
          description={<span className="block break-all font-mono text-xs">{normalized || emptyLabel}</span>}
          actions={
            <>
              {detail ? (
                <Button asChild variant="outline" size="sm">
                  <Link href={buildMetricsLink(detail.trace)}>
                    <BarChart3 data-icon="inline-start" />
                    {t("detail.actions.openInMetrics")}
                  </Link>
                </Button>
              ) : null}
              <Button asChild variant="ghost" size="sm">
                <Link href="/traces">
                  <ArrowRight data-icon="inline-start" />
                  {t("detail.actions.backToTraces")}
                </Link>
              </Button>
            </>
          }
        />

        {!detail ? (
          <Card>
            <CardHeader className="gap-1 px-4 py-3 md:px-4 md:py-3">
              <CardTitle className="text-base tracking-normal">{t("detail.notFound.title")}</CardTitle>
              <CardDescription>{traceResult.errorMessage ?? t("detail.notFound.description")}</CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <>
            <div className="grid gap-4 xl:grid-cols-2">
              <Card>
                <CardHeader className="px-4 py-3 md:px-4 md:py-3"><SectionHeader density="compact"
                  icon={Activity}
                  title={t("detail.summary.title")}
                  description={t("detail.summary.description")}
                /></CardHeader>
                <CardContent className="flex flex-col gap-4 px-4 pb-4 text-sm md:px-4 md:pb-4">
                  <div className="grid grid-cols-2 gap-2 lg:grid-cols-3 2xl:grid-cols-5">
                    <MetricItem
                      icon={Gauge}
                      label={t("detail.summary.status")}
                      value={
                        <Badge variant={statusCodeBadgeVariant(detail.trace.status_code)}>
                          {formatNumber(detail.trace.status_code, locale)}
                        </Badge>
                      }
                    />
                    <MetricItem
                      icon={Clock}
                      label={t("detail.summary.duration")}
                      value={formatDuration(detail.trace.duration_ms, msLabel, locale)}
                    />
                    <MetricItem
                      icon={RefreshCw}
                      label={t("detail.summary.attempts")}
                      value={<CompactValue display={formatCompactNumber(detail.summary.attempts_count, { locale })} />}
                    />
                    <MetricItem
                      icon={Layers}
                      label={t("detail.summary.spans")}
                      value={<CompactValue display={formatCompactNumber(detail.summary.spans_count, { locale })} />}
                    />
                    <MetricItem
                      icon={AlertTriangle}
                      label={t("detail.summary.errors")}
                      value={<CompactValue display={formatCompactNumber(detail.summary.errors_count, { locale })} />}
                      valueClassName={detail.summary.errors_count > 0 ? "text-destructive" : undefined}
                    />
                  </div>
                  <DefinitionList columns={3}>
                    {requestID ? (
                      <CompactField
                        icon={Hash}
                        label={t("detail.summary.requestId")}
                        value={requestID}
                        mono
                        wrap
                      />
                    ) : null}
                    <CompactField icon={Route} label={t("detail.summary.route")} value={detail.trace.route} mono wrap />
                    <CompactField icon={Server} label={t("detail.summary.provider")} value={displayValue(detail.trace.provider, emptyLabel)} />
                    <CompactField icon={Gauge} label={t("detail.summary.model")} value={displayValue(detail.trace.model, emptyLabel)} />
                    <CompactField
                      icon={Clock}
                      label={t("detail.summary.started")}
                      value={formatLocalDateTime(detail.trace.started_at, locale, emptyLabel)}
                    />
                    {hasText(detail.trace.error_code) ? (
                      <CompactField
                        icon={AlertTriangle}
                        label={t("detail.summary.errorCode")}
                        value={detail.trace.error_code}
                        mono
                        wrap
                      />
                    ) : null}
                  </DefinitionList>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="px-4 py-3 md:px-4 md:py-3"><SectionHeader density="compact" icon={Coins} title={t("detail.usage.title")} description={t("detail.usage.description")} /></CardHeader>
                <CardContent className="px-4 pb-4 md:px-4 md:pb-4">
                  <DefinitionList columns={3} className="grid-cols-2 text-sm">
                    <CompactField
                      icon={Hash}
                      label={t("detail.usage.inputTokens")}
                      value={<CompactValue display={formatCompactNumber(detail.usage?.input_tokens ?? detail.metric?.input_tokens ?? 0, { locale })} />}
                    />
                    <CompactField
                      icon={Hash}
                      label={t("detail.usage.outputTokens")}
                      value={<CompactValue display={formatCompactNumber(detail.usage?.output_tokens ?? detail.metric?.output_tokens ?? 0, { locale })} />}
                    />
                    <CompactField
                      icon={Hash}
                      label={t("detail.usage.totalTokens")}
                      value={<CompactValue display={formatCompactNumber(detail.usage?.total_tokens ?? 0, { locale })} />}
                    />
                    <CompactField
                      icon={Coins}
                      label={t("detail.usage.costUsd")}
                      value={typeof (detail.usage?.cost_usd ?? detail.metric?.cost_usd) === "number"
                        ? <CompactValue display={formatCompactCurrency(detail.usage?.cost_usd ?? detail.metric?.cost_usd ?? 0, { locale, currency: "USD", maximumFractionDigits: 8 })} />
                        : emptyLabel}
                    />
                    <CompactField
                      icon={Coins}
                      label={t("detail.usage.costIrr")}
                      value={typeof detail.usage?.cost_irr === "number"
                        ? <CompactValue display={formatCompactIrr(detail.usage.cost_irr, { locale })} />
                        : emptyLabel}
                    />
                    <CompactField
                      icon={Route}
                      label={t("detail.usage.requestType")}
                      value={displayValue(detail.metric?.request_type, emptyLabel)}
                    />
                    {providerModel ? (
                      <CompactField
                        icon={Server}
                        label={t("detail.usage.providerModel")}
                        value={providerModel}
                      />
                    ) : null}
                    {isMeaningfulError(metricErrorType) ? (
                      <CompactField
                        icon={AlertTriangle}
                        label={t("detail.usage.errorType")}
                        value={metricErrorType}
                        mono
                        wrap
                      />
                    ) : null}
                    {showUsageLatency && typeof usageLatency === "number" ? (
                      <CompactField
                        icon={Timer}
                        label={t("detail.usage.latency")}
                        value={formatDuration(usageLatency, msLabel, locale)}
                      />
                    ) : null}
                    <CompactField
                      icon={Timer}
                      label={t("detail.usage.ttft")}
                      value={formatOptionalDuration(detail.metric?.ttft_ms, msLabel, locale, emptyLabel)}
                    />
                    <CompactField
                      icon={Timer}
                      label={t("detail.usage.itl")}
                      value={formatOptionalDuration(detail.metric?.itl_ms, msLabel, locale, emptyLabel)}
                    />
                    <CompactField
                      icon={Timer}
                      label={t("detail.usage.tpot")}
                      value={formatOptionalDuration(detail.metric?.tpot_ms, msLabel, locale, emptyLabel)}
                    />
                  </DefinitionList>
                </CardContent>
              </Card>
            </div>

            <div className="grid gap-4 xl:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]">
              <Card>
                <CardHeader className="px-4 py-3 md:px-4 md:py-3"><SectionHeader density="compact"
                  icon={RefreshCw}
                  title={t("detail.attempts.title")}
                  description={attemptsResult.ok ? t("detail.attempts.count", { count: attempts.length }) : attemptsResult.errorMessage ?? t("detail.attempts.loadFailed")}
                /></CardHeader>
                <CardContent className="px-4 pb-4 md:px-4 md:pb-4">
                  {attempts.length === 0 ? (
                    <EmptyText>{t("detail.attempts.empty")}</EmptyText>
                  ) : (
                    <div className="flex flex-col gap-2">
                      {attempts.map((attempt, idx) => (
                        <article key={`${attempt.trace_id}:${attempt.attempt_no}`} className="rounded-md border p-3">
                          <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                            <p className="font-medium">{t("detail.attempts.heading", { number: formatNumber(attempt.attempt_no, locale) })}</p>
                            <Badge variant={textStatusBadgeVariant(attempt.status)}>{attempt.status}</Badge>
                          </div>
                          <DefinitionList columns={3} className="mt-3 gap-y-2 text-sm">
                            {showAttemptRoutingFields ? (
                              <>
                                <CompactField icon={Server} label={t("detail.summary.provider")} value={displayValue(attempt.provider, emptyLabel)} />
                                {hasText(attempt.model) ? (
                                  <CompactField icon={Gauge} label={t("detail.summary.model")} value={attempt.model} />
                                ) : null}
                              </>
                            ) : null}
                            <CompactField
                              icon={Clock}
                              label={t("detail.attempts.duration")}
                              value={formatDuration(attempt.duration_ms, msLabel, locale)}
                            />
                            <CompactField
                              icon={Clock}
                              label={t("detail.attempts.started")}
                              value={formatLocalDateTime(attempt.started_at, locale, emptyLabel)}
                            />
                            {showAttemptFallbackField ? (
                              <CompactField
                                icon={RefreshCw}
                                label={t("detail.attempts.fallback")}
                                value={displayValue(attempt.fallback_reason, emptyLabel)}
                                wrap
                              />
                            ) : null}
                          </DefinitionList>
                          {attempt.retry_reason ? (
                            <div className="mt-3">
                              <CompactField icon={RefreshCw} label={t("detail.attempts.retryReason")} value={attempt.retry_reason} wrap />
                            </div>
                          ) : null}
                          {idx < attempts.length - 1 ? (
                            <p className="mt-2 text-xs text-muted-foreground">{t("detail.attempts.nextAttempt")}</p>
                          ) : null}
                        </article>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="px-4 py-3 md:px-4 md:py-3"><SectionHeader density="compact"
                  icon={Activity}
                  title={t("detail.timeline.title")}
                  description={timelineResult.ok ? undefined : timelineResult.errorMessage ?? t("detail.timeline.loadFailed")}
                /></CardHeader>
                <CardContent className="px-4 pb-4 md:px-4 md:pb-4">
                  {waterfallRows.length === 0 ? (
                    <EmptyText>{t("detail.timeline.empty")}</EmptyText>
                  ) : (
                    <div className="flex flex-col gap-1">
                      <div className="rounded-md bg-muted/40 px-3 py-1.5 text-xs text-muted-foreground">
                        {hasNonZeroSpanOffset ? t("detail.timeline.waterfallHint") : t("detail.timeline.durationHint")}
                      </div>
                      <div className="grid gap-2 px-2 md:grid-cols-[minmax(14rem,18rem)_1fr]">
                        <div dir="ltr" className="flex items-center justify-between text-xs text-muted-foreground md:hidden">
                          <span>{formatDuration(0, msLabel, locale)}</span>
                          <span>{formatDuration(timelineTotalMS, msLabel, locale)}</span>
                        </div>
                        <div className="hidden md:block" />
                        <TimelineAxis totalMS={timelineTotalMS} msLabel={msLabel} locale={locale} />
                      </div>
                      {waterfallRows.map((row, idx) => (
                        <div key={row.span.span_id} className="grid gap-2 px-2 py-1 md:grid-cols-[minmax(14rem,18rem)_1fr] md:items-center">
                          <div className="min-w-0">
                            <div className="flex min-w-0 flex-wrap items-center gap-2">
                              <TimelineStepMarker index={idx} kind={row.span.kind} total={waterfallRows.length} locale={locale} />
                              <Badge variant={textStatusBadgeVariant(row.span.status)}>{row.span.status}</Badge>
                              <p className="truncate text-sm font-medium">{row.span.name}</p>
                            </div>
                            <div className="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 pe-4 text-xs text-muted-foreground md:pe-9">
                              <span>{formatDuration(row.span.duration_ms, msLabel, locale)}</span>
                              {row.offsetMS > 0 ? <span dir="ltr">+{formatDuration(row.offsetMS, msLabel, locale)}</span> : null}
                            </div>
                          </div>
                          <div dir="ltr">
                            <div className="relative h-8 rounded-md bg-muted">
                              <TimelineGridLines />
                              <div
                                className="absolute top-1 h-6 rounded-md bg-primary/70"
                                style={{ left: `${row.leftPct}%`, width: `${row.widthPct}%` }}
                              />
                              {row.widthPct >= 12 ? (
                                <span
                                  className="absolute top-1/2 -translate-y-1/2 truncate px-2 text-xs font-medium text-primary-foreground"
                                  style={{
                                    left: `${row.leftPct}%`,
                                    width: `${row.widthPct}%`,
                                  }}
                                >
                                  {formatDuration(row.span.duration_ms, msLabel, locale)}
                                </span>
                              ) : null}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>

            <div className="grid gap-4 lg:grid-cols-2">
              <Card>
                <CardHeader className="px-4 py-3 md:px-4 md:py-3"><SectionHeader density="compact" icon={Database} title={t("detail.metadata.title")} description={t("detail.metadata.description")} /></CardHeader>
                <CardContent className="px-4 pb-4 md:px-4 md:pb-4">
                  {metadataRows.length === 0 ? (
                    <EmptyText>{t("detail.metadata.empty")}</EmptyText>
                  ) : (
                    <div className="overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>{t("detail.metadata.columns.source")}</TableHead>
                            <TableHead>{t("detail.metadata.columns.key")}</TableHead>
                            <TableHead>{t("detail.metadata.columns.value")}</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {metadataRows.map((row) => (
                            <TableRow key={`${row.source}:${row.key}`}>
                              <TableCell className="whitespace-nowrap text-xs text-muted-foreground">{metadataSourceLabels[row.source]}</TableCell>
                              <TableCell className="max-w-[12rem] break-all font-mono text-xs text-muted-foreground">{row.key}</TableCell>
                              <TableCell className="max-w-[20rem] break-all font-mono text-xs">{row.value}</TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="px-4 py-3 md:px-4 md:py-3"><SectionHeader density="compact" icon={FileJson} title={t("detail.errors.title")} description={t("detail.errors.description")} /></CardHeader>
                <CardContent className="px-4 pb-4 md:px-4 md:pb-4">
                  {errorRows.length === 0 ? (
                    <EmptyText>{t("detail.errors.empty")}</EmptyText>
                  ) : (
                    <div className="overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>{t("detail.errors.columns.source")}</TableHead>
                            <TableHead>{t("detail.errors.columns.message")}</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {errorRows.map((row, idx) => (
                            <TableRow key={`${row.source}:${idx}`}>
                              <TableCell className="whitespace-nowrap text-xs font-medium">{row.source}</TableCell>
                              <TableCell className="max-w-[24rem] break-words text-xs text-muted-foreground">{row.message}</TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </>
        )}
      </PageContainer>
    </AdminShell>
  );
}

function sortSpans(spans: TraceSpan[]) {
  return [...spans].sort((a, b) => {
    const aMS = new Date(a.started_at).getTime();
    const bMS = new Date(b.started_at).getTime();
    return aMS - bMS;
  });
}

function buildWaterfallRows(spans: TraceSpan[]) {
  if (spans.length === 0) return [];
  const startMS = Math.min(...spans.map((s) => new Date(s.started_at).getTime()));
  const endMS = Math.max(...spans.map((s) => new Date(s.started_at).getTime() + s.duration_ms));
  const totalMS = Math.max(1, endMS - startMS);
  return spans.map((span) => {
    const spanStart = new Date(span.started_at).getTime();
    const offsetMS = Math.max(0, spanStart - startMS);
    const leftPct = (offsetMS / totalMS) * 100;
    const widthPct = Math.max(1.5, (Math.max(span.duration_ms, 1) / totalMS) * 100);
    return { span, offsetMS, leftPct, widthPct, totalMS };
  });
}

function buildMetadataRows(
  rootAttributes?: Record<string, unknown>,
  usageMetadata?: Record<string, unknown>,
): MetadataRow[] {
  const out: MetadataRow[] = [];
  for (const [key, value] of Object.entries(rootAttributes ?? {})) {
    out.push({ source: "span", key, value: stringifyValue(value) });
  }
  for (const [key, value] of Object.entries(usageMetadata ?? {})) {
    out.push({ source: "usage", key, value: stringifyValue(value) });
  }
  return out.sort((a, b) => `${a.source}.${a.key}`.localeCompare(`${b.source}.${b.key}`));
}

function buildErrorRows(
  detail: TraceDetailV2Response | null,
  spans: TraceSpan[],
  t: Awaited<ReturnType<typeof getTranslations>>,
  locale: string,
): Array<{ source: string; message: string }> {
  if (!detail) return [];
  const out: Array<{ source: string; message: string }> = [];
  if (hasText(detail.trace.error_code)) {
    out.push({ source: t("detail.errors.sources.traceErrorCode"), message: detail.trace.error_code });
  }
  if (isMeaningfulError(detail.metric?.error_type)) {
    out.push({ source: t("detail.errors.sources.metricErrorType"), message: detail.metric?.error_type ?? "" });
  }
  for (const span of spans) {
    if (span.status !== "error") continue;
    const hint = spanErrorHint(span.attributes);
    out.push({
      source: t("detail.errors.sources.span", { name: span.name }),
      message: hint || t("detail.errors.statusErrorDuration", { duration: formatDuration(span.duration_ms, t("detail.units.ms"), locale) }),
    });
  }
  return out;
}

function spanErrorHint(attrs: Record<string, unknown> | undefined) {
  if (!attrs) return "";
  for (const key of ["error", "error_message", "failure_reason", "provider_error", "retry_reason"]) {
    const value = attrs[key];
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return "";
}

function stringifyValue(value: unknown): string {
  if (value == null) return "null";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function shouldShowMetadataRow(row: MetadataRow, detail: TraceDetailV2Response | null) {
  const key = row.key.trim().toLowerCase();
  if (row.source === "span" && ["http.method", "http.route", "sampled"].includes(key)) {
    return false;
  }
  if (row.source === "usage" && ["model", "request_id", "trace_id"].includes(key)) {
    return false;
  }
  if (detail && row.source === "usage" && key === "provider_model_name" && sameText(row.value, detail.metric?.provider_model_name)) {
    return false;
  }
  return true;
}

function distinctValue(value: string | undefined, duplicate: string | undefined) {
  if (!hasText(value)) return "";
  return sameText(value, duplicate) ? "" : value.trim();
}

function hasText(value: string | undefined): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function sameText(a: string | undefined, b: string | undefined) {
  if (!hasText(a) || !hasText(b)) return false;
  return a.trim().toLowerCase() === b.trim().toLowerCase();
}

function isMeaningfulError(value: string | undefined): value is string {
  if (!hasText(value)) return false;
  const normalized = value.trim().toLowerCase();
  return !["-", "none", "null", "ok", "success"].includes(normalized);
}

function isNearNumber(a: number | undefined, b: number | undefined, tolerance: number) {
  if (typeof a !== "number" || typeof b !== "number") return false;
  if (!Number.isFinite(a) || !Number.isFinite(b)) return false;
  return Math.abs(a - b) <= tolerance;
}

function formatNumber(value: number, locale: string) {
  return new Intl.NumberFormat(locale).format(value);
}

function formatDuration(value: number, msLabel: string, locale: string) {
  if (!Number.isFinite(value)) return `0 ${msLabel}`;
  return `${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)} ${msLabel}`;
}

function formatOptionalDuration(value: number | undefined, msLabel: string, locale: string, fallback: string) {
  if (typeof value !== "number" || !Number.isFinite(value)) return fallback;
  return formatDuration(value, msLabel, locale);
}

function formatLocalDateTime(value: string, locale: string, fallback: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return fallback;
  return new Intl.DateTimeFormat(locale, {
    timeZone: "Asia/Tehran",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

function displayValue(value: string | undefined, fallback: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : fallback;
}

function statusCodeBadgeVariant(statusCode: number): BadgeVariant {
  if (statusCode >= 500) return "destructive";
  if (statusCode >= 400) return "warning";
  if (statusCode >= 300) return "info";
  return "success";
}

function textStatusBadgeVariant(status: string): BadgeVariant {
  const normalized = status.trim().toLowerCase();
  if (["error", "failed", "failure"].includes(normalized)) return "destructive";
  if (["retry", "fallback", "timeout"].includes(normalized)) return "warning";
  if (["ok", "success", "succeeded", "complete", "completed"].includes(normalized)) return "success";
  return "secondary";
}

function MetricItem({
  icon: Icon,
  label,
  value,
  valueClassName,
}: {
  icon: LucideIcon;
  label: string;
  value: ReactNode;
  valueClassName?: string;
}) {
  return (
    <div className="min-w-0 rounded-md bg-muted/40 px-3 py-2">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Icon className="size-4 shrink-0" aria-hidden="true" />
        <span className="truncate">{label}</span>
      </div>
      <div className={cn("mt-1 truncate text-sm font-semibold", valueClassName)}>{value}</div>
    </div>
  );
}

function EmptyText({ children }: { children: ReactNode }) {
  return <p className="rounded-md bg-muted/40 px-3 py-2 text-sm text-muted-foreground">{children}</p>;
}

function TimelineStepMarker({
  index,
  kind,
  total,
  locale,
}: {
  index: number;
  kind: string;
  total: number;
  locale: string;
}) {
  return (
    <span className="relative flex shrink-0 items-center gap-1.5">
      {index < total - 1 ? (
        <span className="absolute start-3 top-6 h-7 border-s border-border" aria-hidden="true" />
      ) : null}
      <span className="flex size-6 items-center justify-center rounded-full border bg-background text-xs font-semibold">
        {formatNumber(index + 1, locale)}
      </span>
      <Badge variant={spanKindBadgeVariant(kind)}>{kind}</Badge>
    </span>
  );
}

function spanKindBadgeVariant(kind: string): BadgeVariant {
  const normalized = kind.trim().toLowerCase();
  if (normalized === "server") return "info";
  if (normalized === "client") return "warning";
  if (normalized === "internal") return "secondary";
  return "outline";
}

function TimelineAxis({ totalMS, msLabel, locale }: { totalMS: number; msLabel: string; locale: string }) {
  const ticks = [0, 25, 50, 75, 100];
  return (
    <div dir="ltr" className="hidden md:block">
      <div className="relative h-6">
        {ticks.map((tick) => (
          <span
            key={tick}
            className={cn(
              "absolute top-0 text-xs text-muted-foreground",
              tick === 0 ? "translate-x-0" : tick === 100 ? "-translate-x-full" : "-translate-x-1/2",
            )}
            style={{ insetInlineStart: `${tick}%` }}
          >
            {formatDuration((totalMS * tick) / 100, msLabel, locale)}
          </span>
        ))}
      </div>
    </div>
  );
}

function TimelineGridLines() {
  return (
    <>
      {[25, 50, 75].map((tick) => (
        <span
          key={tick}
          className="absolute top-0 h-full border-s border-background/80"
          style={{ insetInlineStart: `${tick}%` }}
          aria-hidden="true"
        />
      ))}
    </>
  );
}

function buildMetricsLink(trace: RequestTraceV2) {
  const started = new Date(trace.started_at);
  const from = new Date(started.getTime() - 15 * 60 * 1000).toISOString();
  const to = new Date(started.getTime() + 15 * 60 * 1000).toISOString();
  const params = new URLSearchParams();
  params.set("from", from);
  params.set("to", to);
  if (trace.model) params.set("model", trace.model);
  if (trace.provider) params.set("provider", trace.provider);
  if (trace.request_id) params.set("request_id", trace.request_id);
  return `/metrics?${params.toString()}`;
}
