import ProviderAccountWizard from "@/features/provider-accounts/ProviderAccountWizard";
import AdminShell from "@/features/admin-shell/AdminShell";
import { QueryProvider } from "@/components/providers/QueryProvider";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

type NewProviderAccountPageProps = {
  searchParams?: Promise<{
    provider?: string;
  }>;
};

export default async function NewProviderAccountPage({ searchParams }: NewProviderAccountPageProps) {
  const params = (await searchParams) ?? {};
  const auth = await requireAuth("/login?next=/provider-accounts/new");
  const catalogT = await getTranslations("modelCatalog");
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
    <AdminShell messageNamespaces={["providerAccounts","modelCatalog"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{
        title: catalogT("page.addTitle"),
        description: catalogT("page.addDescription"),
      }}
    >
      <QueryProvider>
        <ProviderAccountWizard
          mode="create"
          initialProvider={params.provider}
          docsBaseHref={docsBaseHref}
          currentUser={currentUser}
        />
      </QueryProvider>
    </AdminShell>
  );
}
