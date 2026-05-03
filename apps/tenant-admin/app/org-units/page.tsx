import AdminShell from "@/app/components/admin-shell";
import OrgUnitManagement from "@/app/components/org-unit-management";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { extractTenantSlugFromHost } from "@/lib/tenant";
import { headers } from "next/headers";

type OrgUnit = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  budget_id?: string;
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

  const incomingHeaders = await headers();
  const incomingHost =
    incomingHeaders.get("x-forwarded-host") ??
    incomingHeaders.get("host") ??
    "";

  const tenantSlug =
    extractTenantSlugFromHost(incomingHost) ||
    process.env.SHAHRAH_DEFAULT_TENANT ||
    "default";

  const orgUnitsResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/org-units?limit=25&offset=0",
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
      />
    </AdminShell>
  );
}
