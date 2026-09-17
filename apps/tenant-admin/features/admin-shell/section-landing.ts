import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { buildAdminNavItems, filterVisibleNavItems, resolveNavTabs, type AdminNavKey } from "@/features/admin-shell/navigation";

/** Picks the tab a section landing route (e.g. /observe) renders: the first one the user may open. */
export async function resolveSectionLandingTab(rootHref: string, tenantSlug: string, userID: string): Promise<AdminNavKey | null> {
  const permissions = [...getNormalizedPermissionSet(await getCurrentMePermissions(tenantSlug, userID))];
  const visibleItems = filterVisibleNavItems(buildAdminNavItems(rootHref, (key) => key), permissions);
  return resolveNavTabs(rootHref, visibleItems)?.tabs.find((tab) => tab.isCurrent)?.key ?? null;
}
