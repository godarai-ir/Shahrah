import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function DELETE(
  request: NextRequest,
  context: { params: Promise<{ id: string }> },
) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const params = await context.params;
  const id = params.id?.trim() ?? "";
  if (!id) {
    return NextResponse.json({ error: { message: "Invalid token identifier." } }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "DELETE",
    path: `/api/v1/personal-access-tokens/${encodeURIComponent(id)}`,
    userID: auth.user.id,
  });
}
