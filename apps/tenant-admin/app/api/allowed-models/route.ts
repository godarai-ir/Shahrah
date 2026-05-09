import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const keyType = request.nextUrl.searchParams.get("key_type")?.trim().toLowerCase() ?? "pat";
  if (keyType !== "pat" && keyType !== "vat") {
    return NextResponse.json({ error: { message: "key_type must be pat or vat." } }, { status: 400 });
  }

  const teamID = request.nextUrl.searchParams.get("team_id")?.trim() ?? "";
  const query = new URLSearchParams();
  query.set("key_type", keyType);
  if (teamID) query.set("team_id", teamID);

  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/models/allowed?${query.toString()}`,
    userID: auth.user.id,
  });
}

