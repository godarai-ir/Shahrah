import AdminShell from "@/app/components/admin-shell";
import TeamManagement from "@/app/components/team-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";

type Team = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  org_unit_id?: string | null;
  workspace_type?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

type TeamListResponse = {
  data?: Team[];
  total?: number;
};

type OrgUnit = {
  id: string;
  alias: string;
  display_name?: string;
};

type OrgUnitListResponse = {
  data?: OrgUnit[];
};
export const dynamic = "force-dynamic";

export default async function TeamsPage() {
  const auth = await requireAuth("/login?next=/teams");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);
  const teamsDocsHref = `${docsBaseHref}/guides/team-management`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["team:read", "team:manage"]);

  const [teamsResult, orgUnitsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/teams?limit=25&offset=0",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/org-units?limit=200&offset=0",
      userID: auth.user.id,
    }),
  ]);

  const teamsPayload = (teamsResult.payload ?? {}) as TeamListResponse;
  const orgUnitsPayload = (orgUnitsResult.payload ?? {}) as OrgUnitListResponse;
  const initialTeams = teamsResult.ok ? teamsPayload.data ?? [] : [];
  const initialTotal = teamsResult.ok ? teamsPayload.total ?? initialTeams.length : 0;
  const initialError = teamsResult.ok ? null : teamsResult.errorMessage ?? "Request failed.";
  const initialOrgUnits = orgUnitsResult.ok ? orgUnitsPayload.data ?? [] : [];

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <TeamManagement
        initialTeams={initialTeams}
        initialTotal={initialTotal}
        initialError={initialError}
        initialOrgUnits={initialOrgUnits}
        currentUserRole={auth.user.role}
        docsHref={teamsDocsHref}
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
