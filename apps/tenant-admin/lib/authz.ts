export type MePermissionsResponse = {
  user_id?: string;
  role?: string;
  roles?: string[];
  permissions?: string[];
};

export function canPermission(permissions: string[], permission: string): boolean {
  const required = permission.trim().toLowerCase();
  if (!required) return true;
  return permissions.some((item) => item.trim().toLowerCase() === required);
}

export function canAnyPermission(permissions: string[], required: string[]): boolean {
  if (required.length === 0) return true;
  return required.some((permission) => canPermission(permissions, permission));
}
