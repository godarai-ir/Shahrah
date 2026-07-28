import { NextRequest, NextResponse } from "next/server";

import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { readAuthFromRequest } from "@/lib/auth/request-auth";

export const PLAYGROUND_VIDEO_READ_PERMISSIONS = [
  "playground:read",
  "playground:run",
  "policy:list",
  "policy:manage",
] as const;

export const PLAYGROUND_VIDEO_MUTATE_PERMISSIONS = [
  "playground:run",
  "policy:manage",
] as const;

export async function authorizeAPIRequest(
  request: NextRequest,
  requiredPermissions: readonly string[],
) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return {
      ok: false as const,
      response: NextResponse.json(
        { error: { message: auth.message } },
        { status: auth.status },
      ),
    };
  }

  const permissionPayload = await getCurrentMePermissions(
    auth.tenantSlug,
    auth.user.id,
  ).catch(() => null);
  const permissions = permissionPayload
    ? getNormalizedPermissionSet(permissionPayload)
    : new Set<string>();
  const allowed = requiredPermissions.some((permission) =>
    permissions.has(permission.trim().toLowerCase()),
  );
  if (!allowed) {
    return {
      ok: false as const,
      response: NextResponse.json(
        { error: { message: "You do not have permission to access this resource." } },
        { status: 403 },
      ),
    };
  }

  return { ok: true as const, auth };
}
