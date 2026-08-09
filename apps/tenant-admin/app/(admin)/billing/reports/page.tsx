import AdminShell from "@/features/admin-shell/AdminShell";
import { ReportsPanel } from "@/features/billing/BillingFinancePanels";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import {
  buildControlPlaneBillingMonthlyPath,
  buildControlPlaneBillingSummaryPath,
  getCurrentJalaliMonthSelection,
} from "@/features/billing/domain";
import { parseBillingSummary, parseMonthlyUsage } from "@/features/billing/api";

export const dynamic = "force-dynamic";

export default async function BillingReportsPage() {
  const auth = await requireAuth("/login?next=/billing/reports");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:read", "settings:list"]);

  const currentMonth = getCurrentJalaliMonthSelection();
  const [summaryResponse, monthlyResponse] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: buildControlPlaneBillingSummaryPath("current_month"), userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: buildControlPlaneBillingMonthlyPath(currentMonth), userID: auth.user.id }),
  ]);

  const initialError = !summaryResponse.ok
    ? summaryResponse.errorMessage ?? "گزارش خلاصه دریافت نشد."
    : !monthlyResponse.ok
      ? monthlyResponse.errorMessage ?? "گزارش ماهانه دریافت نشد."
      : null;

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: "گزارش‌های مالی", description: "مصرف، هزینه، تخفیف و صرفه‌جویی کش را به تفکیک ارائه‌دهنده ببینید." }}
    >
      <div className="space-y-6">
        <PageHeader title="گزارش‌های مالی" description="مصرف، هزینه، تخفیف و صرفه‌جویی کش را به تفکیک ارائه‌دهنده ببینید." />
        <ReportsPanel
          initialSummary={summaryResponse.ok ? parseBillingSummary(summaryResponse.payload) : null}
          initialMonthly={monthlyResponse.ok ? parseMonthlyUsage(monthlyResponse.payload) : null}
          initialError={initialError}
        />
      </div>
    </AdminShell>
  );
}
