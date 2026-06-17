import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type Context = { params: Promise<{ virtualUserID: string; tokenID: string }> };

export async function POST(request: NextRequest, context: Context) {
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
  const body = await request.json().catch(() => ({}));
  return proxyToControlPlane({
    request,
    method: "POST",
    path: `/api/v1/virtual-users/${encodeURIComponent(virtualUser)}/tokens/${encodeURIComponent(token)}/revoke`,
    userID: auth.user.id,
    body,
  });
}
