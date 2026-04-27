import { NextRequest, NextResponse } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import {
  createTenantSession,
  findOrCreateTenantUser,
  touchLastLogin,
  verifyOTPRequest,
} from "@/lib/auth/store";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

type VerifyOTPBody = {
  request_id?: unknown;
  otp_code?: unknown;
};

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

  const tenantSlug = resolveTenantSlugFromHeaders(
    request.headers,
    process.env.SHAHRAH_DEFAULT_TENANT ?? "default",
  );
  const otpResult = await verifyOTPRequest({
    tenantSlug,
    requestID,
    otpCode,
    purpose: "login",
  });
  if (!otpResult.ok) {
    return NextResponse.json(
      { error: { message: otpResult.error } },
      { status: 400 },
    );
  }

  const user = await findOrCreateTenantUser({
    tenantSlug,
    channel: otpResult.record.channel,
    identifier: otpResult.record.identifier,
  });
  await touchLastLogin(tenantSlug, user.id);
  const session = await createTenantSession({
    tenantSlug,
    userID: user.id,
  });

  const response = NextResponse.json({
    user: {
      id: user.id,
      email: user.email ?? null,
      mobile: user.mobile ?? null,
      role: user.role,
      tenant_slug: tenantSlug,
    },
  });

  response.cookies.set({
    name: TENANT_ADMIN_SESSION_COOKIE,
    value: session.id,
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    expires: new Date(session.expiresAtISO),
  });

  return response;
}
