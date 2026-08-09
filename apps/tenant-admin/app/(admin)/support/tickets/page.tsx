import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { TicketListPage } from "@/features/support-tickets/TicketListPage";
import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function SupportTicketsPage() {
  const auth = await requireAuth("/login?next=/support/tickets");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["support_ticket:read_own", "support_ticket:read_all"]);
  const [t, permissions] = await Promise.all([
    getTranslations("support"),
    getCurrentMePermissions(auth.tenantSlug, auth.user.id),
  ]);
  const permissionSet = getNormalizedPermissionSet(permissions);

  return (
    <AdminShell messageNamespaces={["support"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("list.title"), description: t("list.description") }}
    >
      <TicketListPage canReadAll={permissionSet.has("support_ticket:read_all")} />
    </AdminShell>
  );
}
