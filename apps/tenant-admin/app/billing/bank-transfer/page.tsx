import AdminShell from "@/app/components/admin-shell";
import { BankTransferPanel } from "@/app/components/billing-finance-panels";
import { PageHeader } from "@/components/app-ui";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { type BankTransferRequest, unwrapList } from "@/app/billing/_shared";

export const dynamic = "force-dynamic";

export default async function BankTransferPage() {
  const auth = await requireAuth("/login?next=/billing/bank-transfer");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:manage", "settings:manage"]);

  let initialError: string | null = null;
  let items: BankTransferRequest[] = [];
  const response = await fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/wallet/bank-transfers", userID: auth.user.id });
  if (response.ok) {
    items = unwrapList<BankTransferRequest>(response.payload);
  } else {
    initialError = response.errorMessage ?? "واریزهای بانکی دریافت نشد.";
  }

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: "واریز بانکی", description: "رسید واریز را ثبت کنید تا پس از تایید، اعتبار کیف پول اضافه شود." }}
    >
      <div className="space-y-6">
        <PageHeader title="واریز بانکی" description="رسید واریز را ثبت کنید تا پس از تایید، اعتبار کیف پول اضافه شود." />
        <BankTransferPanel initialItems={items} initialError={initialError} />
      </div>
    </AdminShell>
  );
}
