import AdminShell from "@/features/admin-shell/AdminShell";
import TeamsManagement from "@/features/teams-management/TeamsManagement";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { parseTeamListPayload } from "@/lib/api/teams";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

export default async function TeamsPage() {
  const auth = await requireAuth("/login?next=/teams");
  const t = await getTranslations("teamsPage");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);
  const docsHref = `${docsBaseHref}/guides/tenant-team-account-model`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["team:read", "team:manage"]);

  const teamsResult = await fetchControlPlaneJSON({ tenantSlug, method: "GET", path: "/api/v1/teams?limit=200&offset=0", userID: auth.user.id });

  const teamsPayload = parseTeamListPayload(teamsResult.payload);
  const initialTeams = teamsResult.ok ? teamsPayload.data ?? [] : [];
  const initialError = teamsResult.ok ? null : teamsResult.errorMessage ?? "Request failed.";

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["teamsPage", "teams"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle"), helpHref: docsHref }}
    >
      <TeamsManagement
        initialTeams={initialTeams}
        initialError={initialError}
        currentUserRole={auth.user.role}
      />
    </AdminShell>
  );
}
