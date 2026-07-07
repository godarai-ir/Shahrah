import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export async function GET(request: NextRequest) {
  return proxyToControlPlane({ request, method: "GET", path: "/api/v2/logs/saved-views" });
}

export async function POST(request: NextRequest) {
  const body = await request.json().catch(() => ({}));
  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v2/logs/saved-views",
    body,
  });
}
