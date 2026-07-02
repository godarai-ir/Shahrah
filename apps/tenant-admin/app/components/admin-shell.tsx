import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";
import { headers } from "next/headers";

import AdminShellClient from "@/app/components/admin-shell-client";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type AdminShellProps = {
  children: React.ReactNode;
  tenantSlug: string;
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
};

type WalletPayload = {
  balance_irr?: number;
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
  currentUser,
}: AdminShellProps) {
  const locale = await getLocale();
  const messages = await getMessages();
  const requestHeaders = await headers();
  const docsHref = buildDocsHref(requestHeaders, tenantSlug);
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
      >
        {children}
      </AdminShellClient>
    </NextIntlClientProvider>
  );
}

function buildDocsHref(headersList: Headers, tenantSlug: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const docsHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return docsHost ? `${proto}://${docsHost}/docs/` : "/docs/";
}

async function loadWalletSummary(tenantSlug: string, userID: string): Promise<{ balanceIRR: number | null }> {
  const response = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/wallet",
    userID,
  });

  if (!response.ok) {
    return { balanceIRR: null };
  }

  const payload = response.payload as WalletPayload;
  const balance = Number(payload.balance_irr ?? 0);
  return { balanceIRR: Number.isFinite(balance) ? balance : null };
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
