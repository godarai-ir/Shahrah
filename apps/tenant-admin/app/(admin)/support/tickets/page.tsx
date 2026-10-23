import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { TicketListPage } from "@/features/support-tickets/TicketListPage";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function SupportTicketsPage() {
  const auth = await requireAuth("/login?next=/support/tickets");
  const t = await getTranslations("support");
  const role = auth.user.role?.trim().toLowerCase();
  const canReadAll = role === "tenant_admin" || role === "proxy_admin";

  return (
    <AdminShell messageNamespaces={["support"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("list.title"), description: t("list.description") }}
    >
      <TicketListPage canReadAll={canReadAll} />
    </AdminShell>
  );
}
