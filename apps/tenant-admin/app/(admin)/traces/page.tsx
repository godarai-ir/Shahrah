import AdminShell, { buildDocsHref } from "@/features/admin-shell/AdminShell";
import { getTranslations } from "next-intl/server";
import { headers } from "next/headers";
import TraceManagement from "@/features/traces/TraceManagement";
import { loadTracePrivacySettings } from "@/features/settings/server";
import { parseTraceListPayload, parseTracePrivacySettings } from "@/features/traces/api";
import { buildTraceFiltersQuery, createDefaultLocalFilters, traceFiltersFromObserve } from "@/features/traces/trace-utils";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { observeFiltersFromSearchParams, resolveObserveRange } from "@/lib/observe-filters";

export const dynamic = "force-dynamic";

export default async function TracesPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const auth = await requireAuth("/login?next=/traces");
  const t = await getTranslations("traces");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsHref = `${buildDocsHref(requestHeaders, tenantSlug)}observability-traces`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  const params = (await searchParams) ?? {};
  const observe = observeFiltersFromSearchParams(toURLSearchParams(params));
  const filtersQuery = buildTraceFiltersQuery(traceFiltersFromObserve(createDefaultLocalFilters(), observe, resolveObserveRange(observe)));

  const tracesResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: `/api/v2/traces?limit=25&offset=0${filtersQuery ? `&${filtersQuery}` : ""}`,
    userID: auth.user.id,
  });

  const tracePrivacyResult = await loadTracePrivacySettings(tenantSlug, auth.user.id);
  const payload = parseTraceListPayload(tracesResult.payload);
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["traces","settings"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("description"), helpHref: docsHref }}
    >
      <TraceManagement
        initialData={tracesResult.ok ? payload.data ?? [] : []}
        initialTotal={tracesResult.ok ? payload.total ?? 0 : 0}
        initialError={tracesResult.ok ? null : tracesResult.errorMessage ?? "Request failed."}
        initialTracePrivacySettings={tracePrivacyResult.ok ? parseTracePrivacySettings(tracePrivacyResult.payload) : null}
        initialTracePrivacyError={tracePrivacyResult.ok ? null : tracePrivacyResult.errorMessage ?? "Failed to load tracing privacy settings."}
        initialTracePrivacyOpen={params.tracePrivacy === "open"}
        currentUserID={currentUser.id}
        currentUserLabel={currentUser.email ?? (mobileAuthUIEnabled ? currentUser.mobile : undefined) ?? currentUser.id}
      />
    </AdminShell>
  );
}

function toURLSearchParams(params: Record<string, string | string[] | undefined>) {
  const out = new URLSearchParams();
  for (const [key, raw] of Object.entries(params)) {
    const value = Array.isArray(raw) ? raw[0] : raw;
    if (value !== undefined) out.set(key, value);
  }
  return out;
}
