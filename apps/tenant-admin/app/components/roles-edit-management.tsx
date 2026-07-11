"use client";

import { type FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft, ArrowRight, List, ListFilter, Save } from "lucide-react";

import { AppCard, FormSection, PageHeader, SearchField } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";

type RoleDefinition = {
  id: string;
  tenant_id: string;
  name: string;
  display_name: string;
  description?: string;
  is_system: boolean;
  permissions: string[];
  created_at: string;
  updated_at: string;
};

type PermissionItem = {
  key: string;
  group: string;
  description: string;
};

type RoleEditManagementProps = {
  initialRole: RoleDefinition;
  initialPermissionCatalog: PermissionItem[];
};

const REMOVED_PERMISSIONS = new Set(["provider_account:use_integrations"]);

export default function RoleEditManagement({ initialRole, initialPermissionCatalog }: RoleEditManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const router = useRouter();
  const isRtl = locale.startsWith("fa");
  const BackIcon = isRtl ? ArrowRight : ArrowLeft;

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [query, setQuery] = useState("");
  const [viewMode, setViewMode] = useState<"selected" | "all">("selected");
  const [error, setError] = useState<string | null>(null);

  const [form, setForm] = useState({
    display_name: initialRole.display_name,
    description: initialRole.description ?? "",
    permissions: (initialRole.permissions ?? []).filter((permission) => !REMOVED_PERMISSIONS.has(permission)),
  });

  const filteredPermissionCatalog = useMemo(
    () => initialPermissionCatalog.filter((item) => !REMOVED_PERMISSIONS.has(item.key)),
    [initialPermissionCatalog],
  );

  const groupedPermissions = useMemo(
    () => groupPermissions(filteredPermissionCatalog, query, locale),
    [filteredPermissionCatalog, query, locale],
  );

  const selectionStats = useMemo(() => summarizeSelection(form.permissions, filteredPermissionCatalog), [form.permissions, filteredPermissionCatalog]);
  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);

  async function submitEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!form.display_name.trim()) {
      setError(t("rolesPage.form.displayName"));
      return;
    }
    if (form.permissions.length === 0) {
      setError(locale.startsWith("fa") ? "حداقل یک مجوز انتخاب کنید." : "Select at least one permission.");
      return;
    }

    setError(null);
    setIsSubmitting(true);

    try {
      await requestJSON(`/api/roles/${encodeURIComponent(initialRole.id)}`, {
        method: "PATCH",
        body: JSON.stringify(form),
      });
      router.push("/roles");
      router.refresh();
    } catch (submitError) {
      setError(normalizeError(submitError, t("messages.roleUpdateError")));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
      <PageHeader
        eyebrow={t("title")}
        title={isRtl ? "ویرایش نقش" : "Edit role"}
        description={isRtl ? "مشخصات نقش و سطح دسترسی آن را به‌روزرسانی کنید." : "Update the role details and its access level."}
        titleAddon={
          <Badge variant={initialRole.is_system ? "secondary" : "outline"}>
            {initialRole.is_system ? t("rolesPage.systemRole") : t("rolesPage.customRole")}
          </Badge>
        }
        actions={
          <>
            <Button variant="outline" asChild>
              <Link href="/roles">
                <BackIcon data-icon="inline-start" aria-hidden="true" />
                {t("rolesPage.actions.cancel")}
              </Link>
            </Button>
            <Button type="submit" form="role-edit-form" disabled={isSubmitting}>
              <Save data-icon="inline-start" aria-hidden="true" />
              {isSubmitting ? `${t("rolesPage.actions.save")}...` : t("rolesPage.actions.save")}
            </Button>
          </>
        }
      />

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.error")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid min-w-0 gap-6 xl:grid-cols-[minmax(0,1fr)_18rem]">
        <form id="role-edit-form" className="flex min-w-0 flex-col gap-6" onSubmit={submitEdit}>
          <FormSection
            title={isRtl ? "مشخصات نقش" : "Role details"}
            description={isRtl ? "نام نمایشی و توضیحات قابل مشاهده برای مدیران را ویرایش کنید." : "Edit the display name and administrator-facing description."}
          >
            <div className="grid gap-4 md:grid-cols-2">
              <div className="flex flex-col gap-2">
                <Label htmlFor="role-slug">{isRtl ? "نام نقش (slug)" : "Role slug"}</Label>
                <Input id="role-slug" value={initialRole.name} readOnly dir="ltr" className="font-mono" />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="role-display-name">{t("rolesPage.form.displayName")}</Label>
                <Input
                  id="role-display-name"
                  value={form.display_name}
                  required
                  aria-invalid={Boolean(error && !form.display_name.trim())}
                  onChange={(event) => setForm((prev) => ({ ...prev, display_name: event.target.value }))}
                />
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="role-description">{t("rolesPage.form.description")}</Label>
              <Textarea
                id="role-description"
                value={form.description}
                className="min-h-24 resize-y"
                onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
              />
            </div>
          </FormSection>

          <FormSection
            title={isRtl ? "مجوزهای نقش" : "Role permissions"}
            description={isRtl ? "مجوزها را جست‌وجو کنید و دسترسی‌های این نقش را در سطح گروه یا مورد تغییر دهید." : "Search permissions and adjust this role at group or item level."}
            actions={<Badge variant="secondary">{numberFormatter.format(selectionStats.count)}</Badge>}
          >
            <Tabs value={viewMode} onValueChange={(value) => setViewMode(value as "selected" | "all")} dir={isRtl ? "rtl" : "ltr"}>
              <TabsList className="w-full sm:w-fit">
                <TabsTrigger value="selected">
                  <ListFilter aria-hidden="true" />
                  {isRtl ? "انتخاب‌شده‌ها" : "Selected"}
                </TabsTrigger>
                <TabsTrigger value="all">
                  <List aria-hidden="true" />
                  {isRtl ? "همه مجوزها" : "All permissions"}
                </TabsTrigger>
              </TabsList>
              <TabsContent value={viewMode} className="flex flex-col gap-3">
                <SearchField
                  containerClassName="w-full"
                  type="text"
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder={isRtl ? "جست‌وجوی مجوز یا توضیح" : "Search permission key or description"}
                  aria-label={isRtl ? "جست‌وجوی مجوزها" : "Search permissions"}
                  dir={isRtl ? "rtl" : "ltr"}
                />
                <PermissionPicker
                  groupedPermissions={groupedPermissions}
                  selected={form.permissions}
                  locale={locale}
                  viewMode={viewMode}
                  onToggle={(permission) =>
                    setForm((prev) => ({
                      ...prev,
                      permissions: prev.permissions.includes(permission)
                        ? prev.permissions.filter((item) => item !== permission)
                        : [...prev.permissions, permission],
                    }))
                  }
                  onToggleGroup={(keys, checked) =>
                    setForm((prev) => ({
                      ...prev,
                      permissions: checked ? Array.from(new Set([...prev.permissions, ...keys])) : prev.permissions.filter((item) => !keys.includes(item)),
                    }))
                  }
                />
              </TabsContent>
            </Tabs>
          </FormSection>
        </form>

        <aside className="xl:sticky xl:top-6 xl:h-fit">
          <AppCard
            title={isRtl ? "خلاصه نقش" : "Role summary"}
            description={<span dir="ltr" className="font-mono">{initialRole.name}</span>}
            className="shadow-none"
            contentClassName="flex flex-col gap-3"
          >
            <div className="flex items-center justify-between gap-3 text-sm">
              <span className="text-muted-foreground">{isRtl ? "مجوزهای انتخاب‌شده" : "Selected permissions"}</span>
              <Badge variant="secondary">{numberFormatter.format(selectionStats.count)}</Badge>
            </div>
            <div className="flex items-center justify-between gap-3 text-sm">
              <span className="text-muted-foreground">{isRtl ? "گروه‌های پوشش‌داده‌شده" : "Groups covered"}</span>
              <Badge variant="outline">{numberFormatter.format(selectionStats.groupCount)}</Badge>
            </div>
            {form.permissions.length > 0 ? (
              <Button type="button" size="sm" variant="outline" onClick={() => setForm((prev) => ({ ...prev, permissions: [] }))}>
                {isRtl ? "پاک‌کردن همه مجوزها" : "Clear all permissions"}
              </Button>
            ) : (
              <p className="text-xs leading-5 text-muted-foreground">{isRtl ? "هنوز مجوزی برای این نقش انتخاب نشده است." : "No permissions are selected for this role yet."}</p>
            )}
          </AppCard>
        </aside>
      </div>
    </div>
  );
}

