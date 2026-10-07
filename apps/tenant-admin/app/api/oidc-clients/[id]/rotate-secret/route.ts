import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const id = (await context.params).id?.trim() ?? "";
  if (!id) {
    return NextResponse.json({ error: { code: "oidc_client_not_found", message: "Invalid client identifier." } }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: `/api/v1/oidc-clients/${encodeURIComponent(id)}/rotate-secret`,
    userID: auth.user.id,
  });
}
