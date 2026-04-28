import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const url = new URL(request.url);
  const limit = url.searchParams.get("limit") ?? "20";
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/wallet/transactions?limit=${encodeURIComponent(limit)}`,
  });
}
