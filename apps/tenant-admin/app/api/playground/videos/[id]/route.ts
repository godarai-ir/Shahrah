import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const { id } = await context.params;
  return proxyToControlPlane({ request, method: "GET", path: `/v1/videos/${encodeURIComponent(id)}` });
}

export async function DELETE(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const { id } = await context.params;
  return proxyToControlPlane({ request, method: "DELETE", path: `/v1/videos/${encodeURIComponent(id)}` });
}
