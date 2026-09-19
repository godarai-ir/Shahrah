"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ExternalLink, EyeOff, X } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { EmptyState } from "@/components/app-ui/EmptyState";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

import { getTraceAttempts, getTraceDetail, getTraceTimeline, type TracePrivacySettings } from "./api";
import type { TraceAttempt, TraceDetailV2Response, TraceSpan } from "./detail-contracts";
import type { RequestTrace } from "./trace-types";
import { compactID, formatDuration, formatLocalDateTime, statusBadgeVariant } from "./trace-utils";
import { TraceWaterfall } from "./TraceWaterfall";
import { textStatusBadgeVariant } from "./trace-waterfall";

type CapturePolicy = TracePrivacySettings["capture_policy"];

type PanelState =
  | { status: "loading" }
  | { status: "error" }
  | { status: "ready"; detail: TraceDetailV2Response | null; spans: TraceSpan[]; attempts: TraceAttempt[]; sampled: boolean };

const POLICY_KEYS: Record<CapturePolicy, string> = {
  none: "none",
  "metadata-only": "metadataOnly",
  "sampled-redacted": "sampledRedacted",
  "full-secure": "fullSecure",
};

type TraceDetailPanelProps = {
  traceID: string;
  /** The list row, when the selected trace is on the current page. */
  item: RequestTrace | null;
  capturePolicy: CapturePolicy | null;
  onClose?: () => void;
};

/**
 * Detail side of the /traces master–detail view. Loads one trace on selection only: the summary and
 * attempts always, the span timeline only when the trace was sampled. Render it with `key={traceID}`;
 * switching selection then unmounts the old panel, which aborts its requests.
 */
export function TraceDetailPanel({ traceID, item, capturePolicy, onClose }: TraceDetailPanelProps) {
  const t = useTranslations("traces");
  const policyT = useTranslations("settings.tracePrivacy.policies");
  const locale = useLocale();
  const [state, setState] = useState<PanelState>({ status: "loading" });
  // Captured at selection time, so a list refresh does not refetch the open trace.
  const [knownSampled] = useState(item?.sampled);

  useEffect(() => {
    const controller = new AbortController();
    const { signal } = controller;

    async function load() {
      // The summary and attempts only decorate the panel; the timeline is its content and may fail loudly.
      const [detail, attempts] = await Promise.all([
        getTraceDetail(traceID, signal).catch(() => null),
        getTraceAttempts(traceID, signal).catch((): TraceAttempt[] => []),
      ]);
      if (knownSampled === undefined && !detail) throw new Error("trace summary unavailable");
      const sampled = knownSampled ?? detail?.trace.sampled ?? false;
      const spans = sampled ? await getTraceTimeline(traceID, signal) : [];
      return { detail, attempts, spans, sampled };
    }

    load()
      .then((result) => {
        if (!signal.aborted) setState({ status: "ready", ...result });
      })
      .catch(() => {
        if (!signal.aborted) setState({ status: "error" });
      });
    return () => controller.abort();
  }, [knownSampled, traceID]);

  const trace = item ?? (state.status === "ready" ? state.detail?.trace ?? null : null);
  const msLabel = t("list.units.ms");
  const sLabel = t("list.units.s");

  return (
    <section aria-label={t("panel.title")} className="flex min-w-0 flex-col gap-3">
      <header className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            {trace ? <Badge variant={statusBadgeVariant(trace.status_code)}>{trace.status_code}</Badge> : null}
            <h2 className="truncate font-mono text-sm font-semibold" title={traceID} dir="ltr">{compactID(traceID)}</h2>
          </div>
          {trace ? (
            <p className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
              <span>{formatLocalDateTime(trace.started_at, locale)}</span>
              <span>{formatDuration(trace.duration_ms, msLabel, sLabel, locale)}</span>
              {trace.model ? <span dir="auto">{trace.model}</span> : null}
              {trace.provider ? <span dir="auto">{trace.provider}</span> : null}
              {state.status === "ready" && state.detail ? (
                <span>{t("detail.timeline.count", { count: state.detail.summary.spans_count })}</span>
              ) : null}
            </p>
          ) : null}
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <Button asChild variant="ghost" size="sm">
            <Link href={`/traces/${encodeURIComponent(traceID)}`}>
              <ExternalLink data-icon="inline-start" aria-hidden="true" />
              {t("panel.openFull")}
            </Link>
          </Button>
          {onClose ? (
            <Button type="button" variant="ghost" size="icon-sm" onClick={onClose} aria-label={t("panel.close")}>
              <X aria-hidden="true" />
            </Button>
          ) : null}
        </div>
      </header>

      {state.status === "loading" ? (
        <div className="flex flex-col gap-2" aria-busy="true">
          {Array.from({ length: 4 }, (_, index) => <Skeleton key={index} className="h-8 w-full" />)}
        </div>
      ) : state.status === "error" ? (
        <Alert variant="destructive">
          <AlertDescription>{t("panel.loadFailed")}</AlertDescription>
        </Alert>
      ) : state.sampled ? (
        <TraceWaterfall spans={state.spans} compact />
      ) : (
        <EmptyState
          icon={EyeOff}
          title={t("panel.notSampledTitle")}
          description={capturePolicy ? t("panel.notSampledPolicy", { policy: policyT(`${POLICY_KEYS[capturePolicy]}.title`) }) : t("panel.policyUnavailable")}
          density="compact"
        />
      )}

      {state.status === "ready" && state.attempts.length > 0 ? <AttemptList attempts={state.attempts} /> : null}
    </section>
  );
}

function AttemptList({ attempts }: { attempts: TraceAttempt[] }) {
  const t = useTranslations("traces");
  const locale = useLocale();
  return (
    <div className="flex flex-col gap-2">
      <h3 className="text-xs font-medium text-muted-foreground">{t("panel.attemptsTitle")}</h3>
      <ol className="flex flex-col gap-1.5">
        {attempts.map((attempt) => (
          <li key={attempt.attempt_no} className="flex flex-col gap-1 rounded-md border border-border px-3 py-2 text-xs">
            <div className="flex flex-wrap items-center gap-2">
              <span className="font-medium">{t("detail.attempts.heading", { number: new Intl.NumberFormat(locale).format(attempt.attempt_no) })}</span>
              <Badge variant={textStatusBadgeVariant(attempt.status)}>{attempt.status}</Badge>
              {attempt.provider ? <span dir="auto" className="text-muted-foreground">{attempt.provider}</span> : null}
              {attempt.model ? <span dir="auto" className="text-muted-foreground">{attempt.model}</span> : null}
              <span className="ms-auto text-muted-foreground">{formatDuration(attempt.duration_ms, t("list.units.ms"), t("list.units.s"), locale)}</span>
            </div>
            {attempt.retry_reason ? <p className="text-muted-foreground">{t("panel.retry", { reason: attempt.retry_reason })}</p> : null}
            {attempt.fallback_reason ? <p className="text-muted-foreground">{t("panel.fallback", { reason: attempt.fallback_reason })}</p> : null}
          </li>
        ))}
      </ol>
    </div>
  );
}
