import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import ChatbotsPage from "@/features/managed-rag/ChatbotsPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function ManagedChatbotsRoute() {
  const auth = await requireAuth("/login?next=/chatbots");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["chatbot:list", "chatbot:manage"]);
  const t = await getTranslations("managedRag");

  return (
    <AdminShell
      messageNamespaces={["managedRag"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("chatbots.title"), description: t("chatbots.description") }}
    >
      <ChatbotsPage />
    </AdminShell>
  );
}
