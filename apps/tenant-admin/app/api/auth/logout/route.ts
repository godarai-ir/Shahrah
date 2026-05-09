import { NextRequest, NextResponse } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { getAuthService } from "@/lib/auth/factory";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (sessionID) {
    await getAuthService().revokeSession(sessionID);
  }

  const response = NextResponse.json({ success: true });
  response.cookies.set({
    name: TENANT_ADMIN_SESSION_COOKIE,
    value: "",
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });
  return response;
}
