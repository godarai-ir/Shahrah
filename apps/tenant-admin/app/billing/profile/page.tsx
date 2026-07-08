import AdminShell from "@/app/components/admin-shell";
import { BillingProfilePanel } from "@/app/components/billing-finance-panels";
import { PageHeader } from "@/components/app-ui";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { type BillingProfile } from "@/app/billing/_shared";

export const dynamic = "force-dynamic";

export default async function BillingProfilePage() {
  const auth = await requireAuth("/login?next=/billing/profile");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:profile_manage", "billing:manage", "settings:manage"]);

  const response = await fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/profile", userID: auth.user.id });
  const initialProfile = response.ok ? (response.payload as BillingProfile) : null;
  const initialError = response.ok || response.status === 404 ? null : response.errorMessage ?? "پروفایل مالی دریافت نشد.";

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}>
      <div className="space-y-6">
        <PageHeader title="پروفایل مالی" description="اطلاعات حقیقی یا حقوقی tenant برای فاکتور و اسناد مالی." />
        <BillingProfilePanel initialProfile={initialProfile} initialError={initialError} />
      </div>
    </AdminShell>
  );
}
