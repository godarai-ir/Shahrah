import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { resolveSectionLandingTab } from "@/features/admin-shell/section-landing";
import { GatewayPolicyRulesPage } from "@/features/gateway-policies/GatewayPolicyRulesPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

import GuardrailsPage from "../guardrails/page";

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

function buildDocsBaseHref(headersList: Headers, tenantSlug: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const docsHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return docsHost ? `${proto}://${docsHost}/docs` : "/docs";
}
