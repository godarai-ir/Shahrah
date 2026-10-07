import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";
import { parseOIDCClientInput } from "@/features/oidc-clients/input";

export const dynamic = "force-dynamic";

type RouteContext = { params: Promise<{ id: string }> };

async function clientPath(context: RouteContext): Promise<string | null> {
  const id = (await context.params).id?.trim() ?? "";
  return id ? `/api/v1/oidc-clients/${encodeURIComponent(id)}` : null;
}

export async function PATCH(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const path = await clientPath(context);
  if (!path) {
    return NextResponse.json({ error: { code: "oidc_client_not_found", message: "Invalid client identifier." } }, { status: 400 });
  }

  const input = parseOIDCClientInput(await request.json().catch(() => null), { partial: true });
  if (!input.ok) {
    return NextResponse.json({ error: { code: "oidc_client_invalid", message: input.message } }, { status: 400 });
  }

  return proxyToControlPlane({ request, method: "PATCH", path, userID: auth.user.id, body: input.value });
}

export async function DELETE(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const path = await clientPath(context);
  if (!path) {
    return NextResponse.json({ error: { code: "oidc_client_not_found", message: "Invalid client identifier." } }, { status: 400 });
  }

  return proxyToControlPlane({ request, method: "DELETE", path, userID: auth.user.id });
}
