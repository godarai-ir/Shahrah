import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ slug: string }>;
};

export async function GET(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { slug } = await context.params;
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/guardrails/${encodeURIComponent(slug)}`,
  });
}

export async function PUT(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { slug } = await context.params;
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "Invalid payload." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "PUT",
    path: `/api/v1/guardrails/${encodeURIComponent(slug)}`,
    body,
    userID: auth.user.id,
  });
}

export async function DELETE(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { slug } = await context.params;
  return proxyToControlPlane({
    request,
    method: "DELETE",
    path: `/api/v1/guardrails/${encodeURIComponent(slug)}`,
    userID: auth.user.id,
  });
}
