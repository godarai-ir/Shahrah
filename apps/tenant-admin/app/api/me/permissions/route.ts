import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {

  const response = await proxyToControlPlane({
    request,
    method: "GET",
    path: "/api/v1/me/permissions",
  });
  response.headers.set("Cache-Control", "private, max-age=30, stale-while-revalidate=90");
  response.headers.set("Vary", "Cookie, Host, x-forwarded-host");
  return response;
}
