import { NextRequest, NextResponse } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { getAuthService } from "@/lib/auth/factory";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

type VerifyOTPBody = {
  request_id?: unknown;
  otp_code?: unknown;
};

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const body = (await request.json().catch(() => null)) as VerifyOTPBody | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const requestID = typeof body.request_id === "string" ? body.request_id : "";
  const otpCode = typeof body.otp_code === "string" ? body.otp_code : "";
  if (!requestID || !otpCode) {
    return NextResponse.json(
      { error: { message: "ارسال request_id و otp_code الزامی است." } },
      { status: 400 },
    );
  }

  const tenantSlug = resolveTenantSlugFromHeaders(request.headers);
  if (!tenantSlug) {
    return NextResponse.json(
      { error: { message: "Tenant host is required." } },
      { status: 400 },
    );
  }

  let result;
  try {
    const authService = getAuthService();
    result = await authService.verifyOTPForLogin({
      tenantSlug,
      requestID,
      otpCode,
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : "Authentication failed.";
    const statusMatch =
      error instanceof Error ? error.message.match(/Auth API error:\s+(\d{3})/) : null;
    const status = statusMatch ? Number(statusMatch[1]) : 500;
    return NextResponse.json({ error: { message } }, { status });
  }

  if (!result.ok) {
    return NextResponse.json(
      { error: { message: result.error } },
      { status: result.status ?? 400 },
    );
  }

  const response = NextResponse.json({
    user: {
      id: result.user!.id,
      email: result.user!.email ?? null,
      mobile: result.user!.mobile ?? null,
      role: result.user!.role,
      tenant_slug: tenantSlug,
    },
  });

  response.cookies.set({
    name: TENANT_ADMIN_SESSION_COOKIE,
    value: result.session!.id,
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    expires: new Date(result.session!.expiresAtISO),
  });

  return response;
}
