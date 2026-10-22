import AdminShell from "@/features/admin-shell/AdminShell";
import { SubscriptionOverview } from "@/features/billing/SubscriptionOverview";
import { parseSeatSummary, type SeatSummary } from "@/features/billing/seats/api";
import { parseCreditSettings, type CreditSettings } from "@/features/billing/credits/api";
import { CreditSettingsCard } from "@/features/billing/credits/CreditSettingsCard";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { canAnyPermission } from "@/lib/authz";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

export default async function BillingPlanPage() {
  const t = await getTranslations("billingPlan");
  const auth = await requireAuth("/login?next=/billing/plan");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  let initialError: string | null = null;
  let seatSummary: SeatSummary | null = null;
  let creditSettings: CreditSettings | null = null;

  // The subscription is the seat subscription: plans and their feature tables
  // are retired, so the page shows what the seats summary reports.
  try {
    const [seatSummaryResponse, creditSettingsResponse] = await Promise.all([
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/billing/seats",
        userID: auth.user.id,
      }),
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/billing/credits/settings",
        userID: auth.user.id,
      }),
    ]);
    // Reading credit settings needs billing:read; without it the card is left out.
    creditSettings = creditSettingsResponse.ok ? parseCreditSettings(creditSettingsResponse.payload) : null;

    if (!seatSummaryResponse.ok) {
      throw new Error(seatSummaryResponse.errorMessage ?? t("errors.loadSubscription"));
    }

    seatSummary = parseSeatSummary(seatSummaryResponse.payload);
  } catch (error) {
    initialError = error instanceof Error ? error.message : t("errors.loadPage");
  }

  const permissionsPayload = await getCurrentMePermissions(tenantSlug, auth.user.id).catch(() => ({ permissions: [] as string[] }));
  const permissions = (permissionsPayload.permissions ?? []).map((item) => item.trim().toLowerCase());

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["billingPlan", "billing", "billingSeats"]} tenantSlug={tenantSlug} currentUser={currentUser} pageHeader={{ title: t("title"), description: t("subtitle") }}>
      <SubscriptionOverview summary={seatSummary} initialError={initialError} />
      {creditSettings ? (
        <div className="mt-6">
          <CreditSettingsCard initialSettings={creditSettings} canManage={canAnyPermission(permissions, ["billing:manage"])} />
        </div>
      ) : null}
    </AdminShell>
  );
}
