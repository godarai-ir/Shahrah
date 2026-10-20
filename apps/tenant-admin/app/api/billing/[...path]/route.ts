import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ path?: string[] }>;
};

function buildPath(parts: string[] | undefined, request: NextRequest) {
  const suffix = (parts ?? []).map(encodeURIComponent).join("/");
  const query = request.nextUrl.searchParams.toString();
  return `/api/v1/billing/${suffix}${query ? `?${query}` : ""}`;
}

async function proxyBilling(request: NextRequest, method: "GET" | "POST" | "PUT" | "DELETE", context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const { path } = await context.params;
  const body = method === "GET" || method === "DELETE" ? undefined : await request.json().catch(() => undefined);
  return proxyToControlPlane({
    request,
    method,
    path: buildPath(path, request),
    body,
  });
}

export function GET(request: NextRequest, context: RouteContext) {
  return proxyBilling(request, "GET", context);
}

export function POST(request: NextRequest, context: RouteContext) {
  return proxyBilling(request, "POST", context);
}

export function PUT(request: NextRequest, context: RouteContext) {
  return proxyBilling(request, "PUT", context);
}

export function DELETE(request: NextRequest, context: RouteContext) {
  return proxyBilling(request, "DELETE", context);
}
