import { NextRequest, NextResponse } from "next/server";

import {
  isSocialAuthProvider,
  normalizeSocialReturnPath,
  resolveCentralSocialAuthBaseURL,
} from "@/lib/auth/social-auth-url";
import { resolveTenantContextFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const tenant = resolveTenantContextFromHeaders(request.headers);
  if (tenant.ambiguous) {
    return NextResponse.json({ error: "Ambiguous tenant context." }, { status: 400 });
  }
  if (!tenant.tenantSlug) {
    return NextResponse.json({ error: "Tenant host is required." }, { status: 400 });
  }

  const provider = request.nextUrl.searchParams.get("provider") ?? "";
  if (!isSocialAuthProvider(provider)) {
    return NextResponse.json({ error: "Invalid social provider." }, { status: 400 });
  }

  const authBase = resolveCentralSocialAuthBaseURL();
  if (!authBase) {
    return NextResponse.json({ error: "Social authentication is not configured." }, { status: 503 });
  }

  const redirectURL = new URL("/auth/social/start", authBase);
  redirectURL.searchParams.set("flow", "login");
  redirectURL.searchParams.set("provider", provider);
  redirectURL.searchParams.set("tenant", tenant.tenantSlug);
  redirectURL.searchParams.set(
    "return_to",
    normalizeSocialReturnPath(request.nextUrl.searchParams.get("return_to")),
  );
  return NextResponse.redirect(redirectURL, 307);
}
