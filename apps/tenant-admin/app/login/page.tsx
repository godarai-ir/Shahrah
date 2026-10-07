import { headers } from "next/headers";
import { notFound, redirect } from "next/navigation";

import LoginForm from "@/features/auth/LoginForm";
import { getAuthContext } from "@/lib/auth/session";
import { getAuthService } from "@/lib/auth/factory";
import { decodeLoginIdentifierParam } from "@/lib/auth/login-prefill";
import { decideOIDCLoginStep } from "@/lib/auth/oidc-login-step";
import {
  getOIDCLoginRequest,
  normalizeOIDCRequestID,
  oidcContinuePath,
  rejectOIDCLogin,
} from "@/lib/auth/oidc-bridge";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

type LoginPageProps = {
  searchParams?: Promise<{
    next?: string;
    identifier?: string;
    oidc_request?: string;
    oidc_error?: string;
    oidc_reauth?: string;
  }>;
};

function sanitizeNextPath(nextPath: string | undefined) {
  if (!nextPath || !nextPath.startsWith("/") || nextPath.startsWith("//")) {
    return "/";
  }
  return nextPath;
}

function normalizeOIDCError(value: string | undefined) {
  return value === "expired" || value === "unavailable" ? value : undefined;
}

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const incomingHeaders = await headers();
  const tenantSlug = resolveTenantSlugFromHeaders(incomingHeaders);
  if (!tenantSlug) {
    notFound();
  }

  const auth = await getAuthContext();
  const resolvedSearchParams = (await searchParams) ?? {};
  const initialIdentifier = decodeLoginIdentifierParam(resolvedSearchParams.identifier);
  const oidcRequestID = normalizeOIDCRequestID(resolvedSearchParams.oidc_request);
  let redirectTo = sanitizeNextPath(resolvedSearchParams.next);
  let oidcClientName: string | undefined;

  if (oidcRequestID) {
    const loginRequest = await getOIDCLoginRequest(tenantSlug, oidcRequestID);
    if (!loginRequest.ok) {
      redirect(`/login?oidc_error=${loginRequest.code === "expired" ? "expired" : "unavailable"}`);
    }

    const continuePath = oidcContinuePath(oidcRequestID);
    const step = decideOIDCLoginStep({
      hasSession: Boolean(auth),
      prompt: loginRequest.value.prompt,
      reauthRequested: resolvedSearchParams.oidc_reauth === "1",
    });

    if (step === "continue") {
      redirect(continuePath);
    }

    if (step === "reject") {
      const rejection = await rejectOIDCLogin(tenantSlug, oidcRequestID);
      redirect(rejection.ok ? rejection.value : "/login?oidc_error=unavailable");
    }

    redirectTo = continuePath;
    oidcClientName = loginRequest.value.clientName;
  } else if (auth) {
    redirect(redirectTo);
  }

  const initialCaptcha = await getAuthService().createCaptcha();
  return (
    <LoginForm
      initialCaptcha={initialCaptcha}
      redirectTo={redirectTo}
      initialIdentifier={initialIdentifier}
      oidcClientName={oidcClientName}
      oidcError={normalizeOIDCError(resolvedSearchParams.oidc_error)}
    />
  );
}
