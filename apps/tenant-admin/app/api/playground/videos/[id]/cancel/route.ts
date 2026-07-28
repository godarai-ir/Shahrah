import { NextRequest } from "next/server";

import {
  authorizeAPIRequest,
  PLAYGROUND_VIDEO_MUTATE_PERMISSIONS,
} from "@/lib/auth/api-permissions";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const authorization = await authorizeAPIRequest(request, PLAYGROUND_VIDEO_MUTATE_PERMISSIONS);
  if (!authorization.ok) return authorization.response;

  const { id } = await context.params;
  return proxyToControlPlane({ request, method: "POST", path: `/v1/videos/${encodeURIComponent(id)}/cancel`, body: {} });
}
