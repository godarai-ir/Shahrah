import AdminShell from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { EditRateLimitRulePage } from "@/features/gateway-policies/rate-limiting/EditRateLimitRulePage";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ id: string }> };

export default async function Page({ params }: Params) {
  const auth = await requireAuth("/login?next=/gateway-policies/rate-limiting-rules");
  const t = await getTranslations("rateLimitRules");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:manage"]);
  const tenantSlug = auth.tenantSlug;
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };
  const { id } = await params;
  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("form.editTitle"), description: t("form.headerDescription") }}
    >
      <EditRateLimitRulePage ruleId={id} />
    </AdminShell>
  );
}
