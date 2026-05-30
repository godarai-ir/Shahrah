import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export async function GET(request: NextRequest) {
  const url = new URL(request.url);
  const params = new URLSearchParams();

  for (const key of ["trace_id", "request_id", "route", "model", "provider", "status_code", "error_code", "from", "to", "sort_by", "sort_order"]) {
    const value = url.searchParams.get(key)?.trim();
    if (value) {
      params.set(key, value);
    }
  }

  const path =
    params.toString().length > 0
      ? `/api/v2/traces/facets?${params.toString()}`
      : "/api/v2/traces/facets";

  return proxyToControlPlane({
    request,
    method: "GET",
    path,
  });
}


