/** Translator bound to the `accessManagement` namespace. */
export type AccessTranslator = (key: string, values?: Record<string, string | number>) => string;

const PERMISSION_GROUPS = new Set([
  "policies",
  "roles",
  "users",
  "teams",
  "virtual_accounts",
  "provider_accounts",
  "settings",
  "logs",
  "billing",
  "support",
  "other",
]);

/** Falls back to the raw group name when the control plane adds a new group. */
export function translatePermissionGroup(group: string, t: AccessTranslator): string {
  return PERMISSION_GROUPS.has(group) ? t(`permissionGroups.${group}`) : group.replaceAll("_", " ");
}
