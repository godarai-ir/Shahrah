import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type InvitePayload = {
  email?: unknown;
  mobile?: unknown;
  role?: unknown;
  accept_invite_client_url?: unknown;
};

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const body = (await request.json().catch(() => null)) as InvitePayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const email = typeof body.email === "string" ? body.email.trim() : "";
  const mobile = typeof body.mobile === "string" ? body.mobile.trim() : "";
  const role = typeof body.role === "string" ? body.role.trim() : "tenant_member";
  const acceptInviteClientURL =
    typeof body.accept_invite_client_url === "string"
      ? body.accept_invite_client_url.trim()
      : "";

  if (!email && !mobile) {
    return NextResponse.json(
      { error: { message: "ارسال email یا mobile الزامی است." } },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/users/invite",
    body: {
      email: email || undefined,
      mobile: mobile || undefined,
      role: role || "tenant_member",
      accept_invite_client_url: acceptInviteClientURL || undefined,
    },
  });
}
