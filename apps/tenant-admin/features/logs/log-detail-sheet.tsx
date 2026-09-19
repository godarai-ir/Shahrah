"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { memo } from "react";
import { Filter, Waypoints } from "lucide-react";
import { useTranslations } from "next-intl";

import { CompactValue } from "@/components/app-ui/CompactValue";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import type { TimingSpan } from "./api";
import type { LogDetail, LogsFormatter, RequestLog } from "./logs-types";
import { buildTimingBreakdown, payloadStatus, requestSignals } from "./logs-utils";

const JsonViewer = dynamic(
  () => import("@/components/app-ui/JsonViewer").then((module) => module.JsonViewer),
  {
    ssr: false,
    loading: () => <div className="min-h-40 animate-pulse rounded-xl border bg-muted/35" aria-hidden="true" />,
  },
);

/** Logging modes and their copy in settings.requestLogging.modes. */
const PAYLOAD_MODE_KEYS: Record<string, string> = {
  none: "none",
  "metadata-only": "metadataOnly",
  "redacted-content": "redactedContent",
  "full-content": "fullContent",
};

export type LogDetailSheetProps = {
  selectedLog: RequestLog | null;
  detail: LogDetail | null;
  /** Spans of the log's sampled trace. Null when there is no trace, it was not sampled, or it has not loaded. */
  timingSpans: TimingSpan[] | null;
  /** Link to the trace in the traces tab; null when the log has no trace that exists. */
  traceHref: string | null;
  onFilterModel: (model: string) => void;
  onClose: () => void;
  formatters: LogsFormatter;
  closeLabel: string;
  t: ReturnType<typeof useTranslations>;
};

export const LogDetailSheet = memo(function LogDetailSheet({
  selectedLog,
  detail,
  timingSpans,
  traceHref,
  onFilterModel,
  onClose,
  formatters,
  closeLabel,
  t,
}: LogDetailSheetProps) {
  const model = detail?.log.model?.trim() ?? "";

  return (
    <Sheet open={selectedLog !== null} onOpenChange={(open) => { if (!open) onClose(); }}>
      <SheetContent side="left" closeLabel={closeLabel} className="flex w-full flex-col sm:max-w-2xl">
        <SheetHeader>
          <SheetTitle>{t("detailTitle")}</SheetTitle>
          <SheetDescription asChild>
            <span className="flex min-w-0 items-center justify-end gap-1" dir="ltr">
              <span className="truncate font-mono">{selectedLog?.log_id}</span>
              {selectedLog?.log_id ? (
                <CopyAction
                  value={selectedLog.log_id}
                  copyLabel={t("actions.copyLogId")}
                  copyingLabel={t("actions.copyingLogId")}
                  copiedLabel={t("actions.copiedLogId")}
                  size="icon-xs"
                />
              ) : null}
            </span>
          </SheetDescription>
        </SheetHeader>
        {detail ? (
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
            <ErrorBlock log={detail.log} t={t} />
            <div className="grid grid-cols-3 gap-2">
              <MetricCard label={t("table.status")} value={detail.log.status_code} />
              <MetricCard label={t("table.latency")} value={formatters.ms(detail.log.duration_ms)} />
              <MetricCard label={t("table.cost")} value={<CompactValue display={formatters.usdDisplay(detail.log.cost_usd ?? 0)} />} />
            </div>
            <SignalChips log={detail.log} formatters={formatters} t={t} />
            {detail.log.trace_id && timingSpans ? <TimingBreakdown spans={timingSpans} formatters={formatters} t={t} /> : null}
            <Tabs defaultValue="summary" className="min-w-0">
              <TabsList variant="line" aria-label={t("detailTitle")} className="w-full justify-start overflow-x-auto">
                <TabsTrigger value="summary">{t("tabs.summary")}</TabsTrigger>
                <TabsTrigger value="request">{t("content.request")}</TabsTrigger>
                <TabsTrigger value="response">{t("content.response")}</TabsTrigger>
                <TabsTrigger value="metadata">{t("tabs.metadata")}</TabsTrigger>
              </TabsList>
              <TabsContent value="summary"><JSONBlock label={t("tabs.summary")} value={{ log: detail.log, actor: detail.actor, scope: detail.scope, content_visibility: detail.content_visibility }} /></TabsContent>
              <TabsContent value="request" className="flex flex-col gap-2">
                <PayloadModeNote mode={detail.log.payload_mode} t={t} />
                <PayloadBlock label={t("content.request")} available={detail.content_visibility.request_available} value={detail.request_payload} reason={detail.content_visibility.request_reason} hidden={t("content.hidden")} />
              </TabsContent>
              <TabsContent value="response" className="flex flex-col gap-2">
                <PayloadModeNote mode={detail.log.payload_mode} t={t} />
                <PayloadBlock label={t("content.response")} available={detail.content_visibility.response_available} value={detail.response_payload} reason={detail.content_visibility.response_reason} hidden={t("content.hidden")} />
              </TabsContent>
              <TabsContent value="metadata"><JSONBlock label={t("tabs.metadata")} value={{ metadata: detail.metadata, request_payload_meta: detail.request_payload_meta, response_payload_meta: detail.response_payload_meta }} /></TabsContent>
            </Tabs>
          </div>
        ) : (
          <EmptyState title={t("detailLoading")} description={t("detailLoadingDescription")} />
        )}
        {detail && (model || traceHref) ? (
          <SheetFooter className="flex-row flex-wrap justify-end gap-2 border-t border-border pt-4">
            {model ? (
              <Button type="button" variant="outline" size="sm" onClick={() => onFilterModel(model)}>
                <Filter data-icon="inline-start" aria-hidden="true" />
                {t("actions.filterModel")}
              </Button>
            ) : null}
            {traceHref ? (
              <Button asChild size="sm">
                <Link href={traceHref}>
                  <Waypoints data-icon="inline-start" aria-hidden="true" />
                  {t("actions.viewTrace")}
                </Link>
              </Button>
            ) : null}
          </SheetFooter>
        ) : null}
      </SheetContent>
    </Sheet>
  );
});

