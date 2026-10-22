import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { NewTicketPage } from "@/features/support-tickets/NewTicketPage";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function NewSupportTicketPage() {
  const auth = await requireAuth("/login?next=/support/tickets/new");
  const t = await getTranslations("support");

  return (
    <AdminShell messageNamespaces={["support"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("new.title"), description: t("new.description") }}
    >
      <NewTicketPage />
    </AdminShell>
  );
}
