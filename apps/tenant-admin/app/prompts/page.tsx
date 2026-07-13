import AdminShell from "@/app/components/admin-shell";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import PromptRegistryListPage from "@/app/components/prompt-registry-list-page";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function PromptsPage() {
  const auth = await requireAuth("/login?next=/prompts");
  const t = await getTranslations("prompts");
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, auth.tenantSlug)}/quick-start#prompts`;
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list", "policy:manage"]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell
      tenantSlug={auth.tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("header.title"), description: t("header.description"), helpHref: docsHref }}
    >
      <PromptRegistryListPage docsHref={docsHref} />
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
