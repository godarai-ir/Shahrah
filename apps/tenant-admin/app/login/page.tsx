import { headers } from "next/headers";
import { notFound, redirect } from "next/navigation";

import LoginForm from "@/features/auth/LoginForm";
import { getAuthContext } from "@/lib/auth/session";
import { getAuthService } from "@/lib/auth/factory";
import { decodeLoginIdentifierParam } from "@/lib/auth/login-prefill";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

type LoginPageProps = {
  searchParams?: Promise<{
    next?: string;
    identifier?: string;
  }>;
};

function sanitizeNextPath(nextPath: string | undefined) {
  if (!nextPath || !nextPath.startsWith("/") || nextPath.startsWith("//")) {
    return "/";
  }
  return nextPath;
}

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const incomingHeaders = await headers();
  const tenantSlug = resolveTenantSlugFromHeaders(incomingHeaders);
  if (!tenantSlug) {
    notFound();
  }

  const auth = await getAuthContext();
  const resolvedSearchParams = (await searchParams) ?? {};
  const redirectTo = sanitizeNextPath(resolvedSearchParams.next);
  const initialIdentifier = decodeLoginIdentifierParam(resolvedSearchParams.identifier);

  if (auth) {
    redirect(redirectTo);
  }

  const initialCaptcha = await getAuthService().createCaptcha();
  return <LoginForm initialCaptcha={initialCaptcha} redirectTo={redirectTo} initialIdentifier={initialIdentifier} />;
}
