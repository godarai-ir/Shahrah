import AdminShell, { buildDocsHref } from "@/features/admin-shell/AdminShell";
import PlanManagement from "@/features/billing/PlanManagement";
import { parseCreditSettings, type CreditSettings } from "@/features/billing/credits/api";
import { CreditSettingsCard } from "@/features/billing/credits/CreditSettingsCard";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { canAnyPermission } from "@/lib/authz";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { getTranslations } from "next-intl/server";
import { headers } from "next/headers";

export const dynamic = "force-dynamic";

export default async function BillingPlanPage() {
  const t = await getTranslations("billingPlan");
  const auth = await requireAuth("/login?next=/billing/plan");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsHref(requestHeaders, tenantSlug);
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  let initialError: string | null = null;
  let subscription: unknown = null;
  let entitlements: unknown = null;
  let creditSettings: CreditSettings | null = null;

  try {
    const [subscriptionResponse, entitlementResponse, creditSettingsResponse] = await Promise.all([
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/subscription",
        userID: auth.user.id,
      }),
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/entitlements",
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

    if (!subscriptionResponse.ok) {
      throw new Error(subscriptionResponse.errorMessage ?? t("errors.loadSubscription"));
    }
    if (!entitlementResponse.ok) {
      throw new Error(entitlementResponse.errorMessage ?? t("errors.loadEntitlements"));
    }

    subscription = subscriptionResponse.payload;
    entitlements = entitlementResponse.payload;
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
    <AdminShell messageNamespaces={["billingPlan", "billing"]} tenantSlug={tenantSlug} currentUser={currentUser} pageHeader={{ title: t("title"), description: t("subtitle") }}>
      <PlanManagement
        subscription={subscription as never}
        entitlements={entitlements as never}
        docsBaseHref={docsBaseHref}
        initialError={initialError}
      />
      {creditSettings ? (
        <div className="mt-6">
          <CreditSettingsCard initialSettings={creditSettings} canManage={canAnyPermission(permissions, ["billing:manage"])} />
        </div>
      ) : null}
    </AdminShell>
  );
}
