import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import ChatbotChatPage from "@/features/managed-rag/ChatbotChatPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function ManagedChatbotChatRoute({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const auth = await requireAuth(`/login?next=/chatbots/${encodeURIComponent(id)}/chat`);
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["chatbot:list", "chatbot:manage"]);
  const t = await getTranslations("managedRag");

  return (
    <AdminShell
      messageNamespaces={["managedRag"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("chat.title"), description: t("chat.pageDescription") }}
    >
      <ChatbotChatPage chatbotID={id} />
    </AdminShell>
  );
}
