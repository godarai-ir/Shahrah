import { unstable_cache } from "next/cache";
import { headers } from "next/headers";
import { notFound, redirect } from "next/navigation";

import LoginForm from "@/app/components/login-form";
import { getAuthContext } from "@/lib/auth/session";
import { createCaptchaChallenge } from "@/lib/auth/store";
import { prisma } from "@/lib/prisma";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

type LoginPageProps = {
  searchParams?: Promise<{
    next?: string;
  }>;
};

function sanitizeNextPath(nextPath: string | undefined) {
  if (!nextPath || !nextPath.startsWith("/") || nextPath.startsWith("//")) {
    return "/";
  }
  return nextPath;
}

const tenantIsActive = unstable_cache(
  async (tenantSlug: string) => {
    const rows = await prisma.$queryRaw<Array<{ tenant_id: string }>>`
      SELECT tenant_id
      FROM tenants
      WHERE slug = ${tenantSlug} AND status = 'active'
      LIMIT 1
    `;
    return Boolean(rows[0]);
  },
  ["tenant-admin-login-active-tenant"],
  {
    revalidate: 30,
  },
);

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const incomingHeaders = await headers();
  const tenantSlug = resolveTenantSlugFromHeaders(incomingHeaders);
  if (!tenantSlug || !(await tenantIsActive(tenantSlug))) {
    notFound();
  }

  const auth = await getAuthContext();
  const resolvedSearchParams = (await searchParams) ?? {};
  const redirectTo = sanitizeNextPath(resolvedSearchParams.next);

  if (auth) {
    redirect(redirectTo);
  }

  const initialCaptcha = await createCaptchaChallenge();
  return <LoginForm initialCaptcha={initialCaptcha} redirectTo={redirectTo} />;
}
