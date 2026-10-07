import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";
import { parseOIDCClientInput } from "@/features/oidc-clients/input";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return proxyToControlPlane({ request, method: "GET", path: "/api/v1/oidc-clients", userID: auth.user.id });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const input = parseOIDCClientInput(await request.json().catch(() => null), { partial: false });
  if (!input.ok) {
    return NextResponse.json({ error: { code: "oidc_client_invalid", message: input.message } }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/oidc-clients",
    userID: auth.user.id,
    body: input.value,
  });
}
