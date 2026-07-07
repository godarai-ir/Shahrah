import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

type RangeLog = {
  duration_ms?: number;
  total_tokens?: number;
  cost_usd?: number;
  cost_irr?: number;
};

export async function GET(request: NextRequest) {
  const path = buildLogsPath(request, "/api/v2/logs/range-stats");
  const response = await proxyToControlPlane({ request, method: "GET", path });
  if (!shouldFallbackToList(response)) {
    response.headers.set("Cache-Control", "private, max-age=60");
    return response;
  }

  const fallbackPath = appendQueryParam(buildLogsPath(request, "/api/v2/logs"), "limit", "200");
  const fallbackResponse = await proxyToControlPlane({ request, method: "GET", path: appendQueryParam(fallbackPath, "offset", "0") });
  if (!fallbackResponse.ok) {
    response.headers.set("Cache-Control", "private, max-age=15");
    return response;
  }

  const payload = (await fallbackResponse.json().catch(() => ({}))) as { data?: RangeLog[] };
  const stats = rangeStatsFromLogs(Array.isArray(payload?.data) ? payload.data : []);
  const out = NextResponse.json({ data: stats });
  out.headers.set("Cache-Control", "private, max-age=60");
  return out;
}

function shouldFallbackToList(response: Response) {
  return response.status === 404;
}

function appendQueryParam(path: string, key: string, value: string) {
  const separator = path.includes("?") ? "&" : "?";
  return `${path}${separator}${encodeURIComponent(key)}=${encodeURIComponent(value)}`;
}

function rangeStatsFromLogs(logs: RangeLog[]) {
  return {
    duration_ms: rangeFromValues(logs.map((item) => item.duration_ms)),
    total_tokens: rangeFromValues(logs.map((item) => item.total_tokens)),
    cost_usd: rangeFromValues(logs.map((item) => item.cost_usd)),
    cost_irr: rangeFromValues(logs.map((item) => item.cost_irr)),
  };
}

function rangeFromValues(values: Array<number | undefined>) {
  const numeric = values.filter((value): value is number => typeof value === "number" && Number.isFinite(value));
  if (!numeric.length) return { min: 0, max: 0 };
  return {
    min: Math.min(...numeric),
    max: Math.max(...numeric),
  };
}

function buildLogsPath(request: NextRequest, basePath: string) {
  const url = new URL(request.url);
  const params = new URLSearchParams();
  for (const key of [
    "log_id",
    "request_id",
    "trace_id",
    "provider",
    "model",
    "virtual_model",
    "route",
    "request_type",
    "status_code",
    "error_type",
    "cache_status",
    "fallback_status",
    "load_balance_status",
    "payload_mode",
    "actor_type",
    "actor_id",
    "org_unit_id",
    "team_id",
    "from",
    "to",
  ]) {
    const value = url.searchParams.get(key)?.trim();
    if (value) {
      params.set(key, value);
    }
  }
  return params.size > 0 ? `${basePath}?${params.toString()}` : basePath;
}
