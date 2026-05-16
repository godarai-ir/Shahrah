import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    roleID: string;
  }>;
};

export async function PATCH(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const { roleID } = await context.params;
  const normalizedRoleID = roleID.trim();
  if (!normalizedRoleID) {
    return NextResponse.json(
      { error: { message: "roleID is required." } },
      { status: 400 },
    );
  }

  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "Invalid request body." } },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "PATCH",
    path: `/api/v1/roles/${encodeURIComponent(normalizedRoleID)}`,
    body,
  });
}
