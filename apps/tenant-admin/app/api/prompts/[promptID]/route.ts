import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ promptID: string }> };

export async function GET(request: NextRequest, context: Params) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { promptID } = await context.params;
  if (!promptID?.trim()) {
    return NextResponse.json({ error: { message: "promptID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/prompts/${encodeURIComponent(promptID)}`,
  });
}

export async function DELETE(request: NextRequest, context: Params) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { promptID } = await context.params;
  if (!promptID?.trim()) {
    return NextResponse.json({ error: { message: "promptID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "DELETE",
    path: `/api/v1/prompts/${encodeURIComponent(promptID)}`,
    userID: auth.user.id,
  });
}
