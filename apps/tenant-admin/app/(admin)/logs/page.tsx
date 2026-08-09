import { getTranslations } from "next-intl/server";
import { headers } from "next/headers";

import AdminShell, { buildDocsHref } from "@/app/components/admin-shell";
import LogsCommandCenter from "@/app/components/logs-command-center";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type LogsPageProps = {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
};

export default async function LogsPage({ searchParams }: LogsPageProps) {
  const incomingSearchParams = (await searchParams) ?? {};
  const auth = await requireAuth("/login?next=/logs");
  const tenantSlug = auth.tenantSlug;
  const t = await getTranslations("logs");
  const requestHeaders = await headers();
  const docsHref = `${buildDocsHref(requestHeaders, tenantSlug)}observability-logs`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["logs:read_metadata", "logs:read_content"]);

  const initialQuery = buildInitialLogQuery(incomingSearchParams);
  const querySuffix = initialQuery ? `&${initialQuery}` : "";
  const summarySuffix = initialQuery ? `?${initialQuery}` : "";
  const [logsResult, summaryResult, facetsResult, viewsResult, timeseriesResult, breakdownResult, rangeStatsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/logs?limit=25&offset=0&sort_by=started_at&sort_order=desc${querySuffix}`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/logs/summary${summarySuffix}`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/logs/facets${summarySuffix}`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v2/logs/saved-views",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/logs/analytics?type=timeseries&group_by=&interval_seconds=3600${querySuffix}`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/logs/analytics?type=breakdown&group_by=provider${querySuffix}`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v2/logs/range-stats${summarySuffix}`,
      userID: auth.user.id,
    }),
  ]);

  const currentUser = { id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role };

  return (
    <AdminShell messageNamespaces={["logs","settings","providerAccounts"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle"), helpHref: docsHref }}
    >
      <LogsCommandCenter
        initialLogs={logsResult.ok ? logsResult.payload : undefined}
        initialSummary={summaryResult.ok ? summaryResult.payload : undefined}
        initialFacets={facetsResult.ok ? facetsResult.payload : undefined}
        initialSavedViews={viewsResult.ok ? viewsResult.payload : undefined}
        initialTimeseries={timeseriesResult.ok ? timeseriesResult.payload : undefined}
        initialBreakdown={breakdownResult.ok ? breakdownResult.payload : undefined}
        initialRangeStats={rangeStatsResult.ok ? rangeStatsResult.payload : undefined}
        initialError={!logsResult.ok ? logsResult.errorMessage ?? t("loadError") : undefined}
      />
    </AdminShell>
  );
}

function buildInitialLogQuery(searchParams: Record<string, string | string[] | undefined>): string {
  const params = new URLSearchParams();
  for (const key of ["provider", "model"] as const) {
    const raw = searchParams[key];
    const value = (Array.isArray(raw) ? raw[0] : raw)?.trim();
    if (value) params.set(key, value);
  }
  return params.toString();
}
