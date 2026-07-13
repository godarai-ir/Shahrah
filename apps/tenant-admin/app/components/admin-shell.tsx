import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";
import { headers } from "next/headers";

import AdminShellClient from "@/app/components/admin-shell-client";
import type { ShellPageMetadata } from "@/components/app-ui/ShellPageHeading";
import type { WalletPayload } from "@/app/billing/_shared";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type AdminShellProps = {
  children: React.ReactNode;
  tenantSlug: string;
  pageHeader?: ShellPageMetadata;
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
};

type TenantSubscription = {
  plan_slug?: string;
  plan_display_name?: string;
  status?: string;
  current_period_end?: string;
  trial_ends_at?: string;
};

export default async function AdminShell({
  children,
  tenantSlug,
  pageHeader,
  currentUser,
}: AdminShellProps) {
  const locale = await getLocale();
  const messages = await getMessages();
  const requestHeaders = await headers();
  const docsHref = buildDocsHref(requestHeaders, tenantSlug);
  const appVersion = getAppVersion();
  const [permissions, walletSummary, subscriptionSummary] = await Promise.all([
    loadCurrentPermissions(tenantSlug, currentUser.id),
    loadWalletSummary(tenantSlug, currentUser.id),
    loadSubscriptionSummary(tenantSlug, currentUser.id),
  ]);

  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <AdminShellClient
        tenantSlug={tenantSlug}
        currentUser={currentUser}
        permissions={permissions}
        walletSummary={walletSummary}
        subscriptionSummary={subscriptionSummary}
        docsHref={docsHref}
        appVersion={appVersion}
        pageHeader={pageHeader}
      >
        {children}
      </AdminShellClient>
    </NextIntlClientProvider>
  );
}

function getAppVersion(): string {
  return process.env.NEXT_PUBLIC_APP_VERSION?.replace(/^v/, "") || "0.0.0-dev";
}

function buildDocsHref(headersList: Headers, tenantSlug: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const docsHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return docsHost ? `${proto}://${docsHost}/docs/` : "/docs/";
}

async function loadWalletSummary(tenantSlug: string, userID: string): Promise<{
  availableBalanceIRR: number | null;
  ledgerBalanceIRR: number | null;
  committedUnsettledIRR: number;
  reservedIRR: number;
  syncStatus: string | null;
}> {
  const response = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/wallet",
    userID,
  });

  if (!response.ok) {
    return {
      availableBalanceIRR: null,
      ledgerBalanceIRR: null,
      committedUnsettledIRR: 0,
      reservedIRR: 0,
      syncStatus: null,
    };
  }

  const payload = response.payload as WalletPayload;
  const availableBalance = Number(payload.available_balance_irr ?? payload.balance_irr ?? 0);
  const ledgerBalance = Number(payload.ledger_balance_irr ?? payload.balance_irr ?? 0);
  const committedUnsettled = Number(payload.committed_unsettled_irr ?? payload.pending_irr ?? 0);
  const reserved = Number(payload.reserved_irr ?? 0);

  return {
    availableBalanceIRR: Number.isFinite(availableBalance) ? availableBalance : null,
    ledgerBalanceIRR: Number.isFinite(ledgerBalance) ? ledgerBalance : null,
    committedUnsettledIRR: Number.isFinite(committedUnsettled) ? committedUnsettled : 0,
    reservedIRR: Number.isFinite(reserved) ? reserved : 0,
    syncStatus: payload.sync_status?.trim() || null,
  };
}

async function loadSubscriptionSummary(
  tenantSlug: string,
  userID: string,
): Promise<{
  planName: string | null;
  status: string | null;
  periodEnd: string | null;
}> {
  const response = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/subscription",
    userID,
  });

  if (!response.ok) {
    return { planName: null, status: null, periodEnd: null };
  }

  const payload = response.payload as TenantSubscription;
  return {
    planName: payload.plan_display_name?.trim() || payload.plan_slug?.trim() || null,
    status: payload.status?.trim() || null,
    periodEnd: payload.trial_ends_at?.trim() || payload.current_period_end?.trim() || null,
  };
}

async function loadCurrentPermissions(tenantSlug: string, userID: string): Promise<string[]> {
  try {
    const payload = await getCurrentMePermissions(tenantSlug, userID);
    return (payload.permissions ?? []).map((item) => item.trim().toLowerCase());
  } catch {
    return [];
  }
}
