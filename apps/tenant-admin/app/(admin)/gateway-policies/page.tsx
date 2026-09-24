import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { resolveSectionLandingTab } from "@/features/admin-shell/section-landing";
import { GatewayPolicyRulesPage } from "@/features/gateway-policies/GatewayPolicyRulesPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

import GuardrailsPage from "../guardrails/page";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

/** Rate and budget rules in one list. Falls through to guardrails for callers who cannot list rules. */
export default async function GatewayPoliciesPage() {
  const auth = await requireAuth("/login?next=/gateway-policies");
  const tab = await resolveSectionLandingTab("/gateway-policies", auth.tenantSlug, auth.user.id);

  if (tab === "guardrails") return <GuardrailsPage />;
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
