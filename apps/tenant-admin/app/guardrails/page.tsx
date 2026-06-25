import AdminShell from "@/app/components/admin-shell";
import { GuardrailsPage } from "@/features/guardrails/GuardrailsPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function Page() {
  const auth = await requireAuth("/login?next=/guardrails");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list"]);
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={auth.tenantSlug} currentUser={currentUser}>
      <GuardrailsPage />
    </AdminShell>
  );
}
