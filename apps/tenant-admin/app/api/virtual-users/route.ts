import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const params = request.nextUrl.searchParams;
  const search = new URLSearchParams();
  const query = params.get("query")?.trim() ?? "";
  const limit = params.get("limit")?.trim() ?? "";
  const offset = params.get("offset")?.trim() ?? "";
  if (query) search.set("query", query);
  if (limit) search.set("limit", limit);
  if (offset) search.set("offset", offset);

  const path = search.size > 0 ? `/api/v1/virtual-users?${search.toString()}` : "/api/v1/virtual-users";
  return proxyToControlPlane({ request, method: "GET", path, userID: auth.user.id });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "Invalid request body." } }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/virtual-users",
    userID: auth.user.id,
    body,
  });
}
