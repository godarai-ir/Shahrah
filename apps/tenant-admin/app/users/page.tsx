import AdminShell from "@/app/components/admin-shell";
import UsersAccessManagement from "@/app/components/users-access-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type TenantAdminUser = {
  id: string;
  tenant_slug: string;
  email?: string;
  mobile?: string;
  role: string;
  active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
};

type TenantAdminUserListResponse = {
  data?: TenantAdminUser[];
  total?: number;
  limit?: number;
  offset?: number;
};

export const dynamic = "force-dynamic";

export default async function UsersPage() {
  const auth = await requireAuth("/login?next=/users");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["user:list"]);

  const usersResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/users?limit=25&offset=0",
  });

  const payload = (usersResult.payload ?? {}) as TenantAdminUserListResponse;
  const initialUsers = usersResult.ok ? payload.data ?? [] : [];
  const initialTotal = usersResult.ok ? payload.total ?? initialUsers.length : 0;
  const initialError = usersResult.ok ? null : usersResult.errorMessage ?? "Request failed.";

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <UsersAccessManagement
        initialUsers={initialUsers}
        initialTotal={initialTotal}
        initialError={initialError}
      />
    </AdminShell>
  );
}
