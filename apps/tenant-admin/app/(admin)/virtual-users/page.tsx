import AdminShell from "@/app/components/admin-shell";
import VirtualUserTokenManagement from "@/app/components/virtual-user-token-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

type VirtualUser = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  org_unit_id?: string;
  team_id?: string;
  active: boolean;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

type Team = {
  id: string;
  alias: string;
  display_name?: string;
  org_unit_id?: string | null;
};

type OrgUnit = {
  id: string;
  alias: string;
  display_name?: string;
};

type ListResponse<T> = {
  data?: T[];
  total?: number;
};

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

  const [virtualUsersResult, teamsResult, orgUnitsResult] = await Promise.all([
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
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/org-units?limit=200&offset=0",
      userID: auth.user.id,
    }),
  ]);

  const virtualUsersPayload = (virtualUsersResult.payload ?? {}) as ListResponse<VirtualUser>;
  const teamsPayload = (teamsResult.payload ?? {}) as ListResponse<Team>;
  const orgUnitsPayload = (orgUnitsResult.payload ?? {}) as ListResponse<OrgUnit>;
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
        orgUnits={orgUnitsResult.ok ? orgUnitsPayload.data ?? [] : []}
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
