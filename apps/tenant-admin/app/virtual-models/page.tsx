import AdminShell from "@/app/components/admin-shell";
import VirtualModelManagement from "@/app/components/virtual-model-management";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function VirtualModelsPage() {
  const auth = await requireAuth("/login?next=/virtual-models");
  const tenantSlug = auth.tenantSlug;

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <VirtualModelManagement />
    </AdminShell>
  );
}

