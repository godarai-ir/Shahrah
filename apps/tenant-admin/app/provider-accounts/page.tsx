import ProviderAccountWizard from "@/app/components/provider-account-wizard";
import AdminShell from "@/app/components/admin-shell";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { extractTenantSlugFromHost } from "@/lib/tenant";

export const dynamic = "force-dynamic";

export default async function ProviderAccountsPage() {
  const auth = await requireAuth("/login?next=/provider-accounts");

  const incomingHeaders = await headers();
  const incomingHost =
    incomingHeaders.get("x-forwarded-host") ??
    incomingHeaders.get("host") ??
    "";

  const tenantSlug =
    extractTenantSlugFromHost(incomingHost) ||
    process.env.SHAHRAH_DEFAULT_TENANT ||
    "default";

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <ProviderAccountWizard />
    </AdminShell>
  );
}
