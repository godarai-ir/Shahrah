import { NextRequest, NextResponse } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { revokeTenantSession } from "@/lib/auth/store";
import { resolveRequestOriginFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (sessionID) {
    await revokeTenantSession(sessionID);
  }

  const origin = resolveRequestOriginFromHeaders(request.headers, request.url);
  const response = NextResponse.redirect(new URL("/", origin));
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
