import AdminShell from "@/features/admin-shell/AdminShell";
import UsersAccessManagement from "@/features/users/UsersAccessManagement";
import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { parseSeatUsage, parseUserListPayload } from "@/features/users/api";
import { mePermissionsResponseSchema } from "@/features/roles/contracts";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

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

  const [usersResult, seatsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v1/users?limit=${ROSTER_FETCH_LIMIT}&offset=0`,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/users/seats",
    }),
  ]);

  const payload = parseUserListPayload(usersResult.payload);
  const initialUsers = usersResult.ok ? payload.data ?? [] : [];
  const initialError = usersResult.ok ? null : usersResult.errorMessage ?? "Request failed.";

  // Plan usage belongs in the page header, not a tooltip beside the invite button.
  // Seats count active members and pending invitations, as the plan cap does.
  const seats = seatsResult.ok ? parseSeatUsage(seatsResult.payload) : null;
  const description = seats?.limit == null
    ? t("usersPage.description")
    : t("usersPage.headerUsage", { used: seats.used, limit: seats.limit });

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
        initialError={initialError}
        initialPermissionCatalog={initialPermissionCatalog}
        currentUserID={auth.user.id}
        canManageUsers={canManageUsers}
      />
    </AdminShell>
  );
}
