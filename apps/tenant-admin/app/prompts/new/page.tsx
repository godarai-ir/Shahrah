import AdminShell from "@/app/components/admin-shell";
import PromptBuilderPage from "@/app/components/prompt-builder-page";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

export default async function NewPromptPage() {
  const auth = await requireAuth("/login?next=/prompts/new");
  const t = await getTranslations("prompts");
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, auth.tenantSlug);
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
      pageHeader={{
        title: t("builder.titleNew"),
        description: t("builder.description"),
        helpHref: `${docsBaseHref}/quick-start#prompts`,
      }}
    >
      <PromptBuilderPage mode="create" docsHref={`${docsBaseHref}/quick-start#prompts`} />
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
