import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type CreatePATPayload = {
  name?: unknown;
  models?: unknown;
  allowed_models?: unknown;
  expires_in?: unknown;
};

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const params = request.nextUrl.searchParams;
  const search = new URLSearchParams();
  const query = params.get("query")?.trim() ?? "";
  const limit = params.get("limit")?.trim() ?? "";
  const offset = params.get("offset")?.trim() ?? "";
  if (query) search.set("query", query);
  if (limit) search.set("limit", limit);
  if (offset) search.set("offset", offset);

  const path = search.size > 0
    ? `/api/v1/personal-access-tokens?${search.toString()}`
    : "/api/v1/personal-access-tokens";

  return proxyToControlPlane({
    request,
    method: "GET",
    path,
    userID: auth.user.id,
  });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const body = (await request.json().catch(() => null)) as CreatePATPayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "???? ??????? ??????? ???." } }, { status: 400 });
  }

  const name = typeof body.name === "string" ? body.name.trim() : "";
  const expiresIn = typeof body.expires_in === "string" ? body.expires_in.trim() : "";

  if (!name) {
    return NextResponse.json({ error: { message: "????? name ?????? ???." } }, { status: 400 });
  }
  if ("models" in body || "allowed_models" in body) {
    return NextResponse.json({
      error: {
        message:
          "models/allowed_models are no longer accepted for PAT creation; access is controlled by Provider Account and Virtual Model access settings.",
      },
    }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/personal-access-tokens",
    userID: auth.user.id,
    body: {
      name,
      expires_in: expiresIn || undefined,
    },
  });
}
