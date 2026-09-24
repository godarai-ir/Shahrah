import AdminShell from "@/features/admin-shell/AdminShell";
import { GuardrailWizard } from "@/features/guardrails/wizard/GuardrailWizard";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

export default async function NewGuardrailPage() {
  const auth = await requireAuth("/login?next=/guardrails/new");
  const t = await getTranslations("guardrails");
  const docsHref = `${buildDocsBaseHref(await headers(), auth.tenantSlug)}/guardrails`;
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:manage"]);

  return (
    <AdminShell messageNamespaces={["guardrails"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("wizard.title"), description: t("wizard.description"), helpHref: docsHref }}
    >
      <GuardrailWizard mode="create" />
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
