import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type Context = { params: Promise<{ virtualUserID: string }> };

export async function GET(request: NextRequest, context: Context) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { virtualUserID } = await context.params;
  const id = virtualUserID?.trim() ?? "";
  if (!id) {
    return NextResponse.json({ error: { message: "virtualUserID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/virtual-users/${encodeURIComponent(id)}`,
    userID: auth.user.id,
  });
}

export async function PATCH(request: NextRequest, context: Context) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { virtualUserID } = await context.params;
  const id = virtualUserID?.trim() ?? "";
  if (!id) {
    return NextResponse.json({ error: { message: "virtualUserID is required." } }, { status: 400 });
  }
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object") {
    return NextResponse.json({ error: { message: "Invalid request body." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "PATCH",
    path: `/api/v1/virtual-users/${encodeURIComponent(id)}`,
    userID: auth.user.id,
    body,
  });
}

export async function DELETE(request: NextRequest, context: Context) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const { virtualUserID } = await context.params;
  const id = virtualUserID?.trim() ?? "";
  if (!id) {
    return NextResponse.json({ error: { message: "virtualUserID is required." } }, { status: 400 });
  }
  return proxyToControlPlane({
    request,
    method: "DELETE",
    path: `/api/v1/virtual-users/${encodeURIComponent(id)}`,
    userID: auth.user.id,
  });
}
