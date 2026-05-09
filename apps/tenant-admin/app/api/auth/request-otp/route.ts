import { NextRequest, NextResponse } from "next/server";

import { getAuthRepository } from "@/lib/auth/factory";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

type RequestOTPBody = {
  identifier?: unknown;
  captcha_id?: unknown;
  captcha_answer?: unknown;
};

export async function POST(request: NextRequest) {
  const body = (await request.json().catch(() => null)) as RequestOTPBody | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const identifierRaw =
    typeof body.identifier === "string" ? body.identifier : "";
  const captchaID = typeof body.captcha_id === "string" ? body.captcha_id : "";
  const captchaAnswer =
    typeof body.captcha_answer === "string" ? body.captcha_answer : "";

  const normalized = getAuthRepository().normalizeIdentifier(identifierRaw);
  if (!normalized) {
    return NextResponse.json(
      { error: { message: "ایمیل یا شماره همراه معتبر وارد کنید." } },
      { status: 400 },
    );
  }

  if (!captchaID || !captchaAnswer) {
    return NextResponse.json(
      { error: { message: "ارسال captcha_id و captcha_answer الزامی است." } },
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

  try {
    const repo = getAuthRepository();

    // Validate captcha first
    const captchaResult = await repo.consumeCaptcha(captchaID, captchaAnswer);
    if (!captchaResult.ok) {
      return NextResponse.json(
        { error: { message: captchaResult.error } },
        { status: 400 },
      );
    }

    // Create OTP directly after captcha validation
    const otp = await repo.createOTPRequest({
      tenantSlug,
      purpose: "login",
      channel: normalized.channel,
      identifier: normalized.normalized,
    });

    return NextResponse.json({
      request_id: otp.request_id,
      channel: otp.channel,
      destination_masked: otp.destination_masked,
      expires_in_seconds: otp.expires_in_seconds,
      debug_otp: otp.debug_otp,
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : "Authentication failed.";
    const statusMatch =
      error instanceof Error ? error.message.match(/Auth API error:\s+(\d{3})/) : null;
    const status = statusMatch ? Number(statusMatch[1]) : 500;
    return NextResponse.json({ error: { message } }, { status });
  }
}
