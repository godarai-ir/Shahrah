import { z } from "zod";

import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

const tenantSubscriptionSchema = z.object({
  plan_slug: z.string().optional(),
  plan_display_name: z.string().optional(),
  status: z.string().optional(),
  current_period_end: z.string().optional(),
  trial_ends_at: z.string().optional(),
});

export async function loadAdminShellData(tenantSlug: string, userID: string) {
  const [permissions, subscriptionSummary] = await Promise.all([
    loadCurrentPermissions(tenantSlug, userID),
    loadSubscriptionSummary(tenantSlug, userID),
  ]);

  return { permissions, subscriptionSummary };
}

export function getAppVersion(): string {
  return process.env.NEXT_PUBLIC_APP_VERSION?.replace(/^v/, "") || "0.0.0-dev";
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
