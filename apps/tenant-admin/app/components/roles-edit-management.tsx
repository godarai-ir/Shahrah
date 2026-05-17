"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft, ListFilter, Search, X } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

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

  const permissionMap = useMemo(
    () => new Map(filteredPermissionCatalog.map((item) => [item.key, item] as const)),
    [filteredPermissionCatalog],
  );

  const selectedPermissionItems = useMemo(
    () =>
      form.permissions
        .map((key) => permissionMap.get(key) ?? { key, group: "other", description: "" })
        .sort((a, b) => a.key.localeCompare(b.key)),
    [form.permissions, permissionMap],
  );

  async function submitEdit() {
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
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-3xl font-bold tracking-tight">{locale.startsWith("fa") ? "ویرایش نقش" : "Edit role"}</h1>
        <p className="text-sm text-muted-foreground">{initialRole.name}</p>
      </header>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <Card>
          <CardHeader>
            <CardTitle>{locale.startsWith("fa") ? "مشخصات نقش" : "Role details"}</CardTitle>
            <CardDescription>
              {locale.startsWith("fa") ? "نام نمایشی، توضیحات و مجوزهای نقش را به‌روزرسانی کنید." : "Update role display name, description, and permissions."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-3 md:grid-cols-2">
              <div className="space-y-2">
                <Label>{locale.startsWith("fa") ? "نام نقش (slug)" : "Role slug"}</Label>
                <Input value={initialRole.name} readOnly className="font-mono" />
              </div>
              <div className="space-y-2">
                <Label>{locale.startsWith("fa") ? "تعداد مجوز انتخابی" : "Selected permissions"}</Label>
                <Input value={`${selectionStats.count}`} readOnly />
              </div>
            </div>

            <div className="space-y-2">
              <Label>{t("rolesPage.form.displayName")}</Label>
              <Input value={form.display_name} onChange={(event) => setForm((prev) => ({ ...prev, display_name: event.target.value }))} />
            </div>

            <div className="space-y-2">
              <Label>{t("rolesPage.form.description")}</Label>
              <textarea
                value={form.description}
                onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
                className="min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
            </div>

            <div className="flex flex-wrap items-center gap-2 rounded-md border bg-muted/20 p-2">
              <Button type="button" size="sm" variant={viewMode === "selected" ? "default" : "outline"} className="gap-2" onClick={() => setViewMode("selected")}>
                <ListFilter className="h-4 w-4" />
                {locale.startsWith("fa") ? "انتخاب‌شده‌ها" : "Selected"}
              </Button>
              <Button type="button" size="sm" variant={viewMode === "all" ? "default" : "outline"} onClick={() => setViewMode("all")}>
                {locale.startsWith("fa") ? "همه مجوزها" : "All permissions"}
              </Button>
              <Badge variant="secondary">{locale.startsWith("fa") ? "گروه‌ها" : "Groups"}: {selectionStats.groupCount}</Badge>
              <Badge variant="secondary">{locale.startsWith("fa") ? "مجوزها" : "Permissions"}: {selectionStats.count}</Badge>
            </div>

            <PermissionPicker
              groupedPermissions={groupedPermissions}
              selected={form.permissions}
              locale={locale}
              query={query}
              viewMode={viewMode}
              onQueryChange={setQuery}
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

            {error ? (
              <Alert variant="destructive">
                <AlertTitle>{t("alerts.error")}</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}

            <div className="flex flex-wrap gap-2">
              <Button variant="ghost" asChild>
                <Link href="/roles" className="gap-2">
                  <ArrowLeft className="h-4 w-4" />
                  {t("rolesPage.actions.cancel")}
                </Link>
              </Button>
              <Button onClick={() => void submitEdit()} disabled={isSubmitting}>
                {isSubmitting ? `${t("rolesPage.actions.save")}...` : t("rolesPage.actions.save")}
              </Button>
            </div>
          </CardContent>
        </Card>

        <aside>
          <Card>
            <CardHeader className="py-4">
              <CardTitle className="text-base">{locale.startsWith("fa") ? "مجوزهای انتخاب‌شده" : "Selected permissions"}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {form.permissions.length > 0 ? (
                <Button size="sm" variant="ghost" onClick={() => setForm((prev) => ({ ...prev, permissions: [] }))}>
                  {locale.startsWith("fa") ? "پاک‌کردن همه" : "Clear all"}
                </Button>
              ) : null}

              <div className="max-h-[60vh] space-y-2 overflow-y-auto rounded-md border bg-background p-2">
                {selectedPermissionItems.length === 0 ? (
                  <p className="p-2 text-xs text-muted-foreground">{locale.startsWith("fa") ? "هنوز مجوزی انتخاب نشده است." : "No permission selected yet."}</p>
                ) : (
                  selectedPermissionItems.map((item) => (
                    <div key={item.key} className="rounded-md border p-2">
                      <div className="mb-1 flex items-start justify-between gap-2">
                        <p className="text-xs font-medium leading-5">{translatePermissionDescription(item, locale)}</p>
                        <Button
                          type="button"
                          size="icon"
                          variant="ghost"
                          className="h-6 w-6 shrink-0"
                          onClick={() =>
                            setForm((prev) => ({
                              ...prev,
                              permissions: prev.permissions.filter((permission) => permission !== item.key),
                            }))
                          }
                        >
                          <X className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                      <p className="truncate font-mono text-[11px] text-muted-foreground">{item.key}</p>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>
        </aside>
      </div>
    </div>
  );
}

function PermissionPicker({
  groupedPermissions,
  selected,
  locale,
  query,
  viewMode,
  onQueryChange,
  onToggle,
  onToggleGroup,
}: {
  groupedPermissions: [string, PermissionItem[]][];
  selected: string[];
  locale: string;
  query: string;
  viewMode: "selected" | "all";
  onQueryChange: (value: string) => void;
  onToggle: (permission: string) => void;
  onToggleGroup: (keys: string[], checked: boolean) => void;
}) {
  const filteredGroups = groupedPermissions
    .map(([group, items]) => [group, viewMode === "selected" ? items.filter((item) => selected.includes(item.key)) : items] as [string, PermissionItem[]])
    .filter(([, items]) => items.length > 0);

  return (
    <div className="space-y-3 rounded-md border p-3">
      <div className="relative">
        <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input value={query} onChange={(event) => onQueryChange(event.target.value)} className="ps-9" placeholder={locale.startsWith("fa") ? "جستجوی مجوز یا توضیح" : "Search permission key or description"} />
      </div>

      <div className="max-h-[40vh] space-y-2 overflow-y-auto pe-1 xl:max-h-[50vh]">
        {filteredGroups.length === 0 ? (
          <p className="text-sm text-muted-foreground">{locale.startsWith("fa") ? "نتیجه‌ای پیدا نشد." : "No matching permissions found."}</p>
        ) : null}

        {filteredGroups.map(([group, items]) => {
          const keys = items.map((item) => item.key);
          const selectedCount = keys.filter((key) => selected.includes(key)).length;
          const allSelected = selectedCount > 0 && selectedCount === keys.length;
          return (
            <section key={group} className="rounded-md border">
              <div className="flex flex-wrap items-center justify-between gap-2 border-b bg-muted/20 px-3 py-2">
                <div className="flex items-center gap-2">
                  <p className="text-sm font-semibold">{translateGroupLabel(group, locale)}</p>
                  <Badge variant="outline">{selectedCount}/{keys.length}</Badge>
                </div>
                <Button size="sm" variant="outline" onClick={() => onToggleGroup(keys, !allSelected)} type="button">
                  {allSelected ? (locale.startsWith("fa") ? "حذف انتخاب گروه" : "Clear group") : locale.startsWith("fa") ? "انتخاب کل گروه" : "Select group"}
                </Button>
              </div>
              <div className="space-y-1 p-2">
                {items.map((item) => (
                  <label key={item.key} className="flex cursor-pointer items-start gap-2 rounded-md p-2 hover:bg-muted/40">
                    <input
                      type="checkbox"
                      checked={selected.includes(item.key)}
                      onChange={() => onToggle(item.key)}
                      className="mt-1 h-4 w-4 rounded border-input text-primary focus:ring-2 focus:ring-ring"
                    />
                    <span className="min-w-0">
                      <span className="block text-sm font-medium">{translatePermissionDescription(item, locale)}</span>
                      <span className="block truncate font-mono text-xs text-muted-foreground">{item.key}</span>
                    </span>
                  </label>
                ))}
              </div>
            </section>
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
  };
  const actionMap: Record<string, string> = {
    list: "مشاهده فهرست",
    read: "مشاهده جزئیات",
    create: "ایجاد",
    manage: "مدیریت",
    delete: "حذف",
  };

  const resource = resourceMap[resourceRaw ?? ""];
  const action = actionMap[actionRaw ?? ""];
  if (!resource || !action) return "";
  return `اجازه ${action} ${resource}`;
}
