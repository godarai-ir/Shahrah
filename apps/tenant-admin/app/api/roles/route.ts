import { NextRequest, NextResponse } from "next/server";

import { invalidateAllMePermissionsCache, invalidateMePermissionsCache } from "@/lib/auth/me-permissions";
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
  const query = new URLSearchParams();

  const rawQuery = (url.searchParams.get("query") ?? "").trim();
  const rawLimit = (url.searchParams.get("limit") ?? "").trim();
  const rawOffset = (url.searchParams.get("offset") ?? "").trim();

  if (rawQuery) query.set("query", rawQuery);
  if (rawLimit) query.set("limit", rawLimit);
  if (rawOffset) query.set("offset", rawOffset);

  const path =
    query.toString().length > 0
      ? `/api/v1/roles?${query.toString()}`
      : "/api/v1/roles";

  return proxyToControlPlane({
    request,
    method: "GET",
    path,
  });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "Invalid request body." } },
      { status: 400 },
    );
  }

  const response = await proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/roles",
    body,
  });
  if (response.ok) {
    invalidateAllMePermissionsCache();
    invalidateMePermissionsCache(auth.tenantSlug, auth.user.id);
  }
  return response;
}
