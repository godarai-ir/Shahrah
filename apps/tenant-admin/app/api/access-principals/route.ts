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

  const url = new URL(request.url);
  const kind = (url.searchParams.get("kind") ?? "").trim().toLowerCase();
  const q = (url.searchParams.get("q") ?? "").trim();
  const limit = (url.searchParams.get("limit") ?? "").trim();

  if (kind !== "user" && kind !== "team" && kind !== "org_unit") {
    return NextResponse.json(
      { error: { message: "kind must be user, team, or org_unit." } },
      { status: 400 },
    );
  }

  const query = new URLSearchParams();
  query.set("kind", kind);
  if (q) {
    query.set("q", q);
  }
  if (limit) {
    query.set("limit", limit);
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/access-principals?${query.toString()}`,
  });
}