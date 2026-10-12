import { type PermissionEditorLabels, type PermissionOption } from "@/components/app-ui/PermissionEditor";
import { type PermissionSummaryLabels } from "@/components/app-ui/PermissionSummary";
import type { AccessTranslator } from "@/lib/access/permission-groups";

const REMOVED_PERMISSIONS = new Set(["provider_account:use_integrations"]);

const PERMISSION_RESOURCES = new Set([
  "policy",
  "role",
  "user",
  "team",
  "org_unit",
  "virtual_account",
  "provider_account",
  "settings",
  "logs",
  "billing",
  "support_ticket",
]);

const PERMISSION_ACTIONS = new Set([
  "list",
  "read",
  "create",
  "manage",
  "delete",
  "export",
  "manage_views",
  "read_content",
  "read_metadata",
  "replay",
  "codes_redeem",
  "documents_read",
  "profile_manage",
  "read_own",
  "read_all",
  "reply",
  "rate",
]);

export function filterSupportedPermissions(permissions: PermissionOption[]) {
  return permissions.filter((permission) => !REMOVED_PERMISSIONS.has(permission.key));
}

export function filterSupportedPermissionKeys(permissionKeys: string[]) {
  return permissionKeys.filter((permissionKey) => !REMOVED_PERMISSIONS.has(permissionKey));
}

export function getPermissionEditorLabels(t: AccessTranslator): PermissionEditorLabels {
  return {
    selectedTab: t("permissionEditor.selectedTab"),
    allTab: t("permissionEditor.allTab"),
    searchPlaceholder: t("permissionEditor.searchPlaceholder"),
    searchAriaLabel: t("permissionEditor.searchAriaLabel"),
    emptyTitle: t("permissionEditor.emptyTitle"),
    selectGroup: t("permissionEditor.selectGroup"),
    clearGroup: t("permissionEditor.clearGroup"),
  };
}

export function getPermissionSummaryLabels(t: AccessTranslator): PermissionSummaryLabels {
  return {
    selectedPermissions: t("permissionSummary.selectedPermissions"),
    coveredGroups: t("permissionSummary.coveredGroups"),
    clearAll: t("permissionSummary.clearAll"),
    empty: t("permissionSummary.empty"),
  };
}

/**
 * Prefers a sentence built from the permission key, which reads better than the
 * control plane's own description, and falls back to that description.
 */
export function translatePermissionDescription(permission: PermissionOption, t: AccessTranslator): string {
  return buildPermissionDescription(permission.key, t) || permission.description || permission.key;
}

function buildPermissionDescription(permissionKey: string, t: AccessTranslator): string {
  const [resource, action] = permissionKey.split(":");
  if (!resource || !action) return "";
  if (!PERMISSION_RESOURCES.has(resource) || !PERMISSION_ACTIONS.has(action)) return "";

  return t("permissionDescription", {
    action: t(`permissionActions.${action}`),
    resource: t(`permissionResources.${resource}`),
  });
}
