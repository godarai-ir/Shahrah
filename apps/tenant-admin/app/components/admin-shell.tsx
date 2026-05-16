import AdminShellClient from "@/app/components/admin-shell-client";
import type { MePermissionsResponse } from "@/lib/authz";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type AdminShellProps = {
  children: React.ReactNode;
  tenantSlug: string;
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
};

export default async function AdminShell({
  children,
  tenantSlug,
  currentUser,
}: AdminShellProps) {
  const permissions = await loadCurrentPermissions(tenantSlug, currentUser.id);
  return (
    <AdminShellClient
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      permissions={permissions}
    >
      {children}
    </AdminShellClient>
  );
}

async function loadCurrentPermissions(tenantSlug: string, userID: string): Promise<string[]> {
  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/me/permissions",
    userID,
  });
  if (!upstream.ok) {
    return [];
  }
  const payload = (upstream.payload ?? {}) as MePermissionsResponse;
  return (payload.permissions ?? []).map((item) => item.trim().toLowerCase());
}
