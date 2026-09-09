import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import RoleEditManagement from "@/features/roles/RolesEditManagement";
import { mePermissionsResponseSchema, roleListSchema } from "@/features/roles/contracts";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type PageProps = {
  params: Promise<{
    roleID: string;
  }>;
};

export const dynamic = "force-dynamic";

export default async function EditRolePage({ params }: PageProps) {
  const t = await getTranslations("accessManagement");
  const { roleID } = await params;
  const auth = await requireAuth(`/login?next=/roles/${encodeURIComponent(roleID)}/edit`);
  const tenantSlug = auth.tenantSlug;

  await requirePermissionOr403(tenantSlug, auth.user.id, ["role:manage"]);

  const rolesResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/roles?limit=200&offset=0",
  });

  if (!rolesResult.ok) {
    notFound();
  }

  const rolesPayload = roleListSchema.safeParse(rolesResult.payload);
  const role = rolesPayload.success
    ? (rolesPayload.data.data ?? []).find((item) => item.id === roleID)
    : undefined;
  if (!role) {
    notFound();
  }

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
      pageHeader={{
        title: t("rolesPage.editTitle"),
        description: t("rolesPage.editDescription"),
      }}
    >
      <RoleEditManagement initialRole={role} initialPermissionCatalog={initialPermissionCatalog} />
    </AdminShell>
  );
}



