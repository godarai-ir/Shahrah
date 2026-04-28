import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type AcceptInvitePayload = {
  token?: unknown;
};

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const body = (await request.json().catch(() => null)) as AcceptInvitePayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const token = typeof body.token === "string" ? body.token.trim() : "";
  if (!token) {
    return NextResponse.json(
      { error: { message: "ارسال token الزامی است." } },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/users/invitations/accept",
    body: {
      token,
    },
  });
}
