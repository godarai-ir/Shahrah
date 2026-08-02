import { NextResponse } from "next/server";

import { sendOnboardingOTPEmail } from "@/app/lib/email";
import { callSocialAuthControlPlane } from "@/app/lib/social-auth/control-plane";
import { socialAuth, socialAuthBaseURL } from "@/app/lib/social-auth/config";
import { normalizeOnboardingLoginURL } from "@/app/lib/social-auth/onboarding-login-url";
import { hasExpectedRequestOrigin } from "@/app/lib/social-auth/request-origin";

type OnboardingAction = "check_slug" | "request_otp" | "verify_otp" | "provision_social";

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

const errorMessages = {
  invalidRequest: "درخواست فعال‌سازی نامعتبر است.",
  controlPlaneUnavailable: "ارتباط با سرویس فعال‌سازی برقرار نشد.",
  controlPlaneFailed: "سرویس فعال‌سازی با خطای داخلی روبه‌رو شد.",
  slugInvalid: "شناسه فضای کاری نامعتبر یا رزروشده است.",
  slugTaken: "این شناسه فضای کاری در حال حاضر قابل انتخاب نیست.",
  requestNotFound: "درخواست فعال‌سازی پیدا نشد.",
  requestExpired: "مهلت کد تایید تمام شده است.",
  otpInvalid: "کد تایید نامعتبر است.",
} as const;

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as OnboardingBody | null;
  if (!body?.action) {
    return NextResponse.json({ ok: false, error: errorMessages.invalidRequest }, { status: 400 });
  }

  if (body.action === "provision_social") {
    if (!hasExpectedRequestOrigin(request, socialAuthBaseURL())) {
      return NextResponse.json({ ok: false, error: "مبدا درخواست احراز هویت معتبر نیست." }, { status: 403 });
    }
    const session = await socialAuth.api.getSession({ headers: request.headers });
    if (!session?.user?.id || !session.user.emailVerified) {
      return NextResponse.json({ ok: false, error: "حساب اجتماعی معتبر و ایمیل تاییدشده لازم است." }, { status: 401 });
    }
    const result = await callSocialAuthControlPlane<unknown>("/api/v1/auth/social/provision", {
      account_id: session.user.id,
      tenant_slug: body.tenant_slug,
      tenant_domain: resolveTenantDomain(request, body),
      plan_slug: body.plan_slug ?? "developer",
    });
    if (!result.ok) {
      return NextResponse.json({ ok: false, error: localizeUpstreamError(result.message, result.status) }, { status: result.status });
    }
    return NextResponse.json(normalizeProvisionedPayload(result.payload), { status: 201 });
  }

  let upstream: Awaited<ReturnType<typeof callControlPlane>>;
  try {
    upstream = await callControlPlane(request, body);
  } catch {
    return NextResponse.json(
      {
        ok: false,
        error: errorMessages.controlPlaneUnavailable,
      },
      { status: 502 },
    );
  }

  if (body.action === "request_otp" && upstream.status >= 200 && upstream.status < 300) {
    try {
      await sendOnboardingEmailIfPossible(body, upstream.payload);
    } catch {
      // The OTP response still contains debug_otp in development; do not fail the activation flow.
    }
  }

  return NextResponse.json(
    sanitizeOnboardingPayload(normalizeProvisionedPayload(upstream.payload)),
    { status: upstream.status },
  );
}

function normalizeProvisionedPayload(payload: unknown) {
  return normalizeOnboardingLoginURL(
    payload,
    process.env.GODARAI_TENANT_ORIGIN_TEMPLATE ?? "https://{tenant}.godarai.ir",
  );
}

function sanitizeOnboardingPayload(payload: unknown) {
  if (debugOTPExposureEnabled()) {
    return payload;
  }
  if (!payload || typeof payload !== "object") {
    return payload;
  }

  const rest = { ...(payload as Record<string, unknown>) };
  delete rest.debug_otp;
  return rest;
}

