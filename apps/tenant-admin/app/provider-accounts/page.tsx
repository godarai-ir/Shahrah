import ProviderAccountWizard from "@/app/components/provider-account-wizard";
import AdminShell from "@/app/components/admin-shell";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function ProviderAccountsPage() {
  const auth = await requireAuth("/login?next=/provider-accounts");
  const tenantSlug = auth.tenantSlug;

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
