import Link from "next/link";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type RequestTraceV2 = {
  trace_id: string;
  tenant_id: string;
  request_id?: string;
  model?: string;
  provider?: string;
  route: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  error_code?: string;
  sampled: boolean;
  attempts: number;
};

type TraceSummaryV2 = {
  spans_count: number;
  attempts_count: number;
  errors_count: number;
};

type TraceUsageV2 = {
  model?: string;
  input_tokens?: number;
  output_tokens?: number;
  total_tokens?: number;
  cost_usd?: number;
  cost_irr?: number;
  response_time_ms?: number;
  status?: string;
  metadata?: Record<string, unknown>;
};

type TraceMetricV2 = {
  request_type?: string;
  provider_name?: string;
  provider_model_name?: string;
  error_type?: string;
  latency_ms?: number;
  input_tokens?: number;
  output_tokens?: number;
  cost_usd?: number;
  ttft_ms?: number;
  itl_ms?: number;
  tpot_ms?: number;
  virtual_model_name?: string;
};

type TraceDetailV2Response = {
  trace: RequestTraceV2;
  summary: TraceSummaryV2;
  usage?: TraceUsageV2;
  metric?: TraceMetricV2;
};

type TraceSpan = {
  span_id: string;
  trace_id: string;
  parent_span_id?: string;
  name: string;
  kind: string;
  started_at: string;
  duration_ms: number;
  status: string;
  attributes: Record<string, unknown>;
};

type TraceAttempt = {
  attempt_no: number;
  trace_id: string;
  provider?: string;
  model?: string;
  started_at: string;
  duration_ms: number;
  status: string;
  retry_reason?: string;
  fallback_reason?: string;
  attributes?: Record<string, unknown>;
};

type SpanResponse = { data?: TraceSpan[] };
type AttemptResponse = { data?: TraceAttempt[] };

type Params = { params: Promise<{ traceID: string }> };

export const dynamic = "force-dynamic";

