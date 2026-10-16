import Link from "next/link";
import { headers } from "next/headers";
import { getLocale, getTranslations } from "next-intl/server";
import type { ComponentType } from "react";
import { BarChart3, CreditCard, FileText, Landmark, LineChart, WalletCards } from "lucide-react";

import AdminShell from "@/features/admin-shell/AdminShell";
import { AppCard } from "@/components/app-ui/AppCard";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Button } from "@/components/ui/button";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { buildControlPlaneBillingSummaryPath, buildControlPlaneCreditSummaryPath, formatIRRAmount } from "@/features/billing/domain";
import { parseBillingSummary, parseCreditUsageSummary, parseWallet } from "@/features/billing/api";
import { CreditShadowCard } from "@/features/billing/CreditShadowCard";
import { resolveLocale } from "@/lib/locale";
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

  const [walletResponse, summaryResponse, creditSummaryResponse] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/wallet", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: buildControlPlaneBillingSummaryPath("current_month"), userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: buildControlPlaneCreditSummaryPath("current_month"), userID: auth.user.id }),
  ]);

  const t = await getTranslations("billing");
  const locale = resolveLocale(await getLocale());
  const irrUnit = t("units.irr");
  const formatIRR = (value: number | undefined | null) => `${formatIRRAmount(value, locale)} ${irrUnit}`;

  const wallet = walletResponse.ok ? parseWallet(walletResponse.payload) : null;
  const summary = summaryResponse.ok ? parseBillingSummary(summaryResponse.payload) : null;
  const creditSummary = creditSummaryResponse.ok ? parseCreditUsageSummary(creditSummaryResponse.payload) : null;
  const availableBalance = Number(wallet?.available_balance_irr ?? wallet?.balance_irr ?? 0);
  const changelogHref = buildApexHref(await headers(), tenantSlug, "/pricing/changelog");

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("overview.title"), description: t("overview.description") }}
    >
      <div className="flex flex-col gap-6">
        <PageHeader
          eyebrow={t("overview.eyebrow")}
          title={t("overview.title")}
          description={t("overview.description")}
          actions={
            <>
              <Button asChild><Link href="/billing/topup"><CreditCard className="me-2 size-4" />{t("overview.topUp")}</Link></Button>
              <Button asChild variant="outline"><Link href="/billing/bank-transfer"><Landmark className="me-2 size-4" />{t("overview.bankTransfer")}</Link></Button>
            </>
          }
        />
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3 md:gap-4">
          <MetricCard
            icon={WalletCards}
            label={t("overview.balance")}
            value={formatIRR(availableBalance)}
          />
          <MetricCard icon={LineChart} label={t("overview.currentMonthUsage")} value={formatIRR(summary?.total_usage_irr)} />
          <MetricCard icon={BarChart3} label={t("overview.cacheSavings")} value={formatIRR(summary?.cache_savings_irr)} />
        </div>
        {creditSummary?.available ? (
          <CreditShadowCard
            summary={creditSummary}
            locale={locale}
            changelog={{ href: changelogHref, label: t("credits.changelog") }}
            labels={{
              title: t("credits.title"),
              description: t("credits.description"),
              badge: t("credits.badge"),
              unit: t("credits.unit"),
              total: t("credits.total"),
              cacheSaved: t("credits.cacheSaved"),
              byModel: t("credits.byModel"),
              byTeam: t("credits.byTeam"),
              noTeam: t("credits.noTeam"),
              emptyTitle: t("credits.emptyTitle"),
              emptyDescription: t("credits.emptyDescription"),
              requestCount: (count) => t("credits.requestCount", { count }),
            }}
          />
        ) : null}
        <div className="grid gap-4 xl:grid-cols-2">
          <AppCard title={t("overview.quickActions")} className="lg:col-span-1">
            <div className="grid grid-cols-2 gap-2 lg:grid-cols-1">
              <QuickAction href="/billing/topup" icon={CreditCard} label={t("overview.quickTopUp")} />
              <QuickAction href="/billing/bank-transfer" icon={Landmark} label={t("overview.quickBankTransfer")} />
              <QuickAction href="/billing/reports" icon={LineChart} label={t("overview.quickReports")} />
              <QuickAction href="/billing/profile" icon={WalletCards} label={t("overview.quickProfile")} />
              <QuickAction href="/billing/documents" icon={FileText} label={t("overview.quickDocuments")} />
            </div>
          </AppCard>
          <AppCard title={t("overview.topProviders")} description={t("overview.topProvidersDescription")}>
            <div className="divide-y divide-border">
              {(summary?.providers ?? []).slice(0, 5).map((provider) => (
                <div key={provider.provider_name} className="flex items-center justify-between gap-4 py-3">
                  <div className="min-w-0">
                    <p className="font-medium text-foreground">{provider.provider_name || t("overview.unknownProvider")}</p>
                    <p className="text-xs text-muted-foreground">{t("overview.requestCount", { count: provider.request_count })}</p>
                  </div>
                  <StatusBadge status="neutral">{formatIRR(provider.final_cost_irr)}</StatusBadge>
                </div>
              ))}
              {(summary?.providers ?? []).length === 0 ? <p className="py-8 text-center text-sm text-muted-foreground">{t("overview.noUsage")}</p> : null}
            </div>
          </AppCard>
        </div>
      </div>
    </AdminShell>
  );
}

function QuickAction({ href, icon: Icon, label }: { href: string; icon: ComponentType<{ className?: string }>; label: string }) {
  return (
    <Button asChild variant="outline" size="sm" className="h-9 justify-start px-3 text-xs sm:h-10 sm:text-sm">
      <Link href={href} className="flex w-full items-center justify-between text-start lg:justify-start">
        <Icon className="me-2 size-4 shrink-0" />
        <span className="flex-1 text-start">{label}</span>
      </Link>
    </Button>
  );
}
