import { type NextRequest } from "next/server";

import { proxyToControlPlane, proxyToControlPlaneBodyStream } from "@/lib/control-plane";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

export function GET(request: NextRequest) {
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/support/tickets${request.nextUrl.search}`,
  });
}

export function POST(request: NextRequest) {
  return proxyToControlPlaneBodyStream({
    request,
    method: "POST",
    path: "/api/v1/support/tickets",
  });
}
