import AdminShell from "@/app/components/admin-shell";
import TeamManagement from "@/app/components/team-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

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
type AllowedModelListResponse = {
  data?: Array<{ id: string }>;
};

export const dynamic = "force-dynamic";

export default async function TeamsPage() {
  const auth = await requireAuth("/login?next=/teams");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["team:read", "team:manage"]);

  const [teamsResult, orgUnitsResult, allowedModelsResult] = await Promise.all([
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
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/models/allowed?key_type=vat",
      userID: auth.user.id,
    }),
  ]);

  const teamsPayload = (teamsResult.payload ?? {}) as TeamListResponse;
  const orgUnitsPayload = (orgUnitsResult.payload ?? {}) as OrgUnitListResponse;
  const allowedModelsPayload = (allowedModelsResult.payload ?? {}) as AllowedModelListResponse;

  const initialTeams = teamsResult.ok ? teamsPayload.data ?? [] : [];
  const initialTotal = teamsResult.ok ? teamsPayload.total ?? initialTeams.length : 0;
  const initialError = teamsResult.ok ? null : teamsResult.errorMessage ?? "Request failed.";
  const initialOrgUnits = orgUnitsResult.ok ? orgUnitsPayload.data ?? [] : [];
  const initialAllowedModels = allowedModelsResult.ok
    ? (allowedModelsPayload.data ?? []).map((item) => item.id)
    : [];

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
        initialAllowedModels={initialAllowedModels}
        currentUserRole={auth.user.role}
      />
    </AdminShell>
  );
}
