import AdminShell from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import RolesPermissionsManagement from "@/app/components/roles-permissions-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
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

type MePermissionsResponse = {
  catalog?: Array<{ key: string; group: string; description?: string }>;
};

export const dynamic = "force-dynamic";

export default async function RolesPage() {
  const auth = await requireAuth("/login?next=/roles");
  const t = await getTranslations("accessManagement");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["role:list", "role:manage"]);

  const rolesResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/roles?limit=200&offset=0",
  });

  const payload = (rolesResult.payload ?? {}) as RoleListResponse;
  const initialRoles = rolesResult.ok ? payload.data ?? [] : [];
  const initialError = rolesResult.ok ? null : rolesResult.errorMessage ?? "Request failed.";
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
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("rolesPage.listTitle"), description: t("rolesPage.description") }}
    >
      <RolesPermissionsManagement
        initialRoles={initialRoles}
        initialError={initialError}
        initialPermissionCatalog={initialPermissionCatalog}
      />
    </AdminShell>
  );
}
