import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { CreditPacksPanel } from "@/features/billing/credits/CreditPacksPanel";
import { parseCreditPackOrders, parseCreditPacks } from "@/features/billing/credits/api";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { canAnyPermission } from "@/lib/authz";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type PageProps = {
  searchParams: Promise<{ payment_status?: string }>;
};

export default async function BillingCreditsPage({ searchParams }: PageProps) {
  const auth = await requireAuth("/login?next=/billing/credits");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:read", "billing:manage", "settings:list"]);

  const t = await getTranslations("billing");
  const { payment_status: paymentStatus } = await searchParams;
  const [packsResponse, ordersResponse, permissionsPayload] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/credits/packs", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/credits/orders?limit=20", userID: auth.user.id }),
    getCurrentMePermissions(tenantSlug, auth.user.id).catch(() => ({ permissions: [] as string[] })),
  ]);
  const permissions = (permissionsPayload.permissions ?? []).map((item) => item.trim().toLowerCase());

  return (
    <AdminShell
      messageNamespaces={["billing"]}
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("credits.packs.title"), description: t("credits.packs.description") }}
    >
      <CreditPacksPanel
        packs={packsResponse.ok ? parseCreditPacks(packsResponse.payload) : []}
        orders={ordersResponse.ok ? parseCreditPackOrders(ordersResponse.payload) : []}
        canManage={canAnyPermission(permissions, ["billing:manage"])}
        paymentResult={paymentStatus === "success" || paymentStatus === "failed" ? paymentStatus : null}
        loadError={packsResponse.ok ? null : packsResponse.errorMessage ?? t("credits.packs.loadError")}
      />
    </AdminShell>
  );
}
