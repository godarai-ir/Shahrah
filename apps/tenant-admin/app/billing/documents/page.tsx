import AdminShell from "@/app/components/admin-shell";
import { BillingDocumentsPanel } from "@/app/components/billing-finance-panels";
import { PageHeader } from "@/components/app-ui";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { type DiscountRedemption, type Invoice, type WalletTransaction, unwrapList } from "@/app/billing/_shared";

export const dynamic = "force-dynamic";

export default async function BillingDocumentsPage() {
  const auth = await requireAuth("/login?next=/billing/documents");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:documents_read", "billing:read", "settings:list"]);

  const [invoicesResponse, transactionsResponse, discountsResponse] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/documents/invoices", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/documents/transactions", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/documents/discounts", userID: auth.user.id }),
  ]);

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: "اسناد مالی", description: "فاکتورها، تراکنش‌های کیف پول و تخفیف‌های فضای کاری." }}
    >
      <div className="space-y-6">
        <PageHeader title="اسناد مالی" description="فاکتورها، تراکنش‌های کیف پول و تخفیف‌های فضای کاری." />
        <BillingDocumentsPanel
          invoices={invoicesResponse.ok ? unwrapList<Invoice>(invoicesResponse.payload) : []}
          transactions={transactionsResponse.ok ? unwrapList<WalletTransaction>(transactionsResponse.payload) : []}
          discounts={discountsResponse.ok ? unwrapList<DiscountRedemption>(discountsResponse.payload) : []}
        />
      </div>
    </AdminShell>
  );
}