export default async function TraceDetailPage({ params }: Params) {
  const t = await getTranslations("traces");
  const auth = await requireAuth("/login?next=/traces");
  const tenantSlug = auth.tenantSlug;
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

  const detail = (traceResult.ok ? traceResult.payload : null) as TraceDetailV2Response | null;
  const spansPayload = (timelineResult.payload ?? {}) as SpanResponse;
  const attemptsPayload = (attemptsResult.payload ?? {}) as AttemptResponse;
  const spans = timelineResult.ok ? spansPayload.data ?? [] : [];
  const attempts = attemptsResult.ok ? attemptsPayload.data ?? [] : [];
  const sortedSpans = sortSpans(spans);
  const waterfallRows = buildWaterfallRows(sortedSpans);
  const rootSpan = sortedSpans.find((span) => span.name === "gateway.request" && span.kind === "server");

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  const metadataRows = buildMetadataRows(rootSpan?.attributes, detail?.usage?.metadata);
  const errorRows = buildErrorRows(detail, sortedSpans, t);

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold">{t("detail.title")}</h1>
          <div className="flex items-center gap-4">
            {detail ? (
              <Link href={buildMetricsLink(detail.trace)} className="text-sm text-primary underline-offset-2 hover:underline">
                {t("detail.actions.openInMetrics")}
              </Link>
            ) : null}
            <Link href="/traces" className="text-sm text-primary underline-offset-2 hover:underline">
              {t("detail.actions.backToTraces")}
            </Link>
          </div>
        </div>

        {!detail ? (
          <Card>
            <CardHeader>
              <CardTitle>{t("detail.notFound.title")}</CardTitle>
              <CardDescription>{traceResult.errorMessage ?? t("detail.notFound.description")}</CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <>
            <Card>
              <CardHeader>
                <CardTitle className="font-mono text-sm">{detail.trace.trace_id}</CardTitle>
                <CardDescription>{t("detail.summary.title")}</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-2 text-sm md:grid-cols-4">
                <p><strong>{t("detail.summary.requestId")}:</strong> {detail.trace.request_id || "-"}</p>
                <p><strong>{t("detail.summary.route")}:</strong> {detail.trace.route}</p>
                <p><strong>{t("detail.summary.model")}:</strong> {detail.trace.model || "-"}</p>
                <p><strong>{t("detail.summary.provider")}:</strong> {detail.trace.provider || "-"}</p>
                <p><strong>{t("detail.summary.status")}:</strong> <Badge variant={detail.trace.status_code >= 500 ? "destructive" : "secondary"}>{detail.trace.status_code}</Badge></p>
                <p><strong>{t("detail.summary.duration")}:</strong> {detail.trace.duration_ms} {t("detail.units.ms")}</p>
                <p><strong>{t("detail.summary.started")}:</strong> {new Date(detail.trace.started_at).toLocaleString()}</p>
                <p><strong>{t("detail.summary.sampled")}:</strong> {detail.trace.sampled ? t("detail.values.yes") : t("detail.values.no")}</p>
                <p><strong>{t("detail.summary.attempts")}:</strong> {detail.summary.attempts_count}</p>
                <p><strong>{t("detail.summary.spans")}:</strong> {detail.summary.spans_count}</p>
                <p><strong>{t("detail.summary.errors")}:</strong> {detail.summary.errors_count}</p>
                <p><strong>{t("detail.summary.errorCode")}:</strong> {detail.trace.error_code || "-"}</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>{t("detail.usage.title")}</CardTitle>
                <CardDescription>{t("detail.usage.description")}</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-2 text-sm md:grid-cols-4">
                <p><strong>{t("detail.usage.inputTokens")}:</strong> {detail.usage?.input_tokens ?? detail.metric?.input_tokens ?? 0}</p>
                <p><strong>{t("detail.usage.outputTokens")}:</strong> {detail.usage?.output_tokens ?? detail.metric?.output_tokens ?? 0}</p>
                <p><strong>{t("detail.usage.totalTokens")}:</strong> {detail.usage?.total_tokens ?? 0}</p>
                <p><strong>{t("detail.usage.costUsd")}:</strong> {formatCost(detail.usage?.cost_usd ?? detail.metric?.cost_usd)}</p>
                <p><strong>{t("detail.usage.costIrr")}:</strong> {formatNumber(detail.usage?.cost_irr)}</p>
                <p><strong>{t("detail.usage.requestType")}:</strong> {detail.metric?.request_type || "-"}</p>
                <p><strong>{t("detail.usage.providerModel")}:</strong> {detail.metric?.provider_model_name || "-"}</p>
                <p><strong>{t("detail.usage.errorType")}:</strong> {detail.metric?.error_type || "-"}</p>
                <p><strong>{t("detail.usage.latency")}:</strong> {detail.metric?.latency_ms ?? detail.usage?.response_time_ms ?? 0} {t("detail.units.ms")}</p>
                <p><strong>{t("detail.usage.ttft")}:</strong> {detail.metric?.ttft_ms ?? "-"} {t("detail.units.ms")}</p>
                <p><strong>{t("detail.usage.itl")}:</strong> {detail.metric?.itl_ms ?? "-"} {t("detail.units.ms")}</p>
                <p><strong>{t("detail.usage.tpot")}:</strong> {detail.metric?.tpot_ms ?? "-"} {t("detail.units.ms")}</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>{t("detail.attempts.title")}</CardTitle>
                <CardDescription>{attemptsResult.ok ? t("detail.attempts.count", { count: attempts.length }) : attemptsResult.errorMessage ?? t("detail.attempts.loadFailed")}</CardDescription>
              </CardHeader>
              <CardContent>
                {attempts.length === 0 ? (
                  <p className="text-sm text-muted-foreground">{t("detail.attempts.empty")}</p>
                ) : (
                  <div className="space-y-3">
                    {attempts.map((attempt, idx) => (
                      <div key={`${attempt.trace_id}:${attempt.attempt_no}`} className="rounded-lg border p-3">
                        <div className="flex items-center justify-between">
                          <p className="font-medium">
                            {t("detail.attempts.attempt")} #{attempt.attempt_no} - {attempt.provider || "-"} - {attempt.model || "-"}
                          </p>
                          <Badge variant={attempt.status === "error" ? "destructive" : "secondary"}>{attempt.status}</Badge>
                        </div>
                        <div className="mt-2 grid gap-1 text-sm md:grid-cols-3">
                          <p><strong>{t("detail.attempts.duration")}:</strong> {attempt.duration_ms} {t("detail.units.ms")}</p>
                          <p><strong>{t("detail.attempts.started")}:</strong> {new Date(attempt.started_at).toLocaleString()}</p>
                          <p><strong>{t("detail.attempts.fallback")}:</strong> {attempt.fallback_reason || "-"}</p>
                        </div>
                        {attempt.retry_reason ? <p className="mt-2 text-xs text-muted-foreground"><strong>{t("detail.attempts.retryReason")}:</strong> {attempt.retry_reason}</p> : null}
                        {idx < attempts.length - 1 ? <p className="mt-2 text-xs text-muted-foreground">{t("detail.attempts.nextAttempt")}</p> : null}
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>{t("detail.timeline.title")}</CardTitle>
                <CardDescription>{timelineResult.ok ? t("detail.timeline.count", { count: sortedSpans.length }) : timelineResult.errorMessage ?? t("detail.timeline.loadFailed")}</CardDescription>
              </CardHeader>
              <CardContent>
                {waterfallRows.length === 0 ? (
                  <p className="text-sm text-muted-foreground">{t("detail.timeline.empty")}</p>
                ) : (
                  <div className="space-y-3">
                    {waterfallRows.map((row) => (
                      <div key={row.span.span_id} className="grid grid-cols-12 items-center gap-3">
                        <div className="col-span-4 md:col-span-3">
                          <p className="text-sm font-medium">{row.span.name}</p>
                          <p className="text-xs text-muted-foreground">
                            {row.span.kind} - {row.span.duration_ms} {t("detail.units.ms")} - {row.offsetMS} {t("detail.units.ms")}
                          </p>
                        </div>
                        <div className="col-span-8 md:col-span-9">
                          <div className="relative h-7 rounded-md bg-muted">
                            <div
                              className="absolute top-0 h-7 rounded-md bg-primary/70"
                              style={{ left: `${row.leftPct}%`, width: `${row.widthPct}%` }}
                            />
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            <div className="grid gap-6 lg:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>{t("detail.metadata.title")}</CardTitle>
                  <CardDescription>{t("detail.metadata.description")}</CardDescription>
                </CardHeader>
                <CardContent>
                  {metadataRows.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{t("detail.metadata.empty")}</p>
                  ) : (
                    <div className="space-y-2">
                      {metadataRows.map((row) => (
                        <div key={row.key} className="grid grid-cols-12 gap-2 border-b py-2 text-xs">
                          <p className="col-span-4 font-mono text-muted-foreground">{row.key}</p>
                          <p className="col-span-8 break-all font-mono">{row.value}</p>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>{t("detail.errors.title")}</CardTitle>
                  <CardDescription>{t("detail.errors.description")}</CardDescription>
                </CardHeader>
                <CardContent>
                  {errorRows.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{t("detail.errors.empty")}</p>
                  ) : (
                    <div className="space-y-3">
                      {errorRows.map((row, idx) => (
                        <div key={`${row.source}:${idx}`} className="rounded-md border p-3">
                          <p className="text-sm font-medium">{row.source}</p>
                          <p className="text-xs text-muted-foreground mt-1">{row.message}</p>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </>
        )}
      </div>
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
    return { span, offsetMS, leftPct, widthPct };
  });
}

function buildMetadataRows(
  rootAttributes?: Record<string, unknown>,
  usageMetadata?: Record<string, unknown>,
): Array<{ key: string; value: string }> {
  const out: Array<{ key: string; value: string }> = [];
  for (const [key, value] of Object.entries(rootAttributes ?? {})) {
    out.push({ key: `span.${key}`, value: stringifyValue(value) });
  }
  for (const [key, value] of Object.entries(usageMetadata ?? {})) {
    out.push({ key: `usage.${key}`, value: stringifyValue(value) });
  }
  return out.sort((a, b) => a.key.localeCompare(b.key));
}

function buildErrorRows(
  detail: TraceDetailV2Response | null,
  spans: TraceSpan[],
  t: Awaited<ReturnType<typeof getTranslations>>,
): Array<{ source: string; message: string }> {
  if (!detail) return [];
  const out: Array<{ source: string; message: string }> = [];
  if (detail.trace.error_code) {
    out.push({ source: "trace.error_code", message: detail.trace.error_code });
  }
  if (detail.metric?.error_type) {
    out.push({ source: "metric.error_type", message: detail.metric.error_type });
  }
  for (const span of spans) {
    if (span.status !== "error") continue;
    const hint = spanErrorHint(span.attributes);
    out.push({
      source: `span.${span.name}`,
      message: hint || t("detail.errors.statusErrorDuration", { duration: span.duration_ms }),
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

function formatCost(v?: number) {
  if (typeof v !== "number") return "-";
  return v.toFixed(8);
}

function formatNumber(v?: number) {
  if (typeof v !== "number") return "-";
  return v.toLocaleString();
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
