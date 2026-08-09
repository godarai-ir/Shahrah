import AdminShell from "@/features/admin-shell/AdminShell";
import TeamManagement from "@/features/teams/TeamManagement";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { parseTeamListPayload } from "@/features/teams/api";
import { parseOrgUnitListPayload } from "@/features/org-units/api";
export const dynamic = "force-dynamic";

export default async function TeamsPage() {
  const auth = await requireAuth("/login?next=/teams");
  const t = await getTranslations("teams");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);
  const teamsDocsHref = `${docsBaseHref}/guides/team-management`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["team:read", "team:manage"]);

  const [teamsResult, orgUnitsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/teams?limit=25&offset=0",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/org-units?limit=200&offset=0",
      userID: auth.user.id,
    }),
  ]);

  const teamsPayload = parseTeamListPayload(teamsResult.payload);
  const orgUnitsPayload = parseOrgUnitListPayload(orgUnitsResult.payload);
  const initialTeams = teamsResult.ok ? teamsPayload.data ?? [] : [];
  const initialTotal = teamsResult.ok ? teamsPayload.total ?? initialTeams.length : 0;
  const initialError = teamsResult.ok ? null : teamsResult.errorMessage ?? "Request failed.";
  const initialOrgUnits = orgUnitsResult.ok ? orgUnitsPayload.data ?? [] : [];

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["teams"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle"), helpHref: teamsDocsHref }}
    >
      <TeamManagement
        initialTeams={initialTeams}
        initialTotal={initialTotal}
        initialError={initialError}
        initialOrgUnits={initialOrgUnits}
        currentUserRole={auth.user.role}
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