function debugOTPExposureEnabled() {
  const nodeEnv = process.env.NODE_ENV?.trim().toLowerCase() ?? "";
  if (nodeEnv !== "development" && nodeEnv !== "test") {
    return false;
  }

  return process.env.SHAHRAH_EXPOSE_DEBUG_OTP?.trim().toLowerCase() === "true";
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
    expiresInMinutes: Math.max(1, Math.floor((otpPayload.expires_in_seconds ?? 300) / 60)),
  });
}

async function callControlPlane(request: Request, body: OnboardingBody) {
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
        tenant_domain: resolveTenantDomain(request, body),
        plan_slug: body.plan_slug ?? "developer",
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
        error: localizeUpstreamError(extractErrorMessage(payload), response.status),
        details: payload,
      },
    };
  }
  return {
    status: response.status,
    payload,
  };
}

function resolveTenantDomain(request: Request, body: OnboardingBody) {
  const slug = normalizeTenantSlug(body.tenant_slug);
  if (!slug) {
    return "";
  }

  const tenantOriginTemplate = process.env.GODARAI_TENANT_ORIGIN_TEMPLATE;
  if (tenantOriginTemplate?.includes("{tenant}")) {
    const templateDomain = normalizeHostnameFromURL(tenantOriginTemplate.replace("{tenant}", slug));
    if (templateDomain) return templateDomain;
  }

  const explicitDomain = normalizeHostname(body.tenant_domain);
  if (process.env.NODE_ENV !== "production" && explicitDomain) {
    return explicitDomain;
  }

  const baseDomain =
    normalizeHostname(process.env.SHAHRAH_TENANT_BASE_DOMAIN) ||
    normalizeHostname(process.env.TENANT_BASE_DOMAIN) ||
    normalizeHostnameFromURL(process.env.NEXT_PUBLIC_SITE_URL) ||
    normalizeHostname(request.headers.get("x-forwarded-host")) ||
    normalizeHostname(request.headers.get("host"));

  if (!baseDomain || isLocalhostDomain(baseDomain)) {
    return `${slug}.localhost`;
  }

  return `${slug}.${stripWWW(baseDomain)}`;
}

function localizeUpstreamError(message: string | undefined, status: number) {
  const normalized = (message ?? "").toLowerCase();
  if (normalized.includes("invalid or reserved")) {
    return errorMessages.slugInvalid;
  }
  if (normalized.includes("already taken")) {
    return errorMessages.slugTaken;
  }
  if (normalized.includes("not found")) {
    return errorMessages.requestNotFound;
  }
  if (normalized.includes("expired")) {
    return errorMessages.requestExpired;
  }
  if (normalized.includes("invalid onboarding otp")) {
    return errorMessages.otpInvalid;
  }
  if (status === 409) {
    return errorMessages.slugTaken;
  }
  if (status === 410) {
    return errorMessages.requestExpired;
  }
  if (status >= 500) {
    if (message && message !== "Internal Server Error") {
      return message;
    }
    return errorMessages.controlPlaneFailed;
  }
  return errorMessages.controlPlaneFailed;
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

function normalizeTenantSlug(value: string | undefined) {
  return (value ?? "").trim().toLowerCase().replace(/[^a-z0-9-]/g, "");
}

function normalizeHostname(value: string | null | undefined) {
  const first = (value ?? "").split(",")[0]?.trim().toLowerCase() ?? "";
  if (!first) {
    return "";
  }
  const withoutProtocol = first.replace(/^https?:\/\//, "");
  const host = withoutProtocol.split("/")[0] ?? "";
  return host.replace(/:\d+$/, "").replace(/\.$/, "");
}

function normalizeHostnameFromURL(value: string | undefined) {
  if (!value) {
    return "";
  }
  try {
    return normalizeHostname(new URL(value).hostname);
  } catch {
    return normalizeHostname(value);
  }
}

function stripWWW(hostname: string) {
  return hostname.startsWith("www.") ? hostname.slice(4) : hostname;
}

function isLocalhostDomain(hostname: string) {
  return hostname === "localhost" || hostname.endsWith(".localhost") || hostname.startsWith("127.");
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
