import AdminShell from "@/features/admin-shell/AdminShell";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import PromptRegistryListPage from "@/features/prompts/PromptRegistryListPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

export default async function PromptsPage() {
  const auth = await requireAuth("/login?next=/prompts");
  const t = await getTranslations("prompts");
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, auth.tenantSlug)}/prompt-registry`;
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list", "policy:manage"]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["prompts"]}
      tenantSlug={auth.tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("header.title"), description: t("header.description"), helpHref: docsHref }}
    >
      <PromptRegistryListPage />
    </AdminShell>
  );
}
