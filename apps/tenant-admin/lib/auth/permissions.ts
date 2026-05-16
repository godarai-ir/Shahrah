import { forbidden } from "next/navigation";

import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";

export async function requirePermissionOr403(tenantSlug: string, userID: string, required: string[]) {
  try {
    const payload = await getCurrentMePermissions(tenantSlug, userID);
    const permissionSet = getNormalizedPermissionSet(payload);
    const allowed = required.some((item) => permissionSet.has(item.trim().toLowerCase()));
    if (!allowed) {
      forbidden();
    }
  } catch {
    forbidden();
  }
}
