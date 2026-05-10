import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ id: string }> };

export async function GET(request: NextRequest, context: Params) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { id } = await context.params;
  if (!id?.trim()) {
    return NextResponse.json({ error: { message: "id is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`,
  });
}

export async function PUT(request: NextRequest, context: Params) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { id } = await context.params;
  if (!id?.trim()) {
    return NextResponse.json({ error: { message: "id is required." } }, { status: 400 });
  }
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "Invalid payload." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "PUT",
    path: `/api/v1/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`,
    body,
    userID: auth.user.id,
  });
}

export async function DELETE(request: NextRequest, context: Params) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { id } = await context.params;
  if (!id?.trim()) {
    return NextResponse.json({ error: { message: "id is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "DELETE",
    path: `/api/v1/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`,
    userID: auth.user.id,
  });
}
