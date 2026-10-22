import { getTranslations } from "next-intl/server";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import AdminShell from "@/features/admin-shell/AdminShell";
import { parseSeatSummary } from "@/features/billing/seats/api";
import { BuySeatsForm, type BuySeatsTarget } from "@/features/billing/seats/BuySeatsForm";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { seatMemberListSchema } from "@/lib/seats";

export const dynamic = "force-dynamic";

type PageProps = {
  searchParams: Promise<{ seat?: string; assign?: string; request?: string }>;
};

/**
 * Checkout for paid seats. The seats page opens it plainly; the users list and
 * an approved upgrade request add `assign` (and `request`) to buy one seat for a member.
 */
export default async function BuySeatsPage({ searchParams }: PageProps) {
  const auth = await requireAuth("/login?next=/billing/seats/buy");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["billing:manage"]);

  const t = await getTranslations("billingSeats");
  const { seat, assign, request: requestID } = await searchParams;
  const request = (path: string) => fetchControlPlaneJSON({ tenantSlug, method: "GET", path, userID: auth.user.id });
  const [summaryResponse, target] = await Promise.all([
    request("/api/v1/billing/seats"),
    resolveTarget(seat === "max" ? "max" : "pro", assign, requestID, request),
  ]);
  const summary = summaryResponse.ok ? parseSeatSummary(summaryResponse.payload) : null;

  return (
    <AdminShell
      messageNamespaces={["billing", "billingSeats"]}
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("buy.title"), description: t("buy.pageDescription") }}
    >
      {summary ? (
        <BuySeatsForm target={target} offers={summary.offers} />
      ) : (
        <Alert variant="destructive">
          <AlertTitle>{t("errors.title")}</AlertTitle>
          <AlertDescription>{summaryResponse.errorMessage ?? t("errors.loadSummary")}</AlertDescription>
        </Alert>
      )}
    </AdminShell>
  );
}

/** Only the member's name is needed, so the roster is read for that single lookup. */
async function resolveTarget(
  seat: "pro" | "max",
  assign: string | undefined,
  requestID: string | undefined,
  request: (path: string) => ReturnType<typeof fetchControlPlaneJSON>,
): Promise<BuySeatsTarget> {
  if (!assign) return { seat };

  const response = await request("/api/v1/billing/seats/members?limit=500");
  const parsed = response.ok ? seatMemberListSchema.safeParse(response.payload) : null;
  const member = parsed?.success ? parsed.data.items.find((item) => item.user_id === assign) : undefined;
  if (!member) return { seat };

  return { seat, assignUserID: member.user_id, assignLabel: member.name || member.email, requestID };
}
