import AdminShell from "@/app/components/admin-shell";
import TopUpPanel from "@/app/components/topup-panel";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function TopUpPage() {
  const auth = await requireAuth("/login?next=/billing/topup");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <TopUpPanel initialError={null} />
    </AdminShell>
  );
}
