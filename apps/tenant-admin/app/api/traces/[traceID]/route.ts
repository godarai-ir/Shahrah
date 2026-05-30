import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

type Params = {
  params: Promise<{ traceID: string }>;
};

export async function GET(request: NextRequest, { params }: Params) {
  const { traceID } = await params;
  const normalized = traceID.trim();
  if (!normalized) {
    return Response.json({ error: { message: "traceID is required." } }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v2/traces/${encodeURIComponent(normalized)}`,
  });
}


