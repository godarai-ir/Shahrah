import { getTranslations } from "next-intl/server";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import AdminShell from "@/features/admin-shell/AdminShell";
import { parseCreditSettings } from "@/features/billing/credits/api";
import { parseSeatMembers, parseSeatOrders, parseSeatRequests, parseSeatSummary } from "@/features/billing/seats/api";
import { SeatsWorkspace } from "@/features/billing/seats/SeatsWorkspace";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { canAnyPermission } from "@/lib/authz";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import type { Permission } from "@/lib/permissions";

export const dynamic = "force-dynamic";

type PageProps = {
  searchParams: Promise<{ payment_status?: string }>;
};

const SEAT_READ_PERMISSIONS: Permission[] = ["billing:read", "billing:manage", "settings:list", "user:list", "user:manage"];

export default async function BillingSeatsPage({ searchParams }: PageProps) {
  const auth = await requireAuth("/login?next=/billing/seats");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, SEAT_READ_PERMISSIONS);

  const t = await getTranslations("billingSeats");
  const { payment_status: paymentStatus } = await searchParams;
  const request = (path: string) => fetchControlPlaneJSON({ tenantSlug, method: "GET", path, userID: auth.user.id });
  const [summaryResponse, membersResponse, requestsResponse, ordersResponse, settingsResponse, permissionsPayload] = await Promise.all([
    request("/api/v1/billing/seats"),
    request("/api/v1/billing/seats/members?limit=500"),
    request("/api/v1/billing/seats/requests?limit=100"),
    request("/api/v1/billing/seats/orders?limit=20"),
    request("/api/v1/billing/credits/settings"),
    getCurrentMePermissions(tenantSlug, auth.user.id).catch(() => ({ permissions: [] as string[] })),
  ]);
  const permissions = (permissionsPayload.permissions ?? []).map((item) => item.trim().toLowerCase());
  const summary = summaryResponse.ok ? parseSeatSummary(summaryResponse.payload) : null;

  return (
    <AdminShell
      messageNamespaces={["billing", "billingSeats"]}
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("title"), description: t("description") }}
    >
      {summary ? (
        <SeatsWorkspace
          summary={summary}
          members={membersResponse.ok ? parseSeatMembers(membersResponse.payload) : []}
          requests={requestsResponse.ok ? parseSeatRequests(requestsResponse.payload) : []}
          orders={ordersResponse.ok ? parseSeatOrders(ordersResponse.payload) : []}
          settings={settingsResponse.ok ? parseCreditSettings(settingsResponse.payload) : null}
          canManageMembers={canAnyPermission(permissions, ["user:manage"])}
          canManageBilling={canAnyPermission(permissions, ["billing:manage"])}
          paymentResult={paymentStatus === "success" || paymentStatus === "failed" ? paymentStatus : null}
        />
      ) : (
        <Alert variant="destructive">
          <AlertTitle>{t("errors.title")}</AlertTitle>
          <AlertDescription>{summaryResponse.errorMessage ?? t("errors.loadSummary")}</AlertDescription>
        </Alert>
      )}
    </AdminShell>
  );
}
