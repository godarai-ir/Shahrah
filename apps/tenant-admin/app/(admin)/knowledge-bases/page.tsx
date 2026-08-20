import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import KnowledgeBasesPage from "@/features/managed-rag/KnowledgeBasesPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function ManagedKnowledgeBasesRoute() {
  const auth = await requireAuth("/login?next=/knowledge-bases");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["knowledge_base:list", "knowledge_base:manage"]);
  const t = await getTranslations("managedRag");

  return (
    <AdminShell
      messageNamespaces={["managedRag"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("knowledgeBases.title"), description: t("knowledgeBases.description") }}
    >
      <KnowledgeBasesPage />
    </AdminShell>
  );
}
