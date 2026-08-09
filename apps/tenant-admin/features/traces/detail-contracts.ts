import { z } from "zod";

const traceSchema = z.object({
  trace_id: z.string(), tenant_id: z.string(), request_id: z.string().optional(), model: z.string().optional(),
  provider: z.string().optional(), route: z.string(), status_code: z.number(), started_at: z.string(),
  duration_ms: z.number(), error_code: z.string().optional(), sampled: z.boolean(), attempts: z.number(),
});
const traceSummarySchema = z.object({ spans_count: z.number(), attempts_count: z.number(), errors_count: z.number() });
const traceUsageSchema = z.object({
  model: z.string().optional(), input_tokens: z.number().optional(), output_tokens: z.number().optional(),
  total_tokens: z.number().optional(), cost_usd: z.number().optional(), cost_irr: z.number().optional(),
  response_time_ms: z.number().optional(), status: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});
const traceMetricSchema = z.object({
  request_type: z.string().optional(), provider_name: z.string().optional(), provider_model_name: z.string().optional(),
  error_type: z.string().optional(), latency_ms: z.number().optional(), input_tokens: z.number().optional(),
  output_tokens: z.number().optional(), cost_usd: z.number().optional(), ttft_ms: z.number().optional(),
  itl_ms: z.number().optional(), tpot_ms: z.number().optional(), virtual_model_name: z.string().optional(),
});
const traceDetailSchema = z.object({
  trace: traceSchema, summary: traceSummarySchema, usage: traceUsageSchema.optional(), metric: traceMetricSchema.optional(),
});
const traceSpanSchema = z.object({
  span_id: z.string(), trace_id: z.string(), parent_span_id: z.string().optional(), name: z.string(), kind: z.string(),
  started_at: z.string(), duration_ms: z.number(), status: z.string(), attributes: z.record(z.string(), z.unknown()),
});
const traceAttemptSchema = z.object({
  attempt_no: z.number(), trace_id: z.string(), provider: z.string().optional(), model: z.string().optional(),
  started_at: z.string(), duration_ms: z.number(), status: z.string(), retry_reason: z.string().optional(),
  fallback_reason: z.string().optional(), attributes: z.record(z.string(), z.unknown()).optional(),
});
const spansResponseSchema = z.object({ data: z.array(traceSpanSchema).optional() });
const attemptsResponseSchema = z.object({ data: z.array(traceAttemptSchema).optional() });

export type TraceDetailV2Response = z.infer<typeof traceDetailSchema>;
export type RequestTraceV2 = z.infer<typeof traceSchema>;
export type TraceSpan = z.infer<typeof traceSpanSchema>;
export type TraceAttempt = z.infer<typeof traceAttemptSchema>;

export function parseTraceDetail(payload: unknown): TraceDetailV2Response | null {
  const parsed = traceDetailSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

export function parseTraceSpans(payload: unknown): TraceSpan[] {
  const parsed = spansResponseSchema.safeParse(payload);
  return parsed.success ? parsed.data.data ?? [] : [];
}

export function parseTraceAttempts(payload: unknown): TraceAttempt[] {
  const parsed = attemptsResponseSchema.safeParse(payload);
  return parsed.success ? parsed.data.data ?? [] : [];
}
