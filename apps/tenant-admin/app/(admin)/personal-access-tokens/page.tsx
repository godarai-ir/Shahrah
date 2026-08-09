import AdminShell from "@/features/admin-shell/AdminShell";
import PersonalAccessTokenManagement from "@/features/personal-access-tokens/PersonalAccessTokenManagement";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { parsePersonalAccessTokenListPayload } from "@/features/personal-access-tokens/api";

export const dynamic = "force-dynamic";

export default async function PersonalAccessTokensPage() {
  const auth = await requireAuth("/login?next=/personal-access-tokens");
  const t = await getTranslations("personalAccessTokens");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, tenantSlug)}/guides/api-keys`;

  const patsResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/personal-access-tokens?limit=25&offset=0",
    userID: auth.user.id,
  });

  const patPayload = parsePersonalAccessTokenListPayload(patsResult.payload);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["personalAccessTokens"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("create.description"), helpHref: docsHref }}
    >
      <PersonalAccessTokenManagement
        initialTokens={patsResult.ok ? patPayload.data ?? [] : []}
        initialError={patsResult.ok ? null : patsResult.errorMessage ?? "Request failed."}
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
