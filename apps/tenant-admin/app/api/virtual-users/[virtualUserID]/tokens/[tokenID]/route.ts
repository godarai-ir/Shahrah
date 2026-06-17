import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type Context = { params: Promise<{ virtualUserID: string; tokenID: string }> };

export async function PATCH(request: NextRequest, context: Context) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { virtualUserID, tokenID } = await context.params;
  const virtualUser = virtualUserID?.trim() ?? "";
  const token = tokenID?.trim() ?? "";
  if (!virtualUser || !token) {
    return NextResponse.json({ error: { message: "virtualUserID and tokenID are required." } }, { status: 400 });
  }
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "Invalid request body." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "PATCH",
    path: `/api/v1/virtual-users/${encodeURIComponent(virtualUser)}/tokens/${encodeURIComponent(token)}`,
    userID: auth.user.id,
    body,
  });
}
