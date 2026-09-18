import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { BankTransferPanel } from "@/features/billing/BillingFinancePanels";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { type BankTransferRequest } from "@/features/billing/domain";
import { parseBankTransfers } from "@/features/billing/api";

export const dynamic = "force-dynamic";

export default async function BankTransferPage() {
  const auth = await requireAuth("/login?next=/billing/bank-transfer");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:manage", "settings:manage"]);

  const t = await getTranslations("billing");

  let initialError: string | null = null;
  let items: BankTransferRequest[] = [];
  const response = await fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/wallet/bank-transfers", userID: auth.user.id });
  if (response.ok) {
    items = parseBankTransfers(response.payload);
  } else {
    initialError = response.errorMessage ?? t("bankTransfer.loadError");
  }

  return (
    <AdminShell messageNamespaces={["billing"]}
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("bankTransfer.title"), description: t("bankTransfer.description") }}
    >
      <div className="space-y-6">
        <PageHeader title={t("bankTransfer.title")} description={t("bankTransfer.description")} />
        <BankTransferPanel initialItems={items} initialError={initialError} />
      </div>
    </AdminShell>
  );
}
