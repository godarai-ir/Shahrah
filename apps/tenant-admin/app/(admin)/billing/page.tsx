import Link from "next/link";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { BadgeCheck, Coins } from "lucide-react";

import AdminShell from "@/features/admin-shell/AdminShell";
import { Button } from "@/components/ui/button";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { buildControlPlaneCreditSummaryPath } from "@/features/billing/domain";
import { parseCreditUsageSummary } from "@/features/billing/api";
import { parseCreditBalance } from "@/features/billing/credits/api";
import { CreditOverview } from "@/features/billing/credits/CreditOverview";
import { CreditUsagePanel } from "@/features/billing/credits/CreditUsagePanel";
import { buildApexHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

export default async function BillingDashboardPage() {
  const auth = await requireAuth("/login?next=/billing");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:read", "settings:list"]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  const [balanceResponse, usageResponse] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/billing/credits/balance", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: buildControlPlaneCreditSummaryPath("current_month"), userID: auth.user.id }),
  ]);

  const t = await getTranslations("billing");
  const balance = balanceResponse.ok ? parseCreditBalance(balanceResponse.payload) : null;
  const usage = usageResponse.ok ? parseCreditUsageSummary(usageResponse.payload) : null;
  const changelogHref = buildApexHref(await headers(), tenantSlug, "/pricing/changelog");

  return (
    <AdminShell
      messageNamespaces={["billing"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("overview.title"), description: t("overview.description") }}
    >
      <div className="flex flex-col gap-6">
        <div className="flex flex-wrap gap-2">
          <Button asChild><Link href="/billing/credits"><Coins data-icon="inline-start" />{t("overview.buyCredits")}</Link></Button>
          <Button asChild variant="outline"><Link href="/billing/plan"><BadgeCheck data-icon="inline-start" />{t("overview.plan")}</Link></Button>
        </div>
        <CreditOverview balance={balance} usage={usage} />
        <CreditUsagePanel initialSummary={usage} changelogHref={changelogHref} />
      </div>
    </AdminShell>
  );
}
