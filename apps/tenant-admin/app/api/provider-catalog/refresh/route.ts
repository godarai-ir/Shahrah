import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/provider-catalog/refresh",
  });
}
