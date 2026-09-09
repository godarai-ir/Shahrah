import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { BillingProfilePanel } from "@/features/billing/BillingFinancePanels";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { parseBillingProfile } from "@/features/billing/api";

export const dynamic = "force-dynamic";

export default async function BillingProfilePage() {
  const auth = await requireAuth("/login?next=/billing/profile");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:profile_manage", "billing:manage", "settings:manage"]);

  const t = await getTranslations("billing");
  const response = await fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/profile", userID: auth.user.id });
  const initialProfile = response.ok ? parseBillingProfile(response.payload) : null;
  const initialError = response.ok || response.status === 404 ? null : response.errorMessage ?? t("profile.loadError");

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("profile.title"), description: t("profile.description") }}
    >
      <div className="flex flex-col gap-6">
        <PageHeader title={t("profile.title")} description={t("profile.description")} />
        <BillingProfilePanel initialProfile={initialProfile} initialError={initialError} />
      </div>
    </AdminShell>
  );
}
