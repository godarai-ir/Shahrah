import AdminShell from "@/features/admin-shell/AdminShell";
import { GuardrailsPage } from "@/features/guardrails/GuardrailsPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

import { buildDocsBaseHref } from "@/lib/docs-href";
export const dynamic = "force-dynamic";

export default async function Page() {
  const auth = await requireAuth("/login?next=/guardrails");
  const t = await getTranslations("guardrails");
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, auth.tenantSlug)}/guardrails`;
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list"]);
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["guardrails"]}
      tenantSlug={auth.tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("description"), helpHref: docsHref }}
    >
      <GuardrailsPage docsHref={docsHref} />
    </AdminShell>
  );
}