function PermissionPicker({
  groupedPermissions,
  selected,
  locale,
  viewMode,
  onToggle,
  onToggleGroup,
}: {
  groupedPermissions: [string, PermissionItem[]][];
  selected: string[];
  locale: string;
  viewMode: "selected" | "all";
  onToggle: (permission: string) => void;
  onToggleGroup: (keys: string[], checked: boolean) => void;
}) {
  const filteredGroups = groupedPermissions
    .map(([group, items]) => [group, viewMode === "selected" ? items.filter((item) => selected.includes(item.key)) : items] as [string, PermissionItem[]])
    .filter(([, items]) => items.length > 0);
  const numberFormatter = new Intl.NumberFormat(locale);

  return (
    <div className="max-h-[48vh] overflow-y-auto pe-1">
      <div className="flex flex-col gap-3">
        {filteredGroups.length === 0 ? (
          <p className="rounded-lg border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
            {locale.startsWith("fa") ? "نتیجه‌ای پیدا نشد." : "No matching permissions found."}
          </p>
        ) : null}

        {filteredGroups.map(([group, items]) => {
          const keys = items.map((item) => item.key);
          const selectedCount = keys.filter((key) => selected.includes(key)).length;
          const allSelected = selectedCount > 0 && selectedCount === keys.length;
          return (
            <fieldset key={group} className="rounded-lg border border-border">
              <legend className="sr-only">{translateGroupLabel(group, locale)}</legend>
              <div className="flex flex-wrap items-center justify-between gap-2 border-b bg-muted/20 px-3 py-2">
                <div className="flex items-center gap-2">
                  <p className="text-sm font-semibold">{translateGroupLabel(group, locale)}</p>
                  <Badge variant="outline">
                    {numberFormatter.format(selectedCount)}/{numberFormatter.format(keys.length)}
                  </Badge>
                </div>
                <Button size="sm" variant="outline" onClick={() => onToggleGroup(keys, !allSelected)} type="button">
                  {allSelected ? (locale.startsWith("fa") ? "حذف انتخاب گروه" : "Clear group") : locale.startsWith("fa") ? "انتخاب کل گروه" : "Select group"}
                </Button>
              </div>
              <div className="flex flex-col gap-1 p-2">
                {items.map((item) => {
                  const checkboxID = `permission-${item.key.replaceAll(":", "-")}`;
                  return (
                    <div key={item.key} className="flex items-start gap-3 rounded-md p-2 hover:bg-muted/40">
                      <Checkbox
                        id={checkboxID}
                        checked={selected.includes(item.key)}
                        onCheckedChange={() => onToggle(item.key)}
                        className="mt-0.5"
                      />
                      <Label htmlFor={checkboxID} className="min-w-0 flex-1 cursor-pointer items-start font-normal">
                        <span className="min-w-0">
                          <span className="block text-sm font-medium">{translatePermissionDescription(item, locale)}</span>
                          <span dir="ltr" className="block truncate text-start font-mono text-xs text-muted-foreground">{item.key}</span>
                        </span>
                      </Label>
                    </div>
                  );
                })}
              </div>
            </fieldset>
          );
        })}
      </div>
    </div>
  );
}

