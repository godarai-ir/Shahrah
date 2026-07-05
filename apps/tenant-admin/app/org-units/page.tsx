import AdminShell from "@/app/components/admin-shell";
import OrgUnitManagement from "@/app/components/org-unit-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";

type OrgUnit = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

type OrgUnitListResponse = {
  data?: OrgUnit[];
  total?: number;
  limit?: number;
  offset?: number;
};

export const dynamic = "force-dynamic";

export default async function OrgUnitsPage() {
  const auth = await requireAuth("/login?next=/org-units");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);
  const orgUnitsDocsHref = `${docsBaseHref}/guides/tenant-team-account-model`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["org_unit:list", "org_unit:read", "org_unit:create", "org_unit:manage"]);

  const orgUnitsResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/org-units?limit=25&offset=0",
    userID: auth.user.id,
  });

  const payload = (orgUnitsResult.payload ?? {}) as OrgUnitListResponse;
  const initialOrgUnits = orgUnitsResult.ok ? payload.data ?? [] : [];
  const initialTotal = orgUnitsResult.ok ? payload.total ?? initialOrgUnits.length : 0;
  const initialError = orgUnitsResult.ok ? null : orgUnitsResult.errorMessage ?? "Request failed.";

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <OrgUnitManagement
        initialOrgUnits={initialOrgUnits}
        initialTotal={initialTotal}
        initialError={initialError}
        currentUserRole={auth.user.role}
        docsHref={orgUnitsDocsHref}
      />
    </AdminShell>
  );
}

function buildDocsBaseHref(headersList: Headers, tenantSlug: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const docsHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return docsHost ? `${proto}://${docsHost}/docs` : "/docs";
}
