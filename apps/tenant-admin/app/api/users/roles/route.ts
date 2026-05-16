import { NextRequest, NextResponse } from "next/server";

import { invalidateAllMePermissionsCache, invalidateMePermissionsCache } from "@/lib/auth/me-permissions";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function PATCH(request: NextRequest) {
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
  const payload = body as { email?: string };
  const actorEmail = (auth.user.email ?? "").trim().toLowerCase();
  const targetEmail = (payload.email ?? "").trim().toLowerCase();
  if (actorEmail && targetEmail && actorEmail === targetEmail) {
    return NextResponse.json(
      { error: { message: "کاربر جاری مجاز به تغییر نقش خودش نیست." } },
      { status: 403 },
    );
  }

  const response = await proxyToControlPlane({
    request,
    method: "PATCH",
    path: "/api/v1/users/roles",
    body,
  });
  if (response.ok) {
    invalidateAllMePermissionsCache();
    invalidateMePermissionsCache(auth.tenantSlug, auth.user.id);
  }
  return response;
}
