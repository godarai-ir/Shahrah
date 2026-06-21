import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export async function GET(request: NextRequest) {
  return proxyToControlPlane({
    request,
    method: "GET",
    path: "/api/v1/metrics/facets",
  });
}
