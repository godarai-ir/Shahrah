import { hiddenRolesInUI } from "@/lib/access/role-display";
import { fullName, guessDisplayName, resolveIdentity } from "@/lib/access/identity";
import type { TenantAdminUser } from "@/lib/contracts/tenant-users";
import type { SeatType } from "@/lib/seats";

import type { UsersURLState } from "./url-state";

export function isUserActive(user: TenantAdminUser): boolean {
  return user.active && user.role !== "deactivated";
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
  state: Pick<UsersURLState, "q" | "seat" | "status">,
  seatByUserID: ReadonlyMap<string, SeatType> = new Map(),
): TenantAdminUser[] {
  const needle = state.q.trim().toLocaleLowerCase();

  return members.filter((member) => {
    if (state.status === "active" && !isUserActive(member)) return false;
    if (state.status === "deactivated" && isUserActive(member)) return false;
    if (state.seat && seatByUserID.get(member.id) !== state.seat) return false;
    if (!needle) return true;

    const identity = resolveIdentity(member);
    return [identity, guessDisplayName(identity), fullName(member)].some((value) =>
      value.toLocaleLowerCase().includes(needle),
    );
  });
}
