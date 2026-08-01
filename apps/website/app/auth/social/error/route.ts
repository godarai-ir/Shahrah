import { NextResponse } from "next/server";

import { socialAuthBaseURL } from "@/app/lib/social-auth/config";

export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const incoming = new URL(request.url);
  if (incoming.origin !== new URL(socialAuthBaseURL()).origin) {
    return NextResponse.json({ error: "Invalid authentication origin." }, { status: 400 });
  }

  const flow = incoming.searchParams.get("flow");
  const tenant = normalizeTenantSlug(incoming.searchParams.get("tenant"));
  if ((flow === "login" || flow === "link") && tenant) {
    const destination = new URL("/login", tenantOrigin(tenant));
    destination.searchParams.set("social_error", "provider");
    return NextResponse.redirect(destination, 303);
  }

  return NextResponse.redirect(
    new URL("/onboarding?social_error=provider", socialAuthBaseURL()),
    303,
  );
}

function normalizeTenantSlug(value: string | null) {
  const normalized = (value ?? "").trim().toLowerCase();
  return /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/.test(normalized)
    ? normalized
    : "";
}

function tenantOrigin(slug: string) {
  const template = process.env.GODARAI_TENANT_ORIGIN_TEMPLATE ?? "https://{tenant}.godarai.ir";
  return template.replace("{tenant}", slug);
}
