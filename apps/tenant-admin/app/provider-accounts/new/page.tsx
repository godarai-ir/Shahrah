import ProviderAccountWizard from "@/app/components/provider-account-wizard";
import AdminShell from "@/app/components/admin-shell";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { isModelCatalogV2Enabled } from "@/lib/features/model-catalog";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

type NewProviderAccountPageProps = {
  searchParams?: Promise<{
    provider?: string;
  }>;
};

export default async function NewProviderAccountPage({ searchParams }: NewProviderAccountPageProps) {
  const params = (await searchParams) ?? {};
  const auth = await requireAuth("/login?next=/provider-accounts/new");
  const t = await getTranslations("providerAccounts");
  const catalogT = await getTranslations("modelCatalog");
  const modelCatalogV2Enabled = isModelCatalogV2Enabled();
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);
  await requirePermissionOr403(tenantSlug, auth.user.id, ["provider_account:create"]);

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
      pageHeader={{
        title: modelCatalogV2Enabled ? catalogT("page.addTitle") : t("ux.panel.setupTitle"),
        description: modelCatalogV2Enabled ? catalogT("page.addDescription") : t("ux.panel.setupHint"),
      }}
    >
      <ProviderAccountWizard
        mode="create"
        initialProvider={params.provider}
        docsBaseHref={docsBaseHref}
        catalogExperience={modelCatalogV2Enabled}
        currentUser={currentUser}
      />
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
