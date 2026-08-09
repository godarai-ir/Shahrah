import Link from "next/link";
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
import { buildControlPlaneBillingSummaryPath, formatIRR } from "@/features/billing/domain";
import { parseBillingSummary, parseWallet } from "@/features/billing/api";

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

  const [walletResponse, summaryResponse] = await Promise.all([
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/wallet", userID: auth.user.id }),
    fetchControlPlaneJSON({ tenantSlug, method: "GET", path: buildControlPlaneBillingSummaryPath("current_month"), userID: auth.user.id }),
  ]);

  const wallet = walletResponse.ok ? parseWallet(walletResponse.payload) : null;
  const summary = summaryResponse.ok ? parseBillingSummary(summaryResponse.payload) : null;
  const availableBalance = Number(wallet?.available_balance_irr ?? wallet?.balance_irr ?? 0);

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: "مدیریت مالی", description: "نمای عملیاتی کیف پول، مصرف دوره، تخفیف‌ها و اسناد مالی فضای کاری." }}
    >
      <div className="flex flex-col gap-6">
        <PageHeader
          eyebrow="Billing"
          title="مدیریت مالی"
          description="نمای عملیاتی کیف پول، مصرف دوره، تخفیف‌ها و اسناد مالی فضای کاری."
          actions={
            <>
              <Button asChild><Link href="/billing/topup"><CreditCard className="me-2 size-4" />شارژ کیف پول</Link></Button>
              <Button asChild variant="outline"><Link href="/billing/bank-transfer"><Landmark className="me-2 size-4" />واریز بانکی</Link></Button>
            </>
          }
        />
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3 md:gap-4">
          <MetricCard
            icon={WalletCards}
            label="اعتبار"
            value={formatIRR(availableBalance)}
          />
          <MetricCard icon={LineChart} label="مصرف ماه جاری" value={formatIRR(summary?.total_usage_irr)} />
          <MetricCard icon={BarChart3} label="صرفه‌جویی ناشی از کش" value={formatIRR(summary?.cache_savings_irr)} />
        </div>
        <div className="grid gap-4 xl:grid-cols-2">
          <AppCard title="اقدام‌های سریع" className="lg:col-span-1">
            <div className="grid grid-cols-2 gap-2 lg:grid-cols-1">
              <QuickAction href="/billing/topup" icon={CreditCard} label="شارژ از درگاه" />
              <QuickAction href="/billing/bank-transfer" icon={Landmark} label="ثبت واریز بانکی" />
              <QuickAction href="/billing/reports" icon={LineChart} label="گزارش مصرف" />
              <QuickAction href="/billing/profile" icon={WalletCards} label="پروفایل مالی" />
              <QuickAction href="/billing/documents" icon={FileText} label="اسناد مالی" />
            </div>
          </AppCard>
          <AppCard title="ارائه‌دهندگان پر مصرف" description="براساس هزینه نهایی دوره جاری">
            <div className="divide-y divide-border">
              {(summary?.providers ?? []).slice(0, 5).map((provider) => (
                <div key={provider.provider_name} className="flex items-center justify-between gap-4 py-3">
                  <div className="min-w-0">
                    <p className="font-medium text-foreground">{provider.provider_name || "unknown"}</p>
                    <p className="text-xs text-muted-foreground">{provider.request_count} requests</p>
                  </div>
                  <StatusBadge status="neutral">{formatIRR(provider.final_cost_irr)}</StatusBadge>
                </div>
              ))}
              {(summary?.providers ?? []).length === 0 ? <p className="py-8 text-center text-sm text-muted-foreground">داده مصرفی برای این دوره وجود ندارد.</p> : null}
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
