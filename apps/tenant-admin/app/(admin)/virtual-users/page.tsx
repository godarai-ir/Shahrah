import AdminShell from "@/features/admin-shell/AdminShell";
import VirtualUserTokenManagement from "@/features/virtual-users/VirtualUserTokenManagement";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { parseVirtualUserListPayload } from "@/features/virtual-users/api";
import { parseTeamListPayload } from "@/lib/api/teams";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

export default async function VirtualUsersPage() {
  const auth = await requireAuth("/login?next=/virtual-users");
  const t = await getTranslations("virtualUsers");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);
  const virtualUsersDocsHref = `${docsBaseHref}/guides/virtual-account-management`;
  await requirePermissionOr403(tenantSlug, auth.user.id, [
    "virtual_account:read",
    "virtual_account:manage",
    "virtual_account:create",
    "team:read",
    "team:manage",
  ]);

  const [virtualUsersResult, teamsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/virtual-users?limit=25&offset=0",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/teams?limit=200&offset=0",
      userID: auth.user.id,
    }),
  ]);

  const virtualUsersPayload = parseVirtualUserListPayload(virtualUsersResult.payload);
  const teamsPayload = parseTeamListPayload(teamsResult.payload);
  const initialVirtualUsers = virtualUsersResult.ok ? virtualUsersPayload.data ?? [] : [];
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["virtualUsers"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("description"), helpHref: virtualUsersDocsHref }}
    >
      <VirtualUserTokenManagement
        initialVirtualUsers={initialVirtualUsers}
        initialTotal={virtualUsersResult.ok ? virtualUsersPayload.total ?? initialVirtualUsers.length : 0}
        initialError={virtualUsersResult.ok ? null : virtualUsersResult.errorMessage ?? "Request failed."}
        teams={teamsResult.ok ? teamsPayload.data ?? [] : []}
      />
    </AdminShell>
  );
}
