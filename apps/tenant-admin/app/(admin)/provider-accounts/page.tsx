import ProviderAccountWizard from "@/features/provider-accounts/ProviderAccountWizard";
import AdminShell from "@/features/admin-shell/AdminShell";
import { QueryProvider } from "@/components/providers/QueryProvider";
import { ModelCatalogPage } from "@/features/provider-accounts/model-catalog/model-catalog-page";
import type { WorkspaceTab } from "@/features/provider-accounts/types";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

type ProviderAccountsPageProps = {
  searchParams?: Promise<{
    account?: string;
    section?: string;
  }>;
};

export default async function ProviderAccountsPage({ searchParams }: ProviderAccountsPageProps) {
  const params = (await searchParams) ?? {};
  const auth = await requireAuth("/login?next=/provider-accounts");
  const catalogT = await getTranslations("modelCatalog");
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
    <AdminShell messageNamespaces={["providerAccounts","modelCatalog"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={params.account
          ? {
            title: catalogT("page.detailsTitle"),
            description: catalogT("page.detailsDescription"),
            helpHref: providerAccountsDocsHref,
          }
        : {
            title: catalogT("page.title"),
            description: catalogT("page.description"),
            helpHref: docsBaseHref,
          }}
    >
      {params.account ? (
        <QueryProvider>
          <ProviderAccountWizard
            initialAccountID={params.account}
            initialSection={normalizeWorkspaceSection(params.section)}
            currentUser={currentUser}
            docsBaseHref={docsBaseHref}
          />
        </QueryProvider>
      ) : (
        <ModelCatalogPage docsHref={providerAccountsDocsHref} />
      )}
    </AdminShell>
  );
}

function normalizeWorkspaceSection(value?: string): WorkspaceTab {
  if (value === "models" || value === "access" || value === "credentials") return value;
  return "overview";
}
