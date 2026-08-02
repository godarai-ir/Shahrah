import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane, proxyToControlPlaneBodyStream } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

function pathWithQuery(request: NextRequest) {
  const query = request.nextUrl.searchParams.toString();
  return `/api/v1/wallet/bank-transfers${query ? `?${query}` : ""}`;
}

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: pathWithQuery(request),
  });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const contentType = request.headers.get("content-type") ?? "";
  if (contentType.includes("multipart/form-data")) {
    return proxyToControlPlaneBodyStream({
      request,
      method: "POST",
      path: "/api/v1/wallet/bank-transfers",
    });
  }

  const body = await readJSONBankTransferBody(request);
  if (!body.ok) {
    return NextResponse.json({ error: { message: body.message } }, { status: 400 });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/wallet/bank-transfers",
    body: body.value,
  });
}

async function readJSONBankTransferBody(request: NextRequest) {
  const body = await request.json().catch(() => undefined);
  if (body && typeof body === "object" && "receipt_file_url" in body) {
    return {
      ok: false as const,
      message: "receipt_file_url must be created by uploading a receipt image.",
    };
  }

  return { ok: true as const, value: body };
}
