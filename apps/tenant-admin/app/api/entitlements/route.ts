import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {

  return proxyToControlPlane({
    request,
    method: "GET",
    path: "/api/v1/entitlements",
  });
}
