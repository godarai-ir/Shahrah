import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {

  const url = new URL(request.url);
  const query = new URLSearchParams();

  const rawQuery = (url.searchParams.get("query") ?? "").trim();
  const rawLimit = (url.searchParams.get("limit") ?? "").trim();
  const rawOffset = (url.searchParams.get("offset") ?? "").trim();

  if (rawQuery) {
    query.set("query", rawQuery);
  }
  if (rawLimit) {
    query.set("limit", rawLimit);
  }
  if (rawOffset) {
    query.set("offset", rawOffset);
  }

  const path =
    query.toString().length > 0
      ? `/api/v1/users?${query.toString()}`
      : "/api/v1/users";

  return proxyToControlPlane({
    request,
    method: "GET",
    path,
  });
}
