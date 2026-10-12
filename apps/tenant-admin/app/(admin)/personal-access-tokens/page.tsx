import AdminShell from "@/features/admin-shell/AdminShell";
import AccessKeysManagement from "@/features/access-keys/AccessKeysManagement";
import { parseTenantKeyListPayload } from "@/lib/api/access-keys";
import { parseUserListPayload } from "@/features/users/api";
import { parseVirtualUserListPayload } from "@/features/virtual-users/api";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

export default async function AccessKeysPage() {
  const auth = await requireAuth("/login?next=/personal-access-tokens");
  const t = await getTranslations("accessManagement");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, tenantSlug)}/guides/api-keys`;

  // The tenant-wide key list needs team admin or higher. Everyone else still owns
  // personal keys, so fall back to their own. The member list gives the owner
  // column its names and may fail on its own without taking the page down.
  const [tenantKeysResult, membersResult, virtualUsersResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/keys",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/users?limit=200&offset=0",
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/virtual-users?limit=200&offset=0",
      userID: auth.user.id,
    }),
  ]);

  const keysResult = tenantKeysResult.ok
    ? tenantKeysResult
    : await fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/personal-access-tokens?limit=200&offset=0",
        userID: auth.user.id,
      });
  const isTenantWide = tenantKeysResult.ok;

  const keysPayload = parseTenantKeyListPayload(keysResult.payload);
  const membersPayload = parseUserListPayload(membersResult.payload);
  const virtualUsersPayload = parseVirtualUserListPayload(virtualUsersResult.payload);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["accessManagement", "personalAccessTokens"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("keys.title"), description: isTenantWide ? t("keys.description") : t("keys.ownDescription"), helpHref: docsHref }}
    >
      <AccessKeysManagement
        initialKeys={keysResult.ok ? keysPayload.data ?? [] : []}
        initialError={keysResult.ok ? null : keysResult.errorMessage ?? "Request failed."}
        members={membersResult.ok ? membersPayload.data ?? [] : []}
        virtualUsers={virtualUsersResult.ok ? virtualUsersPayload.data ?? [] : []}
        currentUserID={auth.user.id}
        isTenantWide={isTenantWide}
      />
    </AdminShell>
  );
}
