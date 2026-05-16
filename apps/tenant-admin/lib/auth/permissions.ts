import { forbidden } from "next/navigation";

import { fetchControlPlaneJSON } from "@/lib/control-plane";

type MePermissionsResponse = {
  permissions?: string[];
};

export async function requirePermissionOr403(tenantSlug: string, userID: string, required: string[]) {
  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/me/permissions",
    userID: userID.trim() || undefined,
  });
  if (!upstream.ok) {
    forbidden();
  }
  const payload = (upstream.payload ?? {}) as MePermissionsResponse;
  const permissionSet = new Set((payload.permissions ?? []).map((item) => item.trim().toLowerCase()));
  const allowed = required.some((item) => permissionSet.has(item.trim().toLowerCase()));
  if (!allowed) {
    forbidden();
  }
}
