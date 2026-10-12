import AdminShell, { buildDocsHref } from "@/features/admin-shell/AdminShell";
import { getTranslations } from "next-intl/server";
import { headers } from "next/headers";
import RolesPermissionsManagement from "@/features/roles/RolesPermissionsManagement";
import { mePermissionsResponseSchema, roleListSchema } from "@/lib/contracts/roles";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export default async function RolesPage() {
  const auth = await requireAuth("/login?next=/roles");
  const t = await getTranslations("accessManagement");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const rolesDocsHref = `${buildDocsHref(requestHeaders, tenantSlug)}guides/roles-and-permissions`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["role:list", "role:manage"]);

  const rolesResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/roles?limit=200&offset=0",
  });

  const parsedRoles = roleListSchema.safeParse(rolesResult.payload);
  const initialRoles = rolesResult.ok && parsedRoles.success ? parsedRoles.data.data ?? [] : [];
  const initialError = rolesResult.ok ? null : rolesResult.errorMessage ?? "Request failed.";
  const permissionsPayload = mePermissionsResponseSchema.safeParse(
    await getCurrentMePermissions(tenantSlug, auth.user.id),
  );
  const initialPermissionCatalog = permissionsPayload.success ? permissionsPayload.data.catalog ?? [] : [];
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["accessManagement"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("rolesPage.listTitle"), description: t("rolesPage.description"), helpHref: rolesDocsHref }}
    >
      <RolesPermissionsManagement
        initialRoles={initialRoles}
        initialError={initialError}
        initialPermissionCatalog={initialPermissionCatalog}
      />
    </AdminShell>
  );
}
