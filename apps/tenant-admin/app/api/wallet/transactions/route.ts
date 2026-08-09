import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {

  const url = new URL(request.url);
  const limit = url.searchParams.get("limit") ?? "20";
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/wallet/transactions?limit=${encodeURIComponent(limit)}`,
  });
}
