import { NextResponse } from "next/server";

import {
  isSocialAuthProvider,
  isSocialAuthProviderEnabled,
  socialAuth,
  socialAuthBaseURL,
} from "@/app/lib/social-auth/config";
import { resolveRequestOrigin } from "@/app/lib/social-auth/request-origin";

export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const incoming = new URL(request.url);
  const provider = incoming.searchParams.get("provider") ?? "";
  const flow = incoming.searchParams.get("flow") ?? "login";
  if (!isSocialAuthProvider(provider) || !isSocialAuthProviderEnabled(provider)) {
    return NextResponse.json({ error: "Social provider is not configured." }, { status: 400 });
  }
  if (!new Set(["login", "onboarding", "link"]).has(flow)) {
    return NextResponse.json({ error: "Invalid authentication flow." }, { status: 400 });
  }

  const tenant = normalizeTenantSlug(incoming.searchParams.get("tenant"));
  const returnTo = normalizeReturnTo(incoming.searchParams.get("return_to"));
  const plan = normalizePlan(incoming.searchParams.get("plan"));
  const ticket = normalizeOpaqueCode(incoming.searchParams.get("ticket"));
  if ((flow === "login" && !tenant) || (flow === "link" && !ticket)) {
    return NextResponse.json({ error: "Required authentication context is missing." }, { status: 400 });
  }

  const authBase = socialAuthBaseURL();
  if (resolveRequestOrigin(request, incoming) !== new URL(authBase).origin) {
    const canonical = new URL("/auth/social/start", authBase);
    canonical.searchParams.set("flow", flow);
    canonical.searchParams.set("provider", provider);
    if (tenant) canonical.searchParams.set("tenant", tenant);
    if (returnTo) canonical.searchParams.set("return_to", returnTo);
    if (plan) canonical.searchParams.set("plan", plan);
    if (ticket) canonical.searchParams.set("ticket", ticket);
    return NextResponse.redirect(canonical, 307);
  }

  const complete = new URL(flow === "onboarding" ? "/onboarding" : "/auth/social/complete", authBase);
  complete.searchParams.set("flow", flow);
  complete.searchParams.set("provider", provider);
  if (tenant) complete.searchParams.set("tenant", tenant);
  if (returnTo) complete.searchParams.set("return_to", returnTo);
  if (plan) complete.searchParams.set("plan", plan);
  if (ticket) complete.searchParams.set("ticket", ticket);
  if (flow === "onboarding") complete.searchParams.set("social", "1");
  const errorCallback = new URL("/auth/social/error", authBase);
  errorCallback.searchParams.set("flow", flow);
  if (tenant) errorCallback.searchParams.set("tenant", tenant);

  const authResponse = await socialAuth.api.signInSocial({
    body: {
      provider,
      callbackURL: complete.toString(),
      errorCallbackURL: errorCallback.toString(),
    },
    headers: request.headers,
    asResponse: true,
  });
  const authorizationURL = authResponse.headers.get("location");
  if (!authorizationURL) {
    return authResponse;
  }

  const redirectHeaders = new Headers(authResponse.headers);
  redirectHeaders.set("Location", authorizationURL);
  return new NextResponse(null, {
    status: 302,
    headers: redirectHeaders,
  });
}

function normalizeTenantSlug(value: string | null) {
  const normalized = (value ?? "").trim().toLowerCase();
  return /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/.test(normalized) ? normalized : "";
}
function normalizeReturnTo(value: string | null) {
  const normalized = (value ?? "").trim();
  return normalized.startsWith("/") && !normalized.startsWith("//") ? normalized.slice(0, 1024) : "/";
}
function normalizePlan(value: string | null) {
  const normalized = (value ?? "").trim().toLowerCase();
  return ["developer", "team", "growth", "enterprise"].includes(normalized) ? normalized : "developer";
}
function normalizeOpaqueCode(value: string | null) {
  const normalized = (value ?? "").trim();
  return /^[A-Za-z0-9_-]{32,256}$/.test(normalized) ? normalized : "";
}
