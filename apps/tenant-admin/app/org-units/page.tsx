import AdminShell from "@/app/components/admin-shell";
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

  // Render a placeholder until we create the OrgUnitManagement component
  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <div className="container mx-auto py-8">
        <h1 className="text-2xl font-bold mb-4">Organization Units</h1>
        {initialError ? (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {initialError}
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-muted-foreground">
              Found {initialTotal} organization units
            </p>
            {/* TODO: Create OrgUnitManagement component */}
          </div>
        )}
      </div>
    </AdminShell>
  );
}