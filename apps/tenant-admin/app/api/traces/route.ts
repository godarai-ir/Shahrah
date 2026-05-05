import { proxyToControlPlane } from "@/lib/control-plane";

export async function GET(request: Request) {
  const url = new URL(request.url);
  const params = new URLSearchParams();

  for (const key of ["trace_id", "route", "model", "status_code", "from", "to", "limit", "offset"]) {
    const value = url.searchParams.get(key)?.trim();
    if (value) {
      params.set(key, value);
    }
  }

  const path =
    params.toString().length > 0
      ? `/api/v1/traces?${params.toString()}`
      : "/api/v1/traces";

  return proxyToControlPlane({
    request,
    method: "GET",
    path,
  });
}

