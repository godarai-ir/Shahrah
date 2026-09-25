import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { GatewayPolicyRulesPage } from "@/features/gateway-policies/GatewayPolicyRulesPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

/** Rate and budget rules in one list. */
export default async function GatewayPoliciesPage() {
  const auth = await requireAuth("/login?next=/gateway-policies");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list"]);

  const t = await getTranslations("gatewayPolicies");
  const docsHref = `${buildDocsBaseHref(await headers(), auth.tenantSlug)}/guides/rate-limiting`;
  return (
    <AdminShell
      messageNamespaces={["gatewayPolicies", "rateLimitRules", "budgetLimitRules"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("title"), description: t("description"), helpHref: docsHref }}
    >
      <GatewayPolicyRulesPage />
    </AdminShell>
  );
}
