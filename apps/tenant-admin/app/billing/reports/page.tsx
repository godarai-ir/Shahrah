import AdminShell from "@/app/components/admin-shell";
import { ReportsPanel } from "@/app/components/billing-finance-panels";
import { PageHeader } from "@/components/app-ui";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { type BillingUsageSummary, type MonthlyUsageDetail } from "@/app/billing/_shared";

export const dynamic = "force-dynamic";

export default async function BillingReportsPage() {
  const auth = await requireAuth("/login?next=/billing/reports");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:read", "settings:list"]);

  const month = new Date().toISOString().slice(0, 7);
  const [summaryResponse, monthlyResponse] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/reports/summary?range=current_month", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: `/api/v1/billing/reports/monthly?month=${month}`, userID: auth.user.id }),
  ]);

  const initialError = !summaryResponse.ok
    ? summaryResponse.errorMessage ?? "گزارش خلاصه دریافت نشد."
    : !monthlyResponse.ok
      ? monthlyResponse.errorMessage ?? "گزارش ماهانه دریافت نشد."
      : null;

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}>
      <div className="space-y-6">
        <PageHeader title="گزارش‌های مالی" description="مصرف، هزینه، تخفیف و صرفه‌جویی cache را به تفکیک provider ببینید." />
        <ReportsPanel
          initialSummary={summaryResponse.ok ? (summaryResponse.payload as BillingUsageSummary) : null}
          initialMonthly={monthlyResponse.ok ? (monthlyResponse.payload as MonthlyUsageDetail) : null}
          initialError={initialError}
        />
      </div>
    </AdminShell>
  );
}
