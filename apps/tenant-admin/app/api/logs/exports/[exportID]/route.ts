import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

type Params = {
  params: Promise<{ exportID: string }>;
};

export async function GET(request: NextRequest, { params }: Params) {
  const { exportID } = await params;
  const normalized = exportID.trim();
  if (!normalized) {
    return Response.json({ error: { message: "exportID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v2/logs/exports/${encodeURIComponent(normalized)}`,
  });
}
