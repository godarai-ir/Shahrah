import AdminShell from "@/app/components/admin-shell";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { RateLimitRulesPage } from "@/features/gateway-policies/rate-limiting/RateLimitRulesPage";

export const dynamic = "force-dynamic";

export default async function Page() {
  const auth = await requireAuth("/login?next=/gateway-policies/rate-limiting-rules");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list"]);
  const tenantSlug = auth.tenantSlug;
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };
  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <RateLimitRulesPage />
    </AdminShell>
  );
}
