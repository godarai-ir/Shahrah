import AdminShell from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { RateLimitRulesPage } from "@/features/gateway-policies/rate-limiting/RateLimitRulesPage";
import { headers } from "next/headers";

export const dynamic = "force-dynamic";

export default async function Page() {
  const auth = await requireAuth("/login?next=/gateway-policies/rate-limiting-rules");
  const t = await getTranslations("rateLimitRules");
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, auth.tenantSlug)}/guides/rate-limiting`;
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list"]);
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
      pageHeader={{ title: t("title"), description: t("description"), helpHref: docsHref }}
    >
      <RateLimitRulesPage />
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
