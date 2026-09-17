import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

/**
 * Tenant-wide key list backing the "API keys" tab. Unlike /api/personal-access-tokens,
 * which is scoped to the caller, this returns every PAT and VAT in the tenant and so
 * requires team admin or higher on the control plane.
 *
 * Deliberately not /api/keys: that path is a 410 tombstone for the retired legacy
 * key API and stays that way.
 */
export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: "/api/v1/keys",
    userID: auth.user.id,
  });
}
