import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ id: string }> };

export async function PATCH(request: NextRequest, context: Params) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { id } = await context.params;
  if (!id?.trim()) {
    return NextResponse.json({ error: { message: "id is required." } }, { status: 400 });
  }
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "Invalid payload." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "PATCH",
    path: `/api/v1/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}/status`,
    body,
    userID: auth.user.id,
  });
}
