import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { updateTenantUserContact, verifyOTPRequest } from "@/lib/auth/store";

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

  const otpResult = await verifyOTPRequest({
    tenantSlug: auth.tenantSlug,
    requestID,
    otpCode,
    purpose: field === "email" ? "change_email" : "change_mobile",
    userID: auth.user.id,
  });
  if (!otpResult.ok) {
    return NextResponse.json(
      { error: { message: otpResult.error } },
      { status: 400 },
    );
  }

  const pendingValue = otpResult.record.pendingValue ?? otpResult.record.identifier;
  const updateResult = await updateTenantUserContact({
    tenantSlug: auth.tenantSlug,
    userID: auth.user.id,
    field,
    value: pendingValue,
  });
  if (!updateResult.ok) {
    return NextResponse.json(
      { error: { message: updateResult.error } },
      { status: 400 },
    );
  }

  return NextResponse.json({
    user: {
      id: updateResult.user.id,
      email: updateResult.user.email ?? null,
      mobile: updateResult.user.mobile ?? null,
      role: updateResult.user.role,
    },
  });
}
