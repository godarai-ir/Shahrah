import AdminShell from "@/app/components/admin-shell";
import UsersAccessManagement from "@/app/components/users-access-management";
import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";

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
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, tenantSlug)}/guides/user-management`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["user:list", "user:manage"]);

  const permissionsPayload = await getCurrentMePermissions(tenantSlug, auth.user.id);
  const permissionSet = getNormalizedPermissionSet(permissionsPayload);
  const canManageUsers = permissionSet.has("user:manage");

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
        currentUserID={auth.user.id}
        canManageUsers={canManageUsers}
        docsHref={docsHref}
      />
    </AdminShell>
  );
}

function buildDocsBaseHref(headersList: Headers, tenantSlug: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const docsHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return docsHost ? `${proto}://${docsHost}/docs` : "/docs";
}
