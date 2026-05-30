import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export async function GET(request: NextRequest) {
  return proxyToControlPlane({
    request,
    method: "GET",
    path: "/api/v1/settings/trace-privacy",
  });
}

export async function PUT(request: NextRequest) {
  const payload = await request.json().catch(() => null);
  if (!payload || typeof payload !== "object") {
    return Response.json({ error: { message: "Invalid request body." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "PUT",
    path: "/api/v1/settings/trace-privacy",
    body: payload,
  });
}


