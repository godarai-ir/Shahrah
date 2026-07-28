import { NextRequest } from "next/server";

import {
  authorizeAPIRequest,
  PLAYGROUND_VIDEO_MUTATE_PERMISSIONS,
  PLAYGROUND_VIDEO_READ_PERMISSIONS,
} from "@/lib/auth/api-permissions";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const authorization = await authorizeAPIRequest(request, PLAYGROUND_VIDEO_READ_PERMISSIONS);
  if (!authorization.ok) return authorization.response;

  const { id } = await context.params;
  return proxyToControlPlane({ request, method: "GET", path: `/v1/videos/${encodeURIComponent(id)}` });
}

export async function DELETE(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const authorization = await authorizeAPIRequest(request, PLAYGROUND_VIDEO_MUTATE_PERMISSIONS);
  if (!authorization.ok) return authorization.response;

  const { id } = await context.params;
  return proxyToControlPlane({ request, method: "DELETE", path: `/v1/videos/${encodeURIComponent(id)}` });
}
