import { getTranslations } from "next-intl/server";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import AdminShell from "@/features/admin-shell/AdminShell";
import { parseSeatRenewalPreview, parseSeatSummary, type SeatSummary } from "@/features/billing/seats/api";
import { renewalWindowOpen } from "@/features/billing/seats/domain";
import { SEAT_READ_PERMISSIONS } from "@/features/billing/seats/permissions";
import { RenewSeatsForm } from "@/features/billing/seats/RenewSeatsForm";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { canAnyPermission } from "@/lib/authz";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type PageProps = {
  searchParams: Promise<{ payment_status?: string }>;
};

/** The request time, read outside the component so rendering stays pure. */
function currentTime(): number {
  return Date.now();
}

/** The preview query that renews every seat held. */
function renewEverything(summary: SeatSummary): string {
  const query = new URLSearchParams();
  for (const item of summary.paid) query.set(item.type, String(item.purchased));

  return query.toString();
}

/** Renews the paid seats: how many stay, who loses a seat because of it, and the payment. */
export default async function RenewSeatsPage({ searchParams }: PageProps) {
  const auth = await requireAuth("/login?next=/billing/seats/renew");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, SEAT_READ_PERMISSIONS);

  const t = await getTranslations("billingSeats");
  const { payment_status: paymentStatus } = await searchParams;
  const request = (path: string) => fetchControlPlaneJSON({ tenantSlug, method: "GET", path, userID: auth.user.id });
  const [summaryResponse, permissionsPayload] = await Promise.all([
    request("/api/v1/billing/seats"),
    getCurrentMePermissions(tenantSlug, auth.user.id).catch(() => ({ permissions: [] as string[] })),
  ]);
  const permissions = (permissionsPayload.permissions ?? []).map((item) => item.trim().toLowerCase());
  const summary = summaryResponse.ok ? parseSeatSummary(summaryResponse.payload) : null;

  const nowMs = currentTime();
  const windowOpen = summary ? renewalWindowOpen(summary, nowMs) : false;

  // Counts are chosen when renewing, so the page starts from renewing every seat.
  // Before renewal opens, or once everything is paid, there is nothing to preview.
  const previewResponse =
    summary && windowOpen && summary.renewal_due ? await request(`/api/v1/billing/seats/renewal/preview?${renewEverything(summary)}`) : null;
  const preview = previewResponse?.ok ? parseSeatRenewalPreview(previewResponse.payload) : null;

  return (
    <AdminShell
      messageNamespaces={["billing", "billingSeats"]}
      tenantSlug={tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("renew.title"), description: t("renew.pageDescription") }}
    >
      {summary ? (
        <RenewSeatsForm
          summary={summary}
          preview={preview}
          windowOpen={windowOpen}
          nowMs={nowMs}
          canManage={canAnyPermission(permissions, ["billing:manage"])}
          paymentResult={paymentStatus === "success" || paymentStatus === "failed" ? paymentStatus : null}
        />
      ) : (
        <Alert variant="destructive">
          <AlertTitle>{t("errors.title")}</AlertTitle>
          <AlertDescription>{summaryResponse.errorMessage ?? t("errors.loadSummary")}</AlertDescription>
        </Alert>
      )}
    </AdminShell>
  );
}
