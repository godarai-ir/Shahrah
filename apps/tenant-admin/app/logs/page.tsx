import { getTranslations } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import LogsCommandCenter from "@/app/components/logs-command-center";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export default async function LogsPage() {
  const auth = await requireAuth("/login?next=/logs");
  const tenantSlug = auth.tenantSlug;
  const t = await getTranslations("logs");
  await requirePermissionOr403(tenantSlug, auth.user.id, ["logs:read_metadata", "logs:read_content"]);

  const [logsResult, summaryResult, facetsResult, viewsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v2/logs?limit=25&offset=0&sort_by=started_at&sort_order=desc",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v2/logs/summary",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v2/logs/facets",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v2/logs/saved-views",
      userID: auth.user.id,
    }),
  ]);

  const currentUser = { id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role };

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle"), helpHref: "/docs/observability-logs" }}
    >
      <LogsCommandCenter
        initialLogs={logsResult.ok ? logsResult.payload : undefined}
        initialSummary={summaryResult.ok ? summaryResult.payload : undefined}
        initialFacets={facetsResult.ok ? facetsResult.payload : undefined}
        initialSavedViews={viewsResult.ok ? viewsResult.payload : undefined}
        initialError={!logsResult.ok ? logsResult.errorMessage ?? t("loadError") : undefined}
      />
    </AdminShell>
  );
}
