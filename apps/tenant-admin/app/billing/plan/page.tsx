import AdminShell from "@/app/components/admin-shell";
import PlanManagement from "@/app/components/plan-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

type DataList<T> = {
  data?: T[];
};

export default async function BillingPlanPage() {
  const t = await getTranslations("billingPlan");
  const auth = await requireAuth("/login?next=/billing/plan");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  let initialError: string | null = null;
  let plans: unknown[] = [];
  let subscription: unknown = null;
  let entitlements: unknown = null;

  try {
    const [plansResponse, subscriptionResponse, entitlementResponse] = await Promise.all([
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/commercial-plans",
        userID: auth.user.id,
      }),
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
    ]);

    if (!plansResponse.ok) {
      throw new Error(plansResponse.errorMessage ?? t("errors.loadPlans"));
    }
    if (!subscriptionResponse.ok) {
      throw new Error(subscriptionResponse.errorMessage ?? t("errors.loadSubscription"));
    }
    if (!entitlementResponse.ok) {
      throw new Error(entitlementResponse.errorMessage ?? t("errors.loadEntitlements"));
    }

    plans = ((plansResponse.payload ?? {}) as DataList<unknown>).data ?? [];
    subscription = subscriptionResponse.payload;
    entitlements = entitlementResponse.payload;
  } catch (error) {
    initialError = error instanceof Error ? error.message : t("errors.loadPage");
  }

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser} pageHeader={{ title: t("title") }}>
      <PlanManagement
        plans={plans as never}
        subscription={subscription as never}
        entitlements={entitlements as never}
        initialError={initialError}
      />
    </AdminShell>
  );
}
