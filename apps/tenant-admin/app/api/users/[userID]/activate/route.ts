import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ userID: string }> },
) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

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
