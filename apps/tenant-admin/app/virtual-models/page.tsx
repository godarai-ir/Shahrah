import AdminShell from "@/app/components/admin-shell";
import VirtualModelManagement from "@/app/components/virtual-model-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

export default async function VirtualModelsPage() {
  const auth = await requireAuth("/login?next=/virtual-models");
  const t = await getTranslations("virtualModels");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, [
    "virtual_account:read",
    "virtual_account:manage",
    "virtual_account:create",
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
      pageHeader={{ title: t("title"), description: t("description") }}
    >
      <VirtualModelManagement />
    </AdminShell>
  );
}
