import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const response = await proxyToControlPlane({
    request,
    method: "GET",
    path: "/api/v1/me/permissions",
    userID: auth.user.id,
  });
  response.headers.set("Cache-Control", "private, max-age=30, stale-while-revalidate=90");
  response.headers.set("Vary", "Cookie, Host, x-forwarded-host");
  return response;
}
