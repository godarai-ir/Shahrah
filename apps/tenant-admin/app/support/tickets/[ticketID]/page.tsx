import { getTranslations } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import { TicketDetailPage } from "@/features/support-tickets/TicketDetailPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function SupportTicketDetailPage({ params }: { params: Promise<{ ticketID: string }> }) {
  const { ticketID } = await params;
  const auth = await requireAuth(`/login?next=/support/tickets/${encodeURIComponent(ticketID)}`);
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["support_ticket:read_own", "support_ticket:read_all"]);
  const t = await getTranslations("support");

  return (
    <AdminShell
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("detail.title"), description: t("detail.description") }}
    >
      <TicketDetailPage ticketID={ticketID} currentUserID={auth.user.id} />
    </AdminShell>
  );
}
