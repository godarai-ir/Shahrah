import ProviderAccountWizard from "@/app/components/provider-account-wizard";
import AdminShell from "@/app/components/admin-shell";
import LLMIntegrationsManagement from "@/app/components/llm-integrations-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

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
  const tenantSlug = auth.tenantSlug;
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
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      {params.account ? (
        <ProviderAccountWizard initialAccountID={params.account} />
      ) : (
        <LLMIntegrationsManagement initialProvider={params.provider} initialTab={params.tab === "connected" ? "connected" : "all"} />
      )}
    </AdminShell>
  );
}
