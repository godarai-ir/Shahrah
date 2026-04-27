import { NextRequest, NextResponse } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { revokeTenantSession } from "@/lib/auth/store";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (sessionID) {
    await revokeTenantSession(sessionID);
  }

  const response = NextResponse.redirect(new URL("/login", request.url));
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
