import { z } from "zod";

import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

const walletPayloadSchema = z.object({
  balance_irr: z.number().optional(),
  ledger_balance_irr: z.number().optional(),
  available_balance_irr: z.number().optional(),
  reserved_irr: z.number().optional(),
  committed_unsettled_irr: z.number().optional(),
  pending_irr: z.number().optional(),
  sync_status: z.string().optional(),
});

const tenantSubscriptionSchema = z.object({
  plan_slug: z.string().optional(),
  plan_display_name: z.string().optional(),
  status: z.string().optional(),
  current_period_end: z.string().optional(),
  trial_ends_at: z.string().optional(),
});

export async function loadAdminShellData(tenantSlug: string, userID: string) {
  const [permissions, walletSummary, subscriptionSummary] = await Promise.all([
    loadCurrentPermissions(tenantSlug, userID),
    loadWalletSummary(tenantSlug, userID),
    loadSubscriptionSummary(tenantSlug, userID),
  ]);

  return { permissions, walletSummary, subscriptionSummary };
}

export function getAppVersion(): string {
  return process.env.NEXT_PUBLIC_APP_VERSION?.replace(/^v/, "") || "0.0.0-dev";
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

  const parsedPayload = walletPayloadSchema.safeParse(response.payload);
  if (!parsedPayload.success) {
    return {
      availableBalanceIRR: null,
      ledgerBalanceIRR: null,
      committedUnsettledIRR: 0,
      reservedIRR: 0,
      syncStatus: null,
    };
  }
  const payload = parsedPayload.data;
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

  if (!response.ok) return { planName: null, status: null, periodEnd: null };

  const parsedPayload = tenantSubscriptionSchema.safeParse(response.payload);
  if (!parsedPayload.success) return { planName: null, status: null, periodEnd: null };
  const payload = parsedPayload.data;
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
