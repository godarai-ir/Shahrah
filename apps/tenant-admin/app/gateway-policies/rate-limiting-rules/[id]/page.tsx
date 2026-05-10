import AdminShell from "@/app/components/admin-shell";
import { requireAuth } from "@/lib/auth/session";
import { EditRateLimitRulePage } from "@/features/gateway-policies/rate-limiting/EditRateLimitRulePage";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ id: string }> };

export default async function Page({ params }: Params) {
  const auth = await requireAuth("/login?next=/gateway-policies/rate-limiting-rules");
  const tenantSlug = auth.tenantSlug;
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };
  const { id } = await params;
  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <EditRateLimitRulePage ruleId={id} />
    </AdminShell>
  );
}
