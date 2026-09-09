import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { BillingDocumentsPanel } from "@/features/billing/BillingFinancePanels";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { parseDiscountRedemptions, parseInvoices, parseWalletTransactions } from "@/features/billing/api";

export const dynamic = "force-dynamic";

export default async function BillingDocumentsPage() {
  const auth = await requireAuth("/login?next=/billing/documents");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:documents_read", "billing:read", "settings:list"]);

  const t = await getTranslations("billing");

  const [invoicesResponse, transactionsResponse, discountsResponse] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/documents/invoices", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/documents/transactions", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/documents/discounts", userID: auth.user.id }),
  ]);

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("documents.title"), description: t("documents.description") }}
    >
      <div className="space-y-6">
        <PageHeader title={t("documents.title")} description={t("documents.description")} />
        <BillingDocumentsPanel
          invoices={invoicesResponse.ok ? parseInvoices(invoicesResponse.payload) : []}
          transactions={transactionsResponse.ok ? parseWalletTransactions(transactionsResponse.payload) : []}
          discounts={discountsResponse.ok ? parseDiscountRedemptions(discountsResponse.payload) : []}
        />
      </div>
    </AdminShell>
  );
}
