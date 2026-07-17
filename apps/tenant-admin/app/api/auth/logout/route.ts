import { NextRequest, NextResponse } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { getAuthService } from "@/lib/auth/factory";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  let revocationError: unknown = null;

  if (sessionID) {
    try {
      await getAuthService().revokeSession(sessionID);
    } catch (error) {
      revocationError = error;
      console.error("[tenant-admin][auth] session revocation failed during logout", error);
    }
  }

  const response = NextResponse.json(
    revocationError
      ? { success: false, error: { message: "Session revocation failed." } }
      : { success: true },
    { status: revocationError ? 502 : 200 },
  );
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
