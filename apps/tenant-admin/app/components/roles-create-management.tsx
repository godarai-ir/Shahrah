"use client";

import { FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft, Plus, Search } from "lucide-react";

import AccessManagementHeader from "@/app/components/access-management-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

type PermissionItem = {
  key: string;
  group: string;
  description: string;
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type RoleCreateManagementProps = {
  initialPermissionCatalog: PermissionItem[];
};

type RoleFormState = {
  name: string;
  display_name: string;
  description: string;
  permissions: string[];
};

export default function RoleCreateManagement({ initialPermissionCatalog }: RoleCreateManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const router = useRouter();

  const [form, setForm] = useState<RoleFormState>({ name: "", display_name: "", description: "", permissions: [] });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [query, setQuery] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [banner, setBanner] = useState<Banner | null>(null);

  const groupedPermissions = useMemo(
    () => groupPermissions(initialPermissionCatalog, query, locale),
    [initialPermissionCatalog, query, locale],
  );

  const stats = useMemo(() => summarizeSelection(form.permissions, initialPermissionCatalog), [form.permissions, initialPermissionCatalog]);

  function togglePermission(permission: string) {
    setForm((prev) => ({
      ...prev,
      permissions: prev.permissions.includes(permission)
        ? prev.permissions.filter((item) => item !== permission)
        : [...prev.permissions, permission],
    }));
  }

  function toggleGroup(keys: string[], checked: boolean) {
    setForm((prev) => ({
      ...prev,
      permissions: checked ? Array.from(new Set([...prev.permissions, ...keys])) : prev.permissions.filter((item) => !keys.includes(item)),
    }));
  }

  function validate(next: RoleFormState): string | null {
    if (!next.name.trim()) return t("rolesPage.form.name");
    if (!next.display_name.trim()) return t("rolesPage.form.displayName");
    if (next.permissions.length === 0) return locale.startsWith("fa") ? "حداقل یک مجوز انتخاب کنید." : "Select at least one permission.";
    return null;
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);

    const validation = validate(form);
    if (validation) {
      setError(validation);
      return;
    }
    setError(null);

    setIsSubmitting(true);
    try {
      await requestJSON("/api/roles", {
        method: "POST",
        body: JSON.stringify(form),
      });
      setBanner({ kind: "success", message: t("messages.roleCreated") });
      router.push("/roles");
      router.refresh();
    } catch (submitError) {
      setBanner({ kind: "error", message: normalizeError(submitError, t("messages.roleCreateError")) });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="space-y-6">
      <AccessManagementHeader activeTab="roles" />
      <p className="text-sm text-muted-foreground">
        {locale.startsWith("fa")
          ? "ایجاد نقش در صفحه جدا انجام می‌شود تا جریان ثبت دقیق‌تر و بدون شلوغی باشد."
          : "Create a role in a dedicated page for a focused, low-noise workflow."}
      </p>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-12">
        <Card className="lg:col-span-8">
          <CardHeader>
            <CardTitle>{t("rolesPage.createTitle")}</CardTitle>
            <CardDescription>
              {locale.startsWith("fa") ? "مرحله ۱: اطلاعات نقش. مرحله ۲: انتخاب مجوز. مرحله ۳: ثبت." : "Step 1: Role details. Step 2: Permission selection. Step 3: Submit."}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-5" onSubmit={onSubmit}>
              <section className="space-y-3">
                <div className="grid gap-3 md:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="role-name">{t("rolesPage.form.name")}</Label>
                    <Input id="role-name" value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} placeholder="tenant_finance_admin" />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="role-display-name">{t("rolesPage.form.displayName")}</Label>
                    <Input id="role-display-name" value={form.display_name} onChange={(e) => setForm((p) => ({ ...p, display_name: e.target.value }))} placeholder={locale.startsWith("fa") ? "مدیر مالی تننت" : "Tenant Finance Admin"} />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="role-description">{t("rolesPage.form.description")}</Label>
                  <textarea id="role-description" value={form.description} onChange={(e) => setForm((p) => ({ ...p, description: e.target.value }))} className="min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" />
                </div>
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="relative">
                  <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input value={query} onChange={(e) => setQuery(e.target.value)} className="ps-9" placeholder={locale.startsWith("fa") ? "جستجوی مجوز" : "Search permission"} />
                </div>

                <div className="space-y-2 rounded-md border p-3">
                  {groupedPermissions.map(([group, items]) => {
                    const keys = items.map((i) => i.key);
                    const selectedCount = keys.filter((key) => form.permissions.includes(key)).length;
                    const allSelected = selectedCount > 0 && selectedCount === keys.length;
                    return (
                      <details key={group} className="rounded-md border" open>
                        <summary className="flex cursor-pointer items-center justify-between px-3 py-2">
                          <span className="text-sm font-semibold">{translateGroupLabel(group, locale)}</span>
                          <Badge variant="outline">{selectedCount}/{keys.length}</Badge>
                        </summary>
                        <div className="space-y-2 px-3 pb-3">
                          <Button type="button" size="sm" variant="outline" onClick={() => toggleGroup(keys, !allSelected)}>
                            {allSelected ? (locale.startsWith("fa") ? "حذف انتخاب گروه" : "Clear group") : locale.startsWith("fa") ? "انتخاب کل گروه" : "Select group"}
                          </Button>
                          {items.map((item) => (
                            <label key={item.key} className="flex items-start gap-2 rounded-md p-2 hover:bg-muted/40">
                              <input type="checkbox" checked={form.permissions.includes(item.key)} onChange={() => togglePermission(item.key)} className="mt-1" />
                              <span>
                                <span className="block truncate font-mono text-xs text-muted-foreground">{item.key}</span>
                                <span className="block text-sm font-medium">{translatePermissionDescription(item, locale)}</span>
                              </span>
                            </label>
                          ))}
                        </div>
                      </details>
                    );
                  })}
                </div>
              </section>

              {error ? <p className="text-sm text-destructive">{error}</p> : null}

              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={isSubmitting} className="gap-2">
                  <Plus className="h-4 w-4" />
                  {isSubmitting ? t("rolesPage.form.creating") : t("rolesPage.form.create")}
                </Button>
                <Button asChild type="button" variant="ghost" className="gap-2">
                  <Link href="/roles">
                    <ArrowLeft className="h-4 w-4" />
                    {locale.startsWith("fa") ? "بازگشت به فهرست نقش‌ها" : "Back to roles list"}
                  </Link>
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <Card className="lg:col-span-4 lg:sticky lg:top-6 lg:h-fit">
          <CardHeader>
            <CardTitle>{locale.startsWith("fa") ? "خلاصه انتخاب‌ها" : "Selection summary"}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <SummaryStat label={locale.startsWith("fa") ? "مجوزهای انتخاب‌شده" : "Selected permissions"} value={String(stats.count)} />
            <SummaryStat label={locale.startsWith("fa") ? "گروه‌های درگیر" : "Groups covered"} value={String(stats.groupCount)} />
          </CardContent>
        </Card>
      </div>
    </div>
  );
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

function SummaryStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="text-xl font-semibold">{value}</p>
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
  if (!locale.startsWith("fa")) {
    return group.replaceAll("_", " ");
  }
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
  if (!locale.startsWith("fa")) {
    return item.description || item.key;
  }

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
    use_integrations: "استفاده از یکپارچه‌سازی‌ها",
  };

  const resource = resourceMap[resourceRaw ?? ""];
  const action = actionMap[actionRaw ?? ""];
  if (!resource || !action) return "";
  return `اجازه ${action} ${resource}`;
}
