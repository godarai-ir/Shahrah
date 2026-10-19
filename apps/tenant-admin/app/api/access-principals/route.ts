import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {

  const url = new URL(request.url);
  const kind = (url.searchParams.get("kind") ?? "").trim().toLowerCase();
  const q = (url.searchParams.get("q") ?? "").trim();
  const limit = (url.searchParams.get("limit") ?? "").trim();

  if (kind !== "user" && kind !== "team") {
    return NextResponse.json(
      { error: { message: "kind must be user or team." } },
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
