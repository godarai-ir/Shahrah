import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function PATCH(
  request: NextRequest,
  context: { params: Promise<{ userID: string }> },
) {
  const { userID } = await context.params;
  const normalizedUserID = userID.trim();
  if (!normalizedUserID) {
    return NextResponse.json(
      { error: { message: "ارسال userID الزامی است." } },
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
    path: `/api/v1/users/${encodeURIComponent(normalizedUserID)}/profile`,
    body,
  });
}
