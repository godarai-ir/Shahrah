import TenantAdminDashboard from "@/app/components/tenant-admin-dashboard";
import AdminShell from "@/app/components/admin-shell";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { requireAuth } from "@/lib/auth/session";
import { getTranslations } from "next-intl/server";

type TenantSubscription = {
  plan_slug?: string;
  plan_display_name?: string;
  status: string;
  billing_interval: string;
  current_period_start?: string;
  current_period_end?: string;
  trial_started_at?: string;
  trial_ends_at?: string;
};

type EffectiveEntitlements = {
  subscription?: TenantSubscription | null;
  features?: Record<string, unknown>;
  usage_counters?: Array<{
    feature_key: string;
    used: number;
    limit?: number | null;
    window_end?: string;
  }>;
};

type TenantAdminUserListResponse = {
  total?: number;
};

export const dynamic = "force-dynamic";

export default async function Home() {
  const auth = await requireAuth("/login");
  const t = await getTranslations("dashboard");
  const tenantSlug = auth.tenantSlug;

  const [subscriptionResult, entitlementsResult, usersResult] = await Promise.all([
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
      path: "/api/v1/users?limit=1&offset=0",
      userID: auth.user.id,
    }),
  ]);
  const initialSubscription = subscriptionResult.ok ? (subscriptionResult.payload as TenantSubscription) : null;
  const initialEntitlements = entitlementsResult.ok ? (entitlementsResult.payload as EffectiveEntitlements) : null;
  const initialUsersTotal = usersResult.ok ? ((usersResult.payload as TenantAdminUserListResponse).total ?? 0) : null;

  const errorMessage =
    subscriptionResult.errorMessage ??
    entitlementsResult.errorMessage ??
    usersResult.errorMessage ??
    null;

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["dashboard"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle") }}
    >
      <TenantAdminDashboard
        initialSubscription={initialSubscription}
        initialEntitlements={initialEntitlements}
        initialUsersTotal={initialUsersTotal}
        initialError={errorMessage}
      />
    </AdminShell>
  );
}
