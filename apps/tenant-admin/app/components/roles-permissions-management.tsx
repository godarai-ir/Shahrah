"use client";

import { FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { BadgeCheck, BriefcaseBusiness, KeyRound, Plus, UserCircle2, Users } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

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

type RoleListResponse = {
  data?: RoleDefinition[];
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type RolesPermissionsManagementProps = {
  initialRoles: RoleDefinition[];
  initialError: string | null;
  initialPermissionCatalog: PermissionItem[];
};

type PermissionItem = {
  key: string;
  group: string;
  description: string;
};

const protectedSystemRoles = new Set(["proxy_admin", "tenant_admin"]);
const hiddenRolesInUI = new Set(["proxy_admin"]);

export default function RolesPermissionsManagement(props: RolesPermissionsManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();

  const [roles, setRoles] = useState<RoleDefinition[]>(
    props.initialRoles.filter((role) => !hiddenRolesInUI.has(role.name)),
  );
  const [isSubmittingCreate, setIsSubmittingCreate] = useState(false);
  const [editingRoleID, setEditingRoleID] = useState<string | null>(null);
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError
      ? {
          kind: "error",
          message: props.initialError,
        }
      : null,
  );

  const [createForm, setCreateForm] = useState({
    name: "",
    display_name: "",
    description: "",
    permissions: [] as string[],
  });
  const [editForm, setEditForm] = useState({
    display_name: "",
    description: "",
    permissions: [] as string[],
  });

  const groupedPermissions = useMemo(() => {
    const index = new Map<string, PermissionItem>();
    for (const permission of props.initialPermissionCatalog) {
      if (!permission.key?.trim()) continue;
      index.set(permission.key, permission);
    }
    for (const role of roles) {
      for (const permissionKey of role.permissions ?? []) {
        if (index.has(permissionKey)) continue;
        const [rawGroup] = permissionKey.split(":");
        index.set(permissionKey, {
          key: permissionKey,
          group: rawGroup ? `${rawGroup}s` : "other",
          description: "",
        });
      }
    }

    const out = new Map<string, PermissionItem[]>();
    for (const permission of index.values()) {
      const list = out.get(permission.group) ?? [];
      list.push(permission);
      out.set(permission.group, list);
    }
    return Array.from(out.entries())
      .sort((a, b) => a[0].localeCompare(b[0]))
      .map(([group, items]) => [group, items.sort((a, b) => a.key.localeCompare(b.key))] as [string, PermissionItem[]]);
  }, [props.initialPermissionCatalog, roles]);

  const accessTabs = useMemo(
    () => [
      { href: "/users", label: t("tabs.users"), icon: UserCircle2, active: false },
      { href: "/teams", label: t("tabs.teams"), icon: Users, active: false },
      { href: "/personal-access-tokens", label: t("tabs.personalAccessTokens"), icon: KeyRound, active: false },
      { href: "#", label: t("tabs.virtualAccounts"), icon: BriefcaseBusiness, active: false },
      { href: "/roles", label: t("tabs.roles"), icon: BadgeCheck, active: true },
    ],
    [t],
  );

  function togglePermission(current: string[], permission: string) {
    if (current.includes(permission)) {
      return current.filter((item) => item !== permission);
    }
    return [...current, permission];
  }

  async function refreshRoles() {
    const data = await requestJSON<RoleListResponse>("/api/roles");
    setRoles((data.data ?? []).filter((role) => !hiddenRolesInUI.has(role.name)));
  }

  async function handleCreateRole(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);
    setIsSubmittingCreate(true);
    try {
      await requestJSON("/api/roles", {
        method: "POST",
        body: JSON.stringify(createForm),
      });
      setCreateForm({
        name: "",
        display_name: "",
        description: "",
        permissions: [],
      });
      await refreshRoles();
      setBanner({ kind: "success", message: t("messages.roleCreated") });
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.roleCreateError")),
      });
    } finally {
      setIsSubmittingCreate(false);
    }
  }

  function startEdit(role: RoleDefinition) {
    setEditingRoleID(role.id);
    setEditForm({
      display_name: role.display_name,
      description: role.description ?? "",
      permissions: role.permissions ?? [],
    });
  }

  async function submitEdit(roleID: string) {
    setBanner(null);
    try {
      await requestJSON(`/api/roles/${encodeURIComponent(roleID)}`, {
        method: "PATCH",
        body: JSON.stringify(editForm),
      });
      setEditingRoleID(null);
      await refreshRoles();
      setBanner({ kind: "success", message: t("messages.roleUpdated") });
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.roleUpdateError")),
      });
    }
  }

  return (
    <div className="space-y-5">
      <header className="space-y-4">
        <h1 className="text-4xl font-bold tracking-tight">{t("title")}</h1>
        <nav className="border-b">
          <ul className="flex flex-wrap items-center gap-5 pb-2">
            {accessTabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <li key={tab.label}>
                  <Link
                    href={tab.href}
                    className={`inline-flex items-center gap-2 border-b-2 px-1 py-2 ${tab.active ? "border-indigo-500 text-foreground" : "border-transparent text-muted-foreground hover:text-foreground"}`}
                    aria-current={tab.active ? "page" : undefined}
                  >
                    <Icon className="h-4 w-4" />
                    <span className="text-sm font-semibold">{tab.label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
      </header>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <section className="rounded-xl border bg-card p-4">
        <h2 className="mb-4 text-lg font-semibold">{t("rolesPage.createTitle")}</h2>
        <form className="space-y-4" onSubmit={handleCreateRole}>
          <div className="grid gap-2 md:grid-cols-2">
            <Input
              placeholder={t("rolesPage.form.name")}
              value={createForm.name}
              onChange={(event) => setCreateForm((prev) => ({ ...prev, name: event.target.value }))}
            />
            <Input
              placeholder={t("rolesPage.form.displayName")}
              value={createForm.display_name}
              onChange={(event) => setCreateForm((prev) => ({ ...prev, display_name: event.target.value }))}
            />
          </div>
          <textarea
            placeholder={t("rolesPage.form.description")}
            value={createForm.description}
            onChange={(event) => setCreateForm((prev) => ({ ...prev, description: event.target.value }))}
            className="min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
          />
          <PermissionPicker
            groupedPermissions={groupedPermissions}
            selected={createForm.permissions}
            locale={locale}
            onToggle={(permission) =>
              setCreateForm((prev) => ({
                ...prev,
                permissions: togglePermission(prev.permissions, permission),
              }))
            }
          />
          <Button type="submit" disabled={isSubmittingCreate} className="gap-2">
            <Plus className="h-4 w-4" />
            {isSubmittingCreate ? t("rolesPage.form.creating") : t("rolesPage.form.create")}
          </Button>
        </form>
      </section>

      <section className="rounded-xl border bg-card p-4">
        <h2 className="mb-4 text-lg font-semibold">{t("rolesPage.listTitle")}</h2>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("rolesPage.table.name")}</TableHead>
              <TableHead>{t("rolesPage.table.displayName")}</TableHead>
              <TableHead>{t("rolesPage.table.scope")}</TableHead>
              <TableHead>{t("rolesPage.table.permissions")}</TableHead>
              <TableHead>{t("rolesPage.table.actions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {roles.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5}>{t("rolesPage.table.empty")}</TableCell>
              </TableRow>
            ) : (
              roles.map((role) => {
                const isProtected = role.is_system && protectedSystemRoles.has(role.name);
                const isEditing = editingRoleID === role.id;
                return (
                  <TableRow key={role.id}>
                    <TableCell className="font-medium">{role.name}</TableCell>
                    <TableCell>{role.display_name}</TableCell>
                    <TableCell>
                      <Badge variant={role.is_system ? "secondary" : "outline"}>
                        {role.is_system ? t("rolesPage.systemRole") : t("rolesPage.customRole")}
                      </Badge>
                    </TableCell>
                    <TableCell>{(role.permissions ?? []).length}</TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          disabled={isProtected}
                          onClick={() => startEdit(role)}
                        >
                          {t("rolesPage.actions.edit")}
                        </Button>
                        {isProtected ? <span className="text-xs text-muted-foreground">{t("rolesPage.actions.protected")}</span> : null}
                      </div>
                      {isEditing ? (
                        <div className="mt-3 space-y-3 rounded-md border p-3">
                          <Input
                            value={editForm.display_name}
                            onChange={(event) => setEditForm((prev) => ({ ...prev, display_name: event.target.value }))}
                            placeholder={t("rolesPage.form.displayName")}
                          />
                          <textarea
                            value={editForm.description}
                            onChange={(event) => setEditForm((prev) => ({ ...prev, description: event.target.value }))}
                            placeholder={t("rolesPage.form.description")}
                            className="min-h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                          />
                          <PermissionPicker
                            groupedPermissions={groupedPermissions}
                            selected={editForm.permissions}
                            locale={locale}
                            onToggle={(permission) =>
                              setEditForm((prev) => ({
                                ...prev,
                                permissions: togglePermission(prev.permissions, permission),
                              }))
                            }
                          />
                          <div className="flex gap-2">
                            <Button size="sm" onClick={() => void submitEdit(role.id)}>
                              {t("rolesPage.actions.save")}
                            </Button>
                            <Button size="sm" variant="ghost" onClick={() => setEditingRoleID(null)}>
                              {t("rolesPage.actions.cancel")}
                            </Button>
                          </div>
                        </div>
                      ) : null}
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </section>
    </div>
  );
}

function PermissionPicker({
  groupedPermissions,
  selected,
  locale,
  onToggle,
}: {
  groupedPermissions: [string, PermissionItem[]][];
  selected: string[];
  locale: string;
  onToggle: (permission: string) => void;
}) {
  return (
    <div className="grid gap-3 rounded-md border p-3 md:grid-cols-2">
      {groupedPermissions.map(([group, items]) => (
        <div key={group} className="space-y-2">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{translateGroupLabel(group, locale)}</p>
          {items.map((item) => (
            <label key={item.key} className="flex items-start gap-2 text-sm">
              <input
                type="checkbox"
                checked={selected.includes(item.key)}
                onChange={() => onToggle(item.key)}
                className="mt-1"
              />
              <span>
                <span className="block font-medium">{item.key}</span>
                <span className="block text-xs text-muted-foreground">{translatePermissionDescription(item, locale)}</span>
              </span>
            </label>
          ))}
        </div>
      ))}
    </div>
  );
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
