import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { getAuthService } from "@/lib/auth/factory";

export const dynamic = "force-dynamic";

type VerifyBody = {
  field?: unknown;
  request_id?: unknown;
  otp_code?: unknown;
};

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const body = (await request.json().catch(() => null)) as VerifyBody | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const field = body.field === "email" || body.field === "mobile" ? body.field : null;
  const requestID = typeof body.request_id === "string" ? body.request_id : "";
  const otpCode = typeof body.otp_code === "string" ? body.otp_code : "";
  if (!field || !requestID || !otpCode) {
    return NextResponse.json(
      { error: { message: "ارسال field، request_id و otp_code الزامی است." } },
      { status: 400 },
    );
  }

  const authService = getAuthService();

  try {
    const updatedUser = await authService.verifyContactChangeOTP({
      tenantSlug: auth.tenantSlug,
      userID: auth.user.id,
      requestID,
      otpCode,
      field,
    });

    return NextResponse.json({
      user: {
        id: updatedUser.id,
        email: updatedUser.email ?? null,
        mobile: updatedUser.mobile ?? null,
        role: updatedUser.role,
      },
    });
  } catch (err) {
    const message = err instanceof Error ? err.message : "خطایی رخ داد.";
    return NextResponse.json(
      { error: { message } },
      { status: 400 },
    );
  }
}
