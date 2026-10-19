import AdminShell from "@/features/admin-shell/AdminShell";
import OrgStructure from "@/features/org-structure/OrgStructure";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { parseTeamListPayload } from "@/lib/api/teams";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

export default async function OrgStructurePage() {
  const auth = await requireAuth("/login?next=/org-structure");
  const t = await getTranslations("orgStructure");
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
    <AdminShell messageNamespaces={["orgStructure", "teams"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle"), helpHref: docsHref }}
    >
      <OrgStructure
        initialTeams={initialTeams}
        initialError={initialError}
        currentUserRole={auth.user.role}
      />
    </AdminShell>
  );
}
