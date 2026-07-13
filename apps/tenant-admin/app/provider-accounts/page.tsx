import ProviderAccountWizard from "@/app/components/provider-account-wizard";
import AdminShell from "@/app/components/admin-shell";
import LLMIntegrationsManagement from "@/app/components/llm-integrations-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

type ProviderAccountsPageProps = {
  searchParams?: Promise<{
    account?: string;
    provider?: string;
    tab?: string;
  }>;
};

export default async function ProviderAccountsPage({ searchParams }: ProviderAccountsPageProps) {
  const params = (await searchParams) ?? {};
  const auth = await requireAuth("/login?next=/provider-accounts");
  const t = await getTranslations("providerAccounts");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);
  const providerAccountsDocsHref = `${docsBaseHref}/guides/tenant-team-account-model`;
  await requirePermissionOr403(tenantSlug, auth.user.id, [
    "provider_account:read",
    "provider_account:manage",
    "provider_account:create",
  ]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={params.account
        ? { title: t("ux.panel.workspaceTitle"), description: t("ux.panel.workspaceHint"), helpHref: providerAccountsDocsHref }
        : { title: t("integrations.pageTitle"), description: t("integrations.pageDescription"), helpHref: providerAccountsDocsHref }}
    >
      {params.account ? (
        <ProviderAccountWizard initialAccountID={params.account} />
      ) : (
        <LLMIntegrationsManagement
          initialProvider={params.provider}
          initialTab={params.tab === "connected" ? "connected" : "all"}
          docsHref={providerAccountsDocsHref}
        />
      )}
    </AdminShell>
  );
}

function buildDocsBaseHref(headersList: Headers, tenantSlug: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const docsHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return docsHost ? `${proto}://${docsHost}/docs` : "/docs";
}
