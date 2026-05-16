import AdminShell from "@/app/components/admin-shell";
import RolesPermissionsManagement from "@/app/components/roles-permissions-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type RoleDefinition = {
  id: string;
  tenant_id: string;
  name: string;
  display_name: string;
  description?: string;
  is_system: boolean;
  permissions: string[];
  created_at: string;
  updated_at: string;
};

type RoleListResponse = {
  data?: RoleDefinition[];
};

export const dynamic = "force-dynamic";

export default async function RolesPage() {
  const auth = await requireAuth("/login?next=/roles");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["role:list"]);

  const rolesResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/roles?limit=200&offset=0",
  });

  const payload = (rolesResult.payload ?? {}) as RoleListResponse;
  const initialRoles = rolesResult.ok ? payload.data ?? [] : [];
  const initialError = rolesResult.ok ? null : rolesResult.errorMessage ?? "Request failed.";

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <RolesPermissionsManagement initialRoles={initialRoles} initialError={initialError} />
    </AdminShell>
  );
}
