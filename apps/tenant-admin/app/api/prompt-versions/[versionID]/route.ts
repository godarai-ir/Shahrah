import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ versionID: string }> };

export async function GET(request: NextRequest, context: Params) {
  const { versionID } = await context.params;
  if (!versionID?.trim()) {
    return NextResponse.json({ error: { message: "versionID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/prompt-versions/${encodeURIComponent(versionID)}`,
  });
}
