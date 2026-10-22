import type { PermissionItem } from "@/lib/contracts/roles";

export type PermissionGrant = {
  key: string;
  /** Role names that grant this permission, in the order the member holds them. */
  grantedBy: string[];
};

export type PermissionGroupGrants = {
  group: string;
  grants: PermissionGrant[];
};

export type EffectivePermissions = {
  groups: PermissionGroupGrants[];
  total: number;
};

export type RolePermissionSource = { name: string; permissions?: string[] };

/**
 * Union of every permission the member's roles grant, grouped the way the server's
 * catalog groups them. A permission missing from the catalog still shows up, under
 * the group derived from its key, so a newly added server permission is never hidden.
 */
export function collectEffectivePermissions(
  roleNames: readonly string[],
  roles: readonly RolePermissionSource[],
  catalog: readonly PermissionItem[],
): EffectivePermissions {
  const groupByKey = new Map(catalog.map((item) => [item.key, item.group]));
  const grants = new Map<string, PermissionGrant>();

  for (const roleName of roleNames) {
    const role = roles.find((item) => item.name.trim().toLowerCase() === roleName.trim().toLowerCase());
    for (const key of role?.permissions ?? []) {
      const permissionKey = key.trim();
      if (!permissionKey) continue;
      const existing = grants.get(permissionKey);
      if (existing) {
        if (!existing.grantedBy.includes(roleName)) existing.grantedBy.push(roleName);
      } else {
        grants.set(permissionKey, { key: permissionKey, grantedBy: [roleName] });
      }
    }
  }

  const byGroup = new Map<string, PermissionGrant[]>();
  for (const grant of grants.values()) {
    const group = groupByKey.get(grant.key) ?? deriveGroupFromKey(grant.key);
    const bucket = byGroup.get(group);
    if (bucket) bucket.push(grant);
    else byGroup.set(group, [grant]);
  }

  const groups = [...byGroup.entries()]
    .map(([group, items]) => ({ group, grants: [...items].sort((a, b) => a.key.localeCompare(b.key)) }))
    .sort((a, b) => a.group.localeCompare(b.group));

  return { groups, total: grants.size };
}

/** "team:read" -> "team"; used only when the catalog omits the key. */
function deriveGroupFromKey(permissionKey: string): string {
  const [resource] = permissionKey.split(":");
  return resource?.trim() || "other";
}
