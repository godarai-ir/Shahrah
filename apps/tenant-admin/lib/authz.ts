export type MePermissionsResponse = {
  user_id?: string;
  role?: string;
  roles?: string[];
  permissions?: string[];
};

export function canPermission(permissions: readonly string[], permission: Permission): boolean {
  const required = permission.trim().toLowerCase();
  if (!required) return true;
  return permissions.some((item) => item.trim().toLowerCase() === required);
}

export function canAnyPermission(permissions: readonly string[], required: readonly Permission[]): boolean {
  if (required.length === 0) return true;
  return required.some((permission) => canPermission(permissions, permission));
}
import type { Permission } from "@/lib/permissions";
