import { hiddenRolesInUI } from "@/features/roles/role-display";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";

import type { TenantAdminUser } from "./api";
import type { UsersURLState } from "./url-state";

export function resolveIdentity(user: TenantAdminUser): string {
  return user.email || (mobileAuthUIEnabled ? user.mobile : "") || user.id;
}

export function isUserActive(user: TenantAdminUser): boolean {
  return user.active && user.role !== "deactivated";
}

export function guessDisplayName(identity: string): string {
  const value = identity.trim();
  if (!value) return "";
  return value.includes("@") ? value.split("@")[0] : value;
}

export function hasTenantAdminRole(user: TenantAdminUser): boolean {
  return listUserRoleNames(user).some((role) => role === "tenant_admin");
}

/**
 * Every role on the member, lowercased and deduped, with UI-hidden roles dropped.
 * `roles` is the array the control plane returns; `role` is the legacy single field
 * and is only used when the array is absent.
 */
export function listUserRoleNames(user: TenantAdminUser): string[] {
  const names = (user.roles?.length ? user.roles : [user.role ?? ""])
    .map((role) => role.trim().toLowerCase())
    .filter(Boolean)
    .filter((role) => !hiddenRolesInUI.has(role));
  return [...new Set(names)];
}

export function filterMembers(
  members: TenantAdminUser[],
  state: Pick<UsersURLState, "q" | "role" | "status">,
): TenantAdminUser[] {
  const needle = state.q.trim().toLocaleLowerCase();
  const role = state.role.trim().toLowerCase();

  return members.filter((member) => {
    if (state.status === "active" && !isUserActive(member)) return false;
    if (state.status === "deactivated" && isUserActive(member)) return false;
    if (role && !listUserRoleNames(member).includes(role)) return false;
    if (!needle) return true;

    const identity = resolveIdentity(member);
    return [identity, guessDisplayName(identity)].some((value) =>
      value.toLocaleLowerCase().includes(needle),
    );
  });
}
