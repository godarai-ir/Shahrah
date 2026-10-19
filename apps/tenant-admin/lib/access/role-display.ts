/** Roles the control plane returns but that mean nothing to a tenant admin. */
export const hiddenRolesInUI = new Set(["proxy_admin"]);

/** System roles whose permissions cannot be edited; they are duplicated instead. */
export const protectedSystemRoles = new Set(["proxy_admin", "tenant_admin"]);

/** Built-in roles that carry a translated name rather than the control plane's own. */
const LOCALIZED_ROLE_KEYS: Record<string, string> = {
  billing_manager: "billingManager",
  member: "member",
  readonly_member: "readOnlyMember",
  team_admin: "teamAdmin",
  tenant_admin: "tenantAdmin",
};

export type RoleNameSource = { name: string; display_name?: string };

/**
 * Single source for a role's label: translation first, then the tenant's own
 * display name, then the raw role name.
 */
export function resolveRoleDisplayName(
  roleName: string,
  roles: readonly RoleNameSource[],
  t: (key: string) => string,
): string {
  const normalized = roleName.trim().toLowerCase();
  const localizedKey = LOCALIZED_ROLE_KEYS[normalized];
  if (localizedKey) return t(`roles.${localizedKey}`);

  const match = roles.find((role) => role.name.trim().toLowerCase() === normalized);
  return match?.display_name?.trim() || roleName;
}
