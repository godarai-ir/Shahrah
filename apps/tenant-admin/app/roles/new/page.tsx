import AdminShell from "@/app/components/admin-shell";
import RoleCreateManagement from "@/app/components/roles-create-management";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

type PermissionDefinition = {
  key: string;
  group: string;
  description?: string;
};

type MePermissionsResponse = {
  catalog?: PermissionDefinition[];
};

export const dynamic = "force-dynamic";

export default async function NewRolePage() {
  const auth = await requireAuth("/login?next=/roles/new");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["role:manage"]);

  const permissionsPayload = await getCurrentMePermissions(tenantSlug, auth.user.id) as MePermissionsResponse;
  const initialPermissionCatalog = (permissionsPayload.catalog ?? []).map((item) => ({
    ...item,
    description: item.description ?? "",
  }));

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <RoleCreateManagement initialPermissionCatalog={initialPermissionCatalog} />
    </AdminShell>
  );
}



