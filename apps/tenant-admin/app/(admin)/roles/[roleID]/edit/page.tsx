import { notFound } from "next/navigation";
import { getLocale } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import RoleEditManagement from "@/app/components/roles-edit-management";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
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

type PermissionDefinition = {
  key: string;
  group: string;
  description?: string;
};

type MePermissionsResponse = {
  catalog?: PermissionDefinition[];
};

type PageProps = {
  params: Promise<{
    roleID: string;
  }>;
};

export const dynamic = "force-dynamic";

export default async function EditRolePage({ params }: PageProps) {
  const locale = await getLocale();
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

  const rolesPayload = (rolesResult.payload ?? {}) as RoleListResponse;
  const role = (rolesPayload.data ?? []).find((item) => item.id === roleID);
  if (!role) {
    notFound();
  }

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
    <AdminShell messageNamespaces={["accessManagement"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{
        title: locale.startsWith("fa") ? "ویرایش نقش" : "Edit role",
        description: locale.startsWith("fa")
          ? "مشخصات نقش و سطح دسترسی آن را به‌روزرسانی کنید."
          : "Update the role details and its access level.",
      }}
    >
      <RoleEditManagement initialRole={role} initialPermissionCatalog={initialPermissionCatalog} />
    </AdminShell>
  );
}



