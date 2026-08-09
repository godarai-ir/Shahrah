import TenantAdminDashboard from "@/features/dashboard/TenantAdminDashboard";
import AdminShell from "@/features/admin-shell/AdminShell";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { requireAuth } from "@/lib/auth/session";
import { getTranslations } from "next-intl/server";
import {
  effectiveEntitlementsSchema,
  tenantAdminUserListResponseSchema,
  tenantSubscriptionSchema,
} from "@/features/dashboard/contracts";

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
  const parsedSubscription = tenantSubscriptionSchema.safeParse(subscriptionResult.payload);
  const parsedEntitlements = effectiveEntitlementsSchema.safeParse(entitlementsResult.payload);
  const parsedUsers = tenantAdminUserListResponseSchema.safeParse(usersResult.payload);
  const initialSubscription = subscriptionResult.ok && parsedSubscription.success ? parsedSubscription.data : null;
  const initialEntitlements = entitlementsResult.ok && parsedEntitlements.success ? parsedEntitlements.data : null;
  const initialUsersTotal = usersResult.ok && parsedUsers.success ? parsedUsers.data.total ?? 0 : null;

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
