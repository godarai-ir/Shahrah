import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

function pathWithQuery(request: NextRequest) {
  const query = request.nextUrl.searchParams.toString();
  return `/api/v1/wallet/bank-transfers${query ? `?${query}` : ""}`;
}

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: pathWithQuery(request),
  });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/wallet/bank-transfers",
    body: await request.json().catch(() => undefined),
  });
}
