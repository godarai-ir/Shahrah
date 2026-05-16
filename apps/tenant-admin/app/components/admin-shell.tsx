import AdminShellClient from "@/app/components/admin-shell-client";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";

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
  try {
    const payload = await getCurrentMePermissions(tenantSlug, userID);
    return (payload.permissions ?? []).map((item) => item.trim().toLowerCase());
  } catch {
    return [];
  }
}
