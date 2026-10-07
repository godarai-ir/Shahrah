import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import OIDCClientManagement from "@/features/oidc-clients/OIDCClientManagement";
import { parseOIDCClientListPayload } from "@/features/oidc-clients/api";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { resolveRequestOriginFromHeaders } from "@/lib/tenant";
import { getTenantApiBaseURLForOrigin } from "@/lib/tenant-api-origin";

export const dynamic = "force-dynamic";

export default async function OIDCClientsPage() {
  const auth = await requireAuth("/login?next=/oidc-clients");
  const t = await getTranslations("oidcClients");
  const requestHeaders = await headers();

  // The control plane serves each tenant's issuer and gateway from the tenant API origin.
  const issuerURL = getTenantApiBaseURLForOrigin(resolveRequestOriginFromHeaders(requestHeaders, "http://localhost"));

  const result = await fetchControlPlaneJSON({
    tenantSlug: auth.tenantSlug,
    method: "GET",
    path: "/api/v1/oidc-clients",
    userID: auth.user.id,
  });
  const initialError = result.ok ? null : result.status === 403 ? "forbidden" : "unavailable";

  return (
    <AdminShell
      messageNamespaces={["oidcClients"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("title"), description: t("description") }}
    >
      <OIDCClientManagement
        initialClients={result.ok ? parseOIDCClientListPayload(result.payload) : []}
        initialError={initialError}
        issuerURL={issuerURL}
      />
    </AdminShell>
  );
}
