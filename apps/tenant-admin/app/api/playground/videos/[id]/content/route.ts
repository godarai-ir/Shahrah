import { NextRequest } from "next/server";

import { proxyToControlPlaneAsset } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest, context: { params: Promise<{ id: string }> }) {
  const { id } = await context.params;
  const artifactID = request.nextUrl.searchParams.get("artifact_id");
  const query = artifactID ? `?artifact_id=${encodeURIComponent(artifactID)}` : "";
  return proxyToControlPlaneAsset({
    request,
    method: "GET",
    path: `/v1/videos/${encodeURIComponent(id)}/content${query}`,
  });
}
