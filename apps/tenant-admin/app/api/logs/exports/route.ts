import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export async function GET(request: NextRequest) {
  const url = new URL(request.url);
  const limit = url.searchParams.get("limit")?.trim();
  const path = limit ? `/api/v2/logs/exports?limit=${encodeURIComponent(limit)}` : "/api/v2/logs/exports";
  return proxyToControlPlane({ request, method: "GET", path });
}

export async function POST(request: NextRequest) {
  const body = await request.json().catch(() => ({}));
  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v2/logs/exports",
    body,
  });
}