/** Shown only for failed requests: the error type large, and what the status class means. */
function ErrorBlock({ log, t }: { log: RequestLog; t: ReturnType<typeof useTranslations> }) {
  if (log.status_code < 400) return null;
  const errorType = log.error_type?.trim();
  return (
    <Alert variant="destructive" data-testid="log-error-block">
      <AlertTitle className="break-all font-mono text-base" dir="ltr">{errorType || log.status_code}</AlertTitle>
      <AlertDescription>
        {log.status_code >= 500 ? t("detail.error.serverError", { status: log.status_code }) : t("detail.error.clientError", { status: log.status_code })}
      </AlertDescription>
    </Alert>
  );
}

function SignalChips({ log, formatters, t }: { log: RequestLog; formatters: LogsFormatter; t: ReturnType<typeof useTranslations> }) {
  const { retries, cacheHit, fallback } = requestSignals(log);
  if (!retries && !cacheHit && !fallback) return null;
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs text-muted-foreground">{t("detail.signals")}</span>
      {retries ? <StatusBadge status="warning">{t("signals.retry", { count: formatters.number(retries) })}</StatusBadge> : null}
      {cacheHit ? <StatusBadge status="info">{t("signals.cache")}</StatusBadge> : null}
      {fallback ? <StatusBadge status="warning">{t("signals.fallback")}</StatusBadge> : null}
    </div>
  );
}

const SEGMENT_TONES = ["bg-chart-1", "bg-chart-2", "bg-chart-3", "bg-chart-4", "bg-chart-5"];

/** Real span timings of the log's trace. Rendered only when the trace was sampled; never estimated. */
function TimingBreakdown({ spans, formatters, t }: { spans: TimingSpan[]; formatters: LogsFormatter; t: ReturnType<typeof useTranslations> }) {
  const breakdown = buildTimingBreakdown(spans);
  if (!breakdown) return null;
  const percent = (share: number) => `${(share * 100).toFixed(2)}%`;

  return (
    <section aria-label={t("detail.timing.title")} data-testid="log-timing-breakdown" className="flex flex-col gap-2 rounded-lg border border-border p-3">
      <div className="flex items-baseline justify-between gap-2">
        <h3 className="text-sm font-medium">{t("detail.timing.title")}</h3>
        <span className="text-xs text-muted-foreground">{t("detail.timing.total", { duration: formatters.ms(breakdown.totalMs) })}</span>
      </div>
      {breakdown.stackable ? (
        <div className="flex h-2.5 overflow-hidden rounded-full bg-muted" dir="ltr" aria-hidden="true">
          {breakdown.segments.map((segment, index) => (
            <span key={segment.spanID} className={SEGMENT_TONES[index % SEGMENT_TONES.length]} style={{ width: percent(segment.share) }} />
          ))}
        </div>
      ) : null}
      <ul className="flex flex-col gap-1.5">
        {breakdown.segments.map((segment, index) => (
          <li key={segment.spanID} className="grid grid-cols-2 items-center gap-2 text-xs">
            <span className="flex min-w-0 items-center gap-2">
              <span className={`size-2 shrink-0 rounded-full ${SEGMENT_TONES[index % SEGMENT_TONES.length]}`} aria-hidden="true" />
              <span className="truncate font-mono" dir="ltr">{segment.name}</span>
            </span>
            <span className="flex items-center gap-2">
              <span className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted" dir="ltr" aria-hidden="true">
                <span className={`block h-full rounded-full ${SEGMENT_TONES[index % SEGMENT_TONES.length]}`} style={{ width: percent(segment.share) }} />
              </span>
              <span className="w-16 shrink-0 text-end tabular-nums">{formatters.ms(segment.durationMs)}</span>
            </span>
          </li>
        ))}
      </ul>
      <p className="text-xs text-muted-foreground">{t("detail.timing.source")}</p>
    </section>
  );
}

/** Says under which logging mode this request's payload was recorded, from the log itself. */
function PayloadModeNote({ mode, t }: { mode?: string; t: ReturnType<typeof useTranslations> }) {
  const modesT = useTranslations("settings.requestLogging.modes");
  const trimmed = mode?.trim();
  if (!trimmed) return null;
  const key = PAYLOAD_MODE_KEYS[trimmed];
  return (
    <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
      <StatusBadge status={payloadStatus(trimmed)}>{key ? modesT(`${key}.title`) : trimmed}</StatusBadge>
      <span>{key ? `${t("detail.payloadMode", { mode: modesT(`${key}.title`) })} — ${modesT(`${key}.description`)}` : t("detail.payloadModeUnknown", { mode: trimmed })}</span>
    </div>
  );
}

function JSONBlock({ label, value }: { label: string; value: unknown }) {
  return <JsonViewer aria-label={label} value={value} maxHeight="520px" />;
}

function PayloadBlock({ label, available, value, reason, hidden }: { label: string; available: boolean; value?: unknown; reason?: string; hidden: string }) {
  if (!available || value === undefined || value === null) {
    return <Alert><AlertTitle>{hidden}</AlertTitle><AlertDescription>{reason}</AlertDescription></Alert>;
  }
  return <JSONBlock label={label} value={value} />;
}
