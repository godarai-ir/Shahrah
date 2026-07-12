import type { PermissionEditorLabels, PermissionOption, PermissionSummaryLabels } from "@/components/app-ui";

const REMOVED_PERMISSIONS = new Set(["provider_account:use_integrations"]);

export function filterSupportedPermissions(permissions: PermissionOption[]) {
  return permissions.filter((permission) => !REMOVED_PERMISSIONS.has(permission.key));
}

export function filterSupportedPermissionKeys(permissionKeys: string[]) {
  return permissionKeys.filter((permissionKey) => !REMOVED_PERMISSIONS.has(permissionKey));
}

export function getPermissionEditorLabels(locale: string): PermissionEditorLabels {
  const isRtl = locale.startsWith("fa");
  return {
    selectedTab: isRtl ? "انتخاب‌شده‌ها" : "Selected",
    allTab: isRtl ? "همه مجوزها" : "All permissions",
    searchPlaceholder: isRtl ? "جست‌وجوی مجوز یا توضیح" : "Search permission key or description",
    searchAriaLabel: isRtl ? "جست‌وجوی مجوزها" : "Search permissions",
    emptyTitle: isRtl ? "نتیجه‌ای پیدا نشد." : "No matching permissions found.",
    selectGroup: isRtl ? "انتخاب کل گروه" : "Select group",
    clearGroup: isRtl ? "حذف انتخاب گروه" : "Clear group",
  };
}

export function getPermissionSummaryLabels(locale: string): PermissionSummaryLabels {
  const isRtl = locale.startsWith("fa");
  return {
    selectedPermissions: isRtl ? "مجوزهای انتخاب‌شده" : "Selected permissions",
    coveredGroups: isRtl ? "گروه‌های پوشش‌داده‌شده" : "Groups covered",
    clearAll: isRtl ? "پاک‌کردن همه مجوزها" : "Clear all permissions",
    empty: isRtl ? "هنوز مجوزی برای این نقش انتخاب نشده است." : "No permissions are selected for this role yet.",
  };
}

export function translatePermissionGroup(group: string, locale: string): string {
  if (!locale.startsWith("fa")) return group.replaceAll("_", " ");
  const labels: Record<string, string> = {
    policies: "سیاست‌ها",
    roles: "نقش‌ها",
    users: "کاربران",
    teams: "تیم‌ها",
    organization_units: "واحدهای سازمانی",
    virtual_accounts: "حساب‌های مجازی",
    provider_accounts: "حساب‌های ارائه‌دهنده",
    settings: "تنظیمات",
    logs: "گزارش‌های درخواست",
    billing: "مالی و صورتحساب",
    other: "سایر",
  };
  return labels[group] ?? group.replaceAll("_", " ");
}

export function translatePermissionDescription(permission: PermissionOption, locale: string): string {
  if (!locale.startsWith("fa")) return permission.description || permission.key;
  return generatePersianPermissionDescription(permission.key) || permission.description || permission.key;
}

function generatePersianPermissionDescription(permissionKey: string): string {
  const [resourceRaw, actionRaw] = permissionKey.split(":");
  const resources: Record<string, string> = {
    policy: "سیاست",
    role: "نقش",
    user: "کاربر",
    team: "تیم",
    org_unit: "واحد سازمانی",
    virtual_account: "حساب مجازی",
    provider_account: "حساب ارائه‌دهنده",
    settings: "تنظیمات",
    logs: "گزارش درخواست",
    billing: "صورتحساب",
  };
  const actions: Record<string, string> = {
    list: "مشاهده فهرست",
    read: "مشاهده جزئیات",
    create: "ایجاد",
    manage: "مدیریت",
    delete: "حذف",
    export: "خروجی گرفتن از",
    manage_views: "مدیریت نماهای ذخیره‌شده‌ی",
    read_content: "مشاهده محتوای",
    read_metadata: "مشاهده فراداده‌ی",
    replay: "بازپخش",
    codes_redeem: "استفاده از کدهای تخفیف در",
    documents_read: "مشاهده اسناد",
    profile_manage: "مدیریت پروفایل",
  };
  const resource = resources[resourceRaw ?? ""];
  const action = actions[actionRaw ?? ""];
  return resource && action ? `اجازه ${action} ${resource}` : "";
}
