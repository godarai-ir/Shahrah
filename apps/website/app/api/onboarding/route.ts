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

const errorMessages = {
  invalidRequest: "درخواست فعال‌سازی نامعتبر است.",
  controlPlaneFailed: "ارتباط با سرویس فعال‌سازی برقرار نشد.",
  slugInvalid: "شناسه تننت نامعتبر یا رزروشده است.",
  slugTaken: "این شناسه تننت در حال حاضر قابل انتخاب نیست.",
  requestNotFound: "درخواست فعال‌سازی پیدا نشد.",
  requestExpired: "مهلت کد تایید تمام شده است.",
  otpInvalid: "کد تایید نامعتبر است.",
} as const;

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as OnboardingBody | null;
  if (!body?.action) {
    return NextResponse.json({ ok: false, error: errorMessages.invalidRequest }, { status: 400 });
  }

  try {
    const upstream = await callControlPlane(request, body);
    if (body.action === "request_otp" && upstream.status >= 200 && upstream.status < 300) {
      await sendOnboardingEmailIfPossible(body, upstream.payload);
    }
    return NextResponse.json(upstream.payload, { status: upstream.status });
  } catch {
    return NextResponse.json(
      {
        ok: false,
        error: errorMessages.controlPlaneFailed,
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

  const explicitDomain = normalizeHostname(body.tenant_domain);
  if (explicitDomain && !isLocalhostDomain(explicitDomain)) {
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
