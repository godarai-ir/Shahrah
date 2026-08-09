import AdminShell from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { RateLimitRuleForm } from "@/features/gateway-policies/rate-limiting/RateLimitRuleForm";

export const dynamic = "force-dynamic";

export default async function Page() {
  const auth = await requireAuth("/login?next=/gateway-policies/rate-limiting-rules/new");
  const t = await getTranslations("rateLimitRules");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:manage"]);
  const tenantSlug = auth.tenantSlug;
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };
  return (
    <AdminShell messageNamespaces={["rateLimitRules"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("form.addTitle"), description: t("form.headerDescription") }}
    >
      <RateLimitRuleForm />
    </AdminShell>
  );
}
