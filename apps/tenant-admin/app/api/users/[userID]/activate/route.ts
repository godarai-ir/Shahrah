import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function POST(
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

  return proxyToControlPlane({
    request,
    method: "POST",
    path: `/api/v1/users/${encodeURIComponent(normalizedUserID)}/activate`,
  });
}
