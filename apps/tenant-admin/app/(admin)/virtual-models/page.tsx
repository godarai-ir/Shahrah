import AdminShell from "@/features/admin-shell/AdminShell";
import VirtualModelManagement from "@/features/virtual-models/VirtualModelManagement";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

import { buildDocsBaseHref } from "@/lib/docs-href";
export const dynamic = "force-dynamic";

export default async function VirtualModelsPage() {
  const auth = await requireAuth("/login?next=/virtual-models");
  const t = await getTranslations("virtualModels");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, tenantSlug)}/virtual-model`;
  await requirePermissionOr403(tenantSlug, auth.user.id, [
    "virtual_account:read",
    "virtual_account:manage",
    "virtual_account:create",
  ]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["virtualModels","providerAccounts"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("description"), helpHref: docsHref }}
    >
      <VirtualModelManagement docsHref={docsHref} />
    </AdminShell>
  );
}
