import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

type Params = {
  params: Promise<{ viewID: string }>;
};

export async function PATCH(request: NextRequest, { params }: Params) {
  const { viewID } = await params;
  const normalized = viewID.trim();
  if (!normalized) {
    return Response.json({ error: { message: "viewID is required." } }, { status: 400 });
  }
  const body = await request.json().catch(() => ({}));
  return proxyToControlPlane({
    request,
    method: "PATCH",
    path: `/api/v2/logs/saved-views/${encodeURIComponent(normalized)}`,
    body,
  });
}

export async function DELETE(request: NextRequest, { params }: Params) {
  const { viewID } = await params;
  const normalized = viewID.trim();
  if (!normalized) {
    return Response.json({ error: { message: "viewID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "DELETE",
    path: `/api/v2/logs/saved-views/${encodeURIComponent(normalized)}`,
  });
}
