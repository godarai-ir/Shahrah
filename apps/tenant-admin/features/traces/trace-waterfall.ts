import type { TraceSpan } from "./detail-contracts";

export type BadgeVariant = "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "info";

export type WaterfallRow = {
  span: TraceSpan;
  /** Milliseconds from the earliest span start. */
  offsetMS: number;
  leftPct: number;
  widthPct: number;
  /** Span of the whole trace, identical on every row. */
  totalMS: number;
};

/** Minimum bar width so zero-length spans stay visible. */
const MIN_WIDTH_PCT = 1.5;

export function sortSpans(spans: TraceSpan[]): TraceSpan[] {
  return [...spans].sort((a, b) => spanStartMS(a) - spanStartMS(b));
}

/** Lays spans out on one time axis. Spans with an unreadable start are placed at the origin. */
export function buildWaterfallRows(spans: TraceSpan[]): WaterfallRow[] {
  const starts = spans.map(spanStartMS).filter(Number.isFinite);
  if (starts.length === 0) return [];
  const startMS = Math.min(...starts);
  const endMS = Math.max(...spans.map((span) => (Number.isFinite(spanStartMS(span)) ? spanStartMS(span) : startMS) + spanDurationMS(span)));
  const totalMS = Math.max(1, endMS - startMS);
  return spans.map((span) => {
    const spanStart = spanStartMS(span);
    const offsetMS = Number.isFinite(spanStart) ? Math.max(0, spanStart - startMS) : 0;
    // Keep room for the minimum bar so a span at the end of the window stays visible.
    const leftPct = Math.min(100 - MIN_WIDTH_PCT, (offsetMS / totalMS) * 100);
    const widthPct = Math.min(100 - leftPct, Math.max(MIN_WIDTH_PCT, (Math.max(spanDurationMS(span), 1) / totalMS) * 100));
    return { span, offsetMS, leftPct, widthPct, totalMS };
  });
}

export function textStatusBadgeVariant(status: string): BadgeVariant {
  const normalized = status.trim().toLowerCase();
  if (["error", "failed", "failure"].includes(normalized)) return "destructive";
  if (["retry", "fallback", "timeout"].includes(normalized)) return "warning";
  if (["ok", "success", "succeeded", "complete", "completed"].includes(normalized)) return "success";
  return "secondary";
}

export function spanKindBadgeVariant(kind: string): BadgeVariant {
  const normalized = kind.trim().toLowerCase();
  if (normalized === "server") return "info";
  if (normalized === "client") return "warning";
  if (normalized === "internal") return "secondary";
  return "outline";
}

export function formatSpanDuration(value: number, msLabel: string, locale: string) {
  if (!Number.isFinite(value)) return `0 ${msLabel}`;
  return `${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)} ${msLabel}`;
}

function spanStartMS(span: TraceSpan) {
  return new Date(span.started_at).getTime();
}

function spanDurationMS(span: TraceSpan) {
  return Number.isFinite(span.duration_ms) ? Math.max(0, span.duration_ms) : 0;
}
