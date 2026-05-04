import { NextRequest, NextResponse } from "next/server";

import {
  consumeCaptcha,
  createOTPRequest,
  normalizeIdentifier,
} from "@/lib/auth/store";
import { prisma } from "@/lib/prisma";
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

  const normalized = normalizeIdentifier(identifierRaw);
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

  const captchaResult = await consumeCaptcha(captchaID, captchaAnswer);
  if (!captchaResult.ok) {
    return NextResponse.json(
      { error: { message: captchaResult.error } },
      { status: 400 },
    );
  }

  const tenantSlug = resolveTenantSlugFromHeaders(
    request.headers,
  );
  if (!tenantSlug) {
    return NextResponse.json(
      { error: { message: "Tenant host is required." } },
      { status: 400 },
    );
  }
  const tenantRows = await prisma.$queryRaw<Array<{ tenant_id: string }>>`
    SELECT tenant_id
    FROM tenants
    WHERE slug = ${tenantSlug} AND status = 'active'
    LIMIT 1
  `;
  if (!tenantRows[0]) {
    return NextResponse.json(
      { error: { message: "Tenant not found." } },
      { status: 404 },
    );
  }
  const otp = await createOTPRequest({
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
}
