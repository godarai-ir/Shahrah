import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "Invalid payload." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/guardrails/semantic/test",
    body,
    userID: auth.user.id,
  });
}
