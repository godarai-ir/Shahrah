import { getTranslations } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import { NewTicketPage } from "@/features/support-tickets/NewTicketPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function NewSupportTicketPage() {
  const auth = await requireAuth("/login?next=/support/tickets/new");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["support_ticket:create"]);
  const t = await getTranslations("support");

  return (
    <AdminShell
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("new.title"), description: t("new.description") }}
    >
      <NewTicketPage />
    </AdminShell>
  );
}
