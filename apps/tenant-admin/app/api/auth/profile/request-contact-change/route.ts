import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { getAuthService } from "@/lib/auth/factory";

export const dynamic = "force-dynamic";

type RequestBody = {
  field?: unknown;
  new_value?: unknown;
};

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const body = (await request.json().catch(() => null)) as RequestBody | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const field = body.field === "email" || body.field === "mobile" ? body.field : null;
  const newValueRaw = typeof body.new_value === "string" ? body.new_value : "";
  if (!field || !newValueRaw.trim()) {
    return NextResponse.json(
      { error: { message: "ارسال field و new_value الزامی است." } },
      { status: 400 },
    );
  }

  const authService = getAuthService();
  const normalized = authService.normalizeIdentifier(newValueRaw);
  if (!normalized) {
    return NextResponse.json(
      { error: { message: "مقدار جدید باید ایمیل یا شماره همراه معتبر باشد." } },
      { status: 400 },
    );
  }

  if (normalized.channel !== field) {
    return NextResponse.json(
      {
        error: {
          message:
            field === "email"
              ? "فرمت مقدار جدید با ایمیل سازگار نیست."
              : "فرمت مقدار جدید با شماره همراه سازگار نیست.",
        },
      },
      { status: 400 },
    );
  }

  if (field === "email" && normalized.normalized === auth.user.email) {
    return NextResponse.json(
      { error: { message: "این ایمیل از قبل ثبت شده است." } },
      { status: 400 },
    );
  }

  if (field === "mobile" && normalized.normalized === auth.user.mobile) {
    return NextResponse.json(
      { error: { message: "این شماره همراه از قبل ثبت شده است." } },
      { status: 400 },
    );
  }

  const otp = await authService.requestContactChangeOTP({
    tenantSlug: auth.tenantSlug,
    userID: auth.user.id,
    field,
    identifier: newValueRaw,
  });

  return NextResponse.json({
    request_id: otp.request_id,
    channel: otp.channel,
    destination_masked: otp.destination_masked,
    expires_in_seconds: otp.expires_in_seconds,
    debug_otp: otp.debug_otp,
  });
}
