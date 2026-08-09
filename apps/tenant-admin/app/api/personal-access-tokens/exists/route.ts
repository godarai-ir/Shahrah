import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const name = request.nextUrl.searchParams.get("name")?.trim() ?? "";
  if (!name) {
    return NextResponse.json({ error: { message: "The name field is required." } }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/personal-access-tokens/exists?name=${encodeURIComponent(name)}`,
    userID: auth.user.id,
  });
}
