import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

type Params = {
  params: Promise<{ logID: string }>;
};

export async function POST(request: NextRequest, { params }: Params) {
  const { logID } = await params;
  const normalized = logID.trim();
  if (!normalized) {
    return Response.json({ error: { message: "logID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "POST",
    path: `/api/v2/logs/${encodeURIComponent(normalized)}/replay-draft`,
  });
}
