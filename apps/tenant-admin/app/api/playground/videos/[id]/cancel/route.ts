import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const { id } = await context.params;
  return proxyToControlPlane({ request, method: "POST", path: `/v1/videos/${encodeURIComponent(id)}/cancel`, body: {} });
}
