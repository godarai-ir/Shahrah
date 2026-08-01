import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { isSocialAuthProvider, resolveCentralSocialAuthBaseURL } from "@/lib/auth/social-auth-url";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type LinkTicket = { code: string; expires_at: string };

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  const sessionAge = Date.now() - new Date(auth.session.createdAtISO).getTime();
  if (!Number.isFinite(sessionAge) || sessionAge > 10 * 60 * 1000) {
    return NextResponse.json({ error: { message: "برای اتصال حساب، دوباره وارد شوید تا احراز هویت تازه باشد." } }, { status: 401 });
  }
  const body = (await request.json().catch(() => null)) as { provider?: unknown } | null;
  const provider = typeof body?.provider === "string" ? body.provider : "";
  if (!isSocialAuthProvider(provider)) {
    return NextResponse.json({ error: { message: "ارائه‌دهنده نامعتبر است." } }, { status: 400 });
  }
  const authBase = resolveCentralSocialAuthBaseURL();
  if (!authBase) {
    return NextResponse.json({ error: { message: "ورود اجتماعی پیکربندی نشده است." } }, { status: 503 });
  }
  const result = await fetchControlPlaneJSON({
    tenantSlug: auth.tenantSlug,
    userID: auth.user.id,
    method: "POST",
    path: "/api/v1/auth/social/link-ticket",
    body: { tenant_slug: auth.tenantSlug, user_id: auth.user.id },
    requestHeaders: request.headers,
  });
  if (!result.ok) return NextResponse.json({ error: { message: result.errorMessage } }, { status: result.status });
  const ticket = result.payload as LinkTicket;
  const redirectURL = new URL("/auth/social/start", authBase);
  redirectURL.searchParams.set("flow", "link");
  redirectURL.searchParams.set("provider", provider);
  redirectURL.searchParams.set("ticket", ticket.code);
  redirectURL.searchParams.set("tenant", auth.tenantSlug);
  redirectURL.searchParams.set("return_to", "/settings/account?linked=1");
  return NextResponse.json({ redirect_url: redirectURL.toString(), expires_at: ticket.expires_at });
}
