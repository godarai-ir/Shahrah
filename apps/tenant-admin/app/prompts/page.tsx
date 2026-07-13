import AdminShell from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import PromptRegistryListPage from "@/app/components/prompt-registry-list-page";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function PromptsPage() {
  const auth = await requireAuth("/login?next=/prompts");
  const t = await getTranslations("prompts");
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
      pageHeader={{ title: t("header.title"), description: t("header.description") }}
    >
      <PromptRegistryListPage />
    </AdminShell>
  );
}
