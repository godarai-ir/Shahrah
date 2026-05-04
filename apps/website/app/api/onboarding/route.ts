import { NextResponse } from "next/server";

import { sendOnboardingOTPEmail } from "@/app/lib/email";

type OnboardingAction = "check_slug" | "request_otp" | "verify_otp";

type OnboardingBody = {
  action?: OnboardingAction;
  tenant_slug?: string;
  tenant_domain?: string;
  plan_slug?: string;
  email?: string;
  mobile?: string;
  customer_name?: string;
  request_id?: string;
  otp_code?: string;
  metadata?: Record<string, unknown>;
};

type OnboardingOTPPayload = {
  request_id?: string;
  tenant_slug?: string;
  expires_in_seconds?: number;
  debug_otp?: string;
};

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as OnboardingBody | null;
  if (!body?.action) {
    return NextResponse.json({ ok: false, error: "Invalid onboarding request" }, { status: 400 });
  }

  try {
    const upstream = await callControlPlane(body);
    if (body.action === "request_otp" && upstream.status >= 200 && upstream.status < 300) {
      await sendOnboardingEmailIfPossible(body, upstream.payload);
    }
    return NextResponse.json(upstream.payload, { status: upstream.status });
  } catch (error) {
    return NextResponse.json(
      {
        ok: false,
        error: error instanceof Error ? error.message : "Control-plane onboarding request failed",
      },
      { status: 502 },
    );
  }
}

async function sendOnboardingEmailIfPossible(body: OnboardingBody, payload: unknown) {
  if (!body.email || typeof payload !== "object" || payload === null) {
    return;
  }

  const otpPayload = payload as OnboardingOTPPayload;
  if (!otpPayload.debug_otp) {
    return;
  }

  await sendOnboardingOTPEmail({
    to: body.email,
    tenantSlug: otpPayload.tenant_slug ?? body.tenant_slug ?? "",
    otpCode: otpPayload.debug_otp,
    expiresInMinutes: Math.max(1, Math.floor((otpPayload.expires_in_seconds ?? 600) / 60)),
  });
}

async function callControlPlane(body: OnboardingBody) {
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  let path = "";
  let init: RequestInit = { method: "POST" };

  if (body.action === "check_slug") {
    const slug = encodeURIComponent(String(body.tenant_slug ?? ""));
    path = `/api/v1/onboarding/tenant-slugs/${slug}/availability`;
    init = { method: "GET" };
  } else if (body.action === "request_otp") {
    path = "/api/v1/onboarding/otp";
    init = {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        tenant_slug: body.tenant_slug,
        tenant_domain: body.tenant_domain,
        plan_slug: body.plan_slug ?? "free",
        email: body.email,
        mobile: body.mobile,
        customer_name: body.customer_name,
        metadata: body.metadata ?? {},
      }),
    };
  } else {
    path = "/api/v1/onboarding/otp/verify";
    init = {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        request_id: body.request_id,
        otp_code: body.otp_code,
      }),
    };
  }

  const response = await fetch(new URL(path, ensureTrailingSlash(baseURL)), {
    ...init,
    cache: "no-store",
  });
  const text = await response.text();
  const payload = safeParsePayload(text);
  if (!response.ok) {
    return {
      status: response.status,
      payload: {
        ok: false,
        error:
          extractErrorMessage(payload) ??
          `Control-plane onboarding endpoint returned ${response.status}`,
        details: payload,
      },
    };
  }
  return {
    status: response.status,
    payload,
  };
}

function ensureTrailingSlash(value: string) {
  return value.endsWith("/") ? value : `${value}/`;
}

function safeParsePayload(text: string) {
  if (!text) {
    return {};
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractErrorMessage(payload: unknown) {
  if (!payload || typeof payload !== "object") {
    return undefined;
  }
  const value = payload as { error?: { message?: string } | string; message?: string };
  if (typeof value.error === "string") {
    return value.error;
  }
  return value.error?.message ?? value.message;
}