function groupPermissions(permissionCatalog: PermissionItem[], query: string, locale: string): [string, PermissionItem[]][] {
  const normalizedQuery = query.trim().toLowerCase();
  const filtered = !normalizedQuery
    ? permissionCatalog
    : permissionCatalog.filter((item) => {
        const text = `${item.key} ${item.description} ${translatePermissionDescription(item, locale)}`.toLowerCase();
        return text.includes(normalizedQuery);
      });

  const out = new Map<string, PermissionItem[]>();
  for (const permission of filtered) {
    const list = out.get(permission.group) ?? [];
    list.push(permission);
    out.set(permission.group, list);
  }

  return Array.from(out.entries())
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([group, items]) => [group, items.sort((a, b) => a.key.localeCompare(b.key))] as [string, PermissionItem[]]);
}

function summarizeSelection(selected: string[], permissionCatalog: PermissionItem[]) {
  const index = new Map(permissionCatalog.map((item) => [item.key, item] as const));
  const groups = new Set<string>();

  for (const key of selected) {
    const item = index.get(key);
    groups.add(item?.group ?? "other");
  }

  return { count: selected.length, groupCount: groups.size };
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(path, {
    ...options,
    headers,
    cache: "no-store",
  });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) {
    throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
  }
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessage(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}

function translateGroupLabel(group: string, locale: string): string {
  if (!locale.startsWith("fa")) return group.replaceAll("_", " ");
  const map: Record<string, string> = {
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
  return map[group] ?? group.replaceAll("_", " ");
}

function translatePermissionDescription(item: PermissionItem, locale: string): string {
  if (!locale.startsWith("fa")) return item.description || item.key;
  const generated = generatePersianPermissionDescription(item.key);
  if (generated) return generated;
  return item.description || item.key;
}

function generatePersianPermissionDescription(permissionKey: string): string {
  const [resourceRaw, actionRaw] = permissionKey.split(":");
  const resourceMap: Record<string, string> = {
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
  const actionMap: Record<string, string> = {
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

  const resource = resourceMap[resourceRaw ?? ""];
  const action = actionMap[actionRaw ?? ""];
  if (!resource || !action) return "";
  return `اجازه ${action} ${resource}`;
}
