import AdminShell, { buildDocsHref } from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import { headers } from "next/headers";
import TraceManagement from "@/app/components/trace-management";
import { loadTracePrivacySettings, type TracePrivacySettingsData } from "@/app/settings/_shared";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type RequestTrace = {
  trace_id: string;
  tenant_id: string;
  request_id?: string;
  model?: string;
  provider?: string;
  route: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  error_code?: string;
  sampled: boolean;
  attempts?: number;
};

type TraceListResponse = {
  data?: RequestTrace[];
  total?: number;
};

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

  const tracesResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v2/traces?limit=25&offset=0",
    userID: auth.user.id,
  });

  const tracePrivacyResult = await loadTracePrivacySettings(tenantSlug, auth.user.id);
  const payload = (tracesResult.payload ?? {}) as TraceListResponse;
  const params = (await searchParams) ?? {};
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
        initialTracePrivacySettings={tracePrivacyResult.ok ? (tracePrivacyResult.payload as TracePrivacySettingsData) : null}
        initialTracePrivacyError={tracePrivacyResult.ok ? null : tracePrivacyResult.errorMessage ?? "Failed to load tracing privacy settings."}
        initialTracePrivacyOpen={params.tracePrivacy === "open"}
        currentUserID={currentUser.id}
        currentUserLabel={currentUser.email ?? (mobileAuthUIEnabled ? currentUser.mobile : undefined) ?? currentUser.id}
      />
    </AdminShell>
  );
}
