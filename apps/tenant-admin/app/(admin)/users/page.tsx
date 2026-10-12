import AdminShell from "@/features/admin-shell/AdminShell";
import UsersAccessManagement from "@/features/users/UsersAccessManagement";
import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { parseUserListPayload } from "@/features/users/api";
import { mePermissionsResponseSchema } from "@/lib/contracts/roles";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

const USERS_FEATURE_KEY = "users.max";
/** Matches the roster the client filters over; the control plane caps a page at 200. */
const ROSTER_FETCH_LIMIT = 200;

export default async function UsersPage() {
  const auth = await requireAuth("/login?next=/users");
  const t = await getTranslations("accessManagement");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const docsHref = `${buildDocsBaseHref(requestHeaders, tenantSlug)}/guides/user-management`;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["user:list", "user:manage"]);

  const permissionsPayload = await getCurrentMePermissions(tenantSlug, auth.user.id);
  const permissionSet = getNormalizedPermissionSet(permissionsPayload);
  const canManageUsers = permissionSet.has("user:manage");
  const parsedCatalog = mePermissionsResponseSchema.safeParse(permissionsPayload);
  const initialPermissionCatalog = parsedCatalog.success ? parsedCatalog.data.catalog ?? [] : [];

  const [usersResult, entitlementsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v1/users?limit=${ROSTER_FETCH_LIMIT}&offset=0`,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/entitlements",
    }),
  ]);

  const payload = parseUserListPayload(usersResult.payload);
  const initialUsers = usersResult.ok ? payload.data ?? [] : [];
  const initialTotal = usersResult.ok ? payload.total ?? initialUsers.length : 0;
  const initialError = usersResult.ok ? null : usersResult.errorMessage ?? "Request failed.";

  // Plan usage belongs in the page header, not a tooltip beside the invite button.
  const usersLimit = entitlementsResult.ok ? readUsersLimit(entitlementsResult.payload) : null;
  const description = usersLimit === null
    ? t("usersPage.description")
    : t("usersPage.headerUsage", { used: initialTotal, limit: usersLimit });

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["accessManagement"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("usersPage.title"), description, helpHref: docsHref }}
    >
      <UsersAccessManagement
        initialUsers={initialUsers}
        initialTotal={initialTotal}
        initialError={initialError}
        initialPermissionCatalog={initialPermissionCatalog}
        currentUserID={auth.user.id}
        canManageUsers={canManageUsers}
      />
    </AdminShell>
  );
}

function readUsersLimit(payload: unknown): number | null {
  if (!payload || typeof payload !== "object") return null;
  const features = (payload as { features?: Record<string, unknown> }).features;
  const value = features?.[USERS_FEATURE_KEY];
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}
