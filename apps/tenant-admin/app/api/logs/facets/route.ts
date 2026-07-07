import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export async function GET(request: NextRequest) {
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
    "duration_min",
    "duration_max",
    "tokens_min",
    "tokens_max",
    "cost_usd_min",
    "cost_usd_max",
    "cost_irr_min",
    "cost_irr_max",
    "from",
    "to",
    "sort_by",
    "sort_order",
  ]) {
    const value = url.searchParams.get(key)?.trim();
    if (value) {
      params.set(key, value);
    }
  }

  const path = params.size > 0 ? `/api/v2/logs/facets?${params.toString()}` : "/api/v2/logs/facets";
  return proxyToControlPlane({ request, method: "GET", path });
}
