import { NextResponse } from "next/server";

import { callSocialAuthControlPlane } from "@/app/lib/social-auth/control-plane";
import { socialAuth } from "@/app/lib/social-auth/config";

type Handoff = { code: string; tenant_slug: string; expires_at: string };

export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const url = new URL(request.url);
  const session = await socialAuth.api.getSession({ headers: request.headers });
  if (!session?.user?.id || !session.user.emailVerified) {
    return NextResponse.redirect(new URL("/onboarding?social_error=unverified", url.origin), 303);
  }

  const flow = url.searchParams.get("flow");
  const tenant = normalizeTenantSlug(url.searchParams.get("tenant"));
  const returnTo = normalizeReturnTo(url.searchParams.get("return_to"));
  let result;
  if (flow === "login" && tenant) {
    result = await callSocialAuthControlPlane<Handoff>("/api/v1/auth/social/handoff", {
      tenant_slug: tenant,
      account_id: session.user.id,
    });
  } else if (flow === "link") {
    result = await callSocialAuthControlPlane<Handoff>("/api/v1/auth/social/link", {
      ticket_code: url.searchParams.get("ticket") ?? "",
      account_id: session.user.id,
    });
  } else {
    return NextResponse.redirect(new URL("/onboarding?social_error=context", url.origin), 303);
  }
  if (!result.ok) {
    if (tenant) {
      const errorURL = new URL("/login", tenantOrigin(tenant));
      errorURL.searchParams.set("social_error", String(result.status));
      return NextResponse.redirect(errorURL, 303);
    }
    return NextResponse.redirect(new URL(`/onboarding?social_error=${result.status}`, url.origin), 303);
  }

  const destination = tenantOrigin(result.payload.tenant_slug);
  const handoffURL = new URL("/auth/handoff", destination);
  handoffURL.searchParams.set("code", result.payload.code);
  handoffURL.searchParams.set("next", returnTo);
  return NextResponse.redirect(handoffURL, 303);
}

function tenantOrigin(slug: string) {
  const template = process.env.GODARAI_TENANT_ORIGIN_TEMPLATE ?? "https://{tenant}.godarai.ir";
  return template.replace("{tenant}", slug);
}
function normalizeTenantSlug(value: string | null) {
  const normalized = (value ?? "").trim().toLowerCase();
  return /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/.test(normalized) ? normalized : "";
}
function normalizeReturnTo(value: string | null) {
  const normalized = (value ?? "").trim();
  return normalized.startsWith("/") && !normalized.startsWith("//") ? normalized.slice(0, 1024) : "/";
}
