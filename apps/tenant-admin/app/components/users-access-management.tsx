"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Plus,
  Search,
} from "lucide-react";

import AccessManagementHeader from "@/app/components/access-management-header";
import { RowActionsMenu } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type TenantAdminUser = {
  id: string;
  tenant_slug: string;
  email?: string;
  mobile?: string;
  role: string;
  roles?: string[];
  active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
};

type UserListResponse = {
  data?: TenantAdminUser[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

type InviteResponse = {
  link?: string;
  invitation?: {
    email?: string;
    role?: string;
  };
};

type RoleDefinition = {
  id: string;
  name: string;
  display_name: string;
  description?: string;
  is_system: boolean;
};

type RoleListResponse = {
  data?: RoleDefinition[];
};

const hiddenRolesInUI = new Set(["proxy_admin"]);

type Banner = {
  kind: "success" | "error";
  message: string;
};

type UsersAccessManagementProps = {
  initialUsers: TenantAdminUser[];
  initialTotal: number;
  initialError: string | null;
  currentUserID: string;
  canManageUsers: boolean;
};

const PAGE_SIZE = 25;

export default function UsersAccessManagement(props: UsersAccessManagementProps) {
  const t = useTranslations("accessManagement");

  const [users, setUsers] = useState<TenantAdminUser[]>(props.initialUsers);
  const [total, setTotal] = useState(props.initialTotal);
  const [query, setQuery] = useState("");
  const [offset, setOffset] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [isInviting, setIsInviting] = useState(false);
  const [isUpdatingUserID, setIsUpdatingUserID] = useState<string | null>(null);
  const [isAssigningRoleUserID, setIsAssigningRoleUserID] = useState<string | null>(null);
  const [availableRoles, setAvailableRoles] = useState<RoleDefinition[]>([]);
  const [roleFormOpen, setRoleFormOpen] = useState(false);
  const [selectedUserForRole, setSelectedUserForRole] = useState<TenantAdminUser | null>(null);
  const [selectedRoleName, setSelectedRoleName] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("member");
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError
      ? {
          kind: "error",
          message: props.initialError,
        }
      : null,
  );
  const [lastInviteLink, setLastInviteLink] = useState<string | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setOffset(0);
      void loadUsers(0, query.trim());
    }, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  useEffect(() => {
    void loadRoles();
  }, []);

  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + users.length, total);
  const canPrev = offset > 0;
  const canNext = offset + users.length < total;

  async function loadUsers(nextOffset: number, q: string) {
    setIsLoading(true);
    try {
      const search = new URLSearchParams();
      search.set("limit", String(PAGE_SIZE));
      search.set("offset", String(nextOffset));
      if (q) search.set("query", q);

      const data = await requestJSON<UserListResponse>(`/api/users?${search.toString()}`);
      setUsers(data.data ?? []);
      setTotal(data.total ?? 0);
      setOffset(nextOffset);
      setBanner(null);
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.loadUsersError")),
      });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadRoles() {
    try {
      const data = await requestJSON<RoleListResponse>("/api/roles?limit=200&offset=0");
      setAvailableRoles((data.data ?? []).filter((role) => !hiddenRolesInUI.has(role.name)));
    } catch {
      // Ignore role loading failure for users table.
    }
  }

  async function handleInvite(event: FormEvent<HTMLFormElement>) {
    if (!props.canManageUsers) return;
    event.preventDefault();
    setBanner(null);
    setLastInviteLink(null);

    const email = inviteEmail.trim();
    if (!email) {
      setBanner({
        kind: "error",
        message: t("messages.provideEmail"),
      });
      return;
    }

    setIsInviting(true);
    try {
      const response = await requestJSON<InviteResponse>("/api/users/invite", {
        method: "POST",
        body: JSON.stringify({
          email,
          role: inviteRole,
        }),
      });
      setLastInviteLink(response.link ?? null);
      setBanner({
        kind: "success",
        message: t("messages.invitationSuccess"),
      });
      setInviteEmail("");
      setInviteRole("member");
      await loadUsers(0, query.trim());
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.inviteError")),
      });
    } finally {
      setIsInviting(false);
    }
  }

  async function changeUserStatus(userID: string, activate: boolean) {
    if (!props.canManageUsers) return;
    setIsUpdatingUserID(userID);
    setBanner(null);
    try {
      await requestJSON(`/api/users/${encodeURIComponent(userID)}/${activate ? "activate" : "deactivate"}`, {
        method: "POST",
      });
      setBanner({
        kind: "success",
        message: activate ? t("messages.activateSuccess") : t("messages.deactivateSuccess"),
      });
      await loadUsers(offset, query.trim());
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.updateStatusError")),
      });
    } finally {
      setIsUpdatingUserID(null);
    }
  }

  async function assignRole(user: TenantAdminUser, roleName: string) {
    if (!props.canManageUsers) return;
    const email = (user.email ?? "").trim();
    if (!email) {
      setBanner({ kind: "error", message: t("messages.roleAssignmentNeedsEmail") });
      return;
    }
    setIsAssigningRoleUserID(user.id);
    setBanner(null);
    try {
      await requestJSON("/api/users/roles", {
        method: "PATCH",
        body: JSON.stringify({
          email,
          roles: [roleName],
          resourceType: "tenant",
        }),
      });
      setBanner({ kind: "success", message: t("messages.roleAssignmentSuccess") });
      await loadUsers(offset, query.trim());
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.roleAssignmentError")),
      });
    } finally {
      setIsAssigningRoleUserID(null);
    }
  }

  function openRoleForm(user: TenantAdminUser) {
    if (!props.canManageUsers) return;
    const currentRole = (user.roles?.[0] ?? user.role ?? "").trim();
    setSelectedUserForRole(user);
    setSelectedRoleName(currentRole);
    setRoleFormOpen(true);
  }

  async function submitRoleForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedUserForRole || !selectedRoleName) return;
    await assignRole(selectedUserForRole, selectedRoleName);
    setRoleFormOpen(false);
    setSelectedUserForRole(null);
  }

  function resolveRoleDisplayName(roleName: string) {
    const normalized = roleName.trim().toLowerCase();
    const match = availableRoles.find((item) => item.name.trim().toLowerCase() === normalized);
    if (match?.display_name?.trim()) {
      return match.display_name;
    }
    if (normalized === "tenant_admin") return t("roles.tenantAdmin");
    if (normalized === "member") {
      return t("roles.member");
    }
    return roleName;
  }

  return (
    <div className="space-y-5">
      <AccessManagementHeader activeTab="users" />

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      {lastInviteLink ? (
        <Alert className="border-primary/40 bg-primary/5">
          <AlertTitle>{t("alerts.invitationLink")}</AlertTitle>
          <AlertDescription>
            <code className="block overflow-x-auto rounded bg-muted px-3 py-2 text-xs">
              {lastInviteLink}
            </code>
          </AlertDescription>
        </Alert>
      ) : null}

      <section className="rounded-xl border bg-card px-4 py-5">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold tracking-wide text-foreground">{t("allUsers")}</h2>
          <div className="flex flex-wrap items-center gap-2">
            <div className="relative">
              <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={t("searchPlaceholder")}
                className="w-64 ps-9"
              />
            </div>

            <Sheet>
              <SheetTrigger asChild>
                <Button className="gap-2" disabled={!props.canManageUsers}>
                  <Plus className="h-4 w-4" />
                  {t("inviteUser")}
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-[94vw] max-w-md">
                <SheetHeader>
                  <SheetTitle>{t("inviteTitle")}</SheetTitle>
                  <SheetDescription>
                    {t("inviteDescription")}
                  </SheetDescription>
                </SheetHeader>

                <form className="mt-6 space-y-4" onSubmit={handleInvite}>
                  <div className="grid gap-2">
                    <label className="text-sm font-medium" htmlFor="invite-email">
                      {t("form.email")}
                    </label>
                    <Input
                      id="invite-email"
                      value={inviteEmail}
                      onChange={(event) => setInviteEmail(event.target.value)}
                      placeholder={t("form.emailPlaceholder")}
                    />
                  </div>
                  <div className="grid gap-2">
                    <label className="text-sm font-medium" htmlFor="invite-role">
                      {t("form.role")}
                    </label>
                    <select
                      id="invite-role"
                      value={inviteRole}
                      onChange={(event) => setInviteRole(event.target.value)}
                      className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                    >
                      <option value="member">{t("roles.member")}</option>
                      <option value="readonly_member">Read-only Member</option>
                      <option value="tenant_admin">{t("roles.tenantAdmin")}</option>
                    </select>
                  </div>

                  <Button type="submit" disabled={isInviting} className="w-full">
                    {isInviting ? t("button.inviting") : t("button.createInvite")}
                  </Button>
                </form>
              </SheetContent>
            </Sheet>
          </div>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("table.emailMobile")}</TableHead>
              <TableHead>{t("table.displayName")}</TableHead>
              <TableHead>{t("table.role")}</TableHead>
              <TableHead className="w-16 text-right"> </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="text-muted-foreground">
                  {t("table.noUsersFound")}
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => {
                const identity = user.email || user.mobile || user.id;
                const displayName = guessDisplayName(identity);
                const active = user.active && user.role !== "deactivated";
                const effectiveRole = user.roles?.[0] ?? user.role;
                const effectiveRoleLabel = resolveRoleDisplayName(effectiveRole);
                const isSelf = user.id === props.currentUserID;
                const isTenantAdminRole = hasTenantAdminRole(user);
                const canDeactivate = !isSelf && !isTenantAdminRole;
                const canChangeRole = !isSelf;

                return (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{identity}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Avatar className="h-8 w-8">
                          <AvatarFallback>{displayName.slice(0, 1).toUpperCase()}</AvatarFallback>
                        </Avatar>
                        <span>{displayName || "-"}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={active ? "secondary" : "outline"}>
                        {active ? effectiveRoleLabel : t("status.deactivated")}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-end">
                      <RowActionsMenu
                        label={t("actions.changeRole")}
                        disabled={isSelf || isUpdatingUserID === user.id}
                        actions={[
                          {
                            label: t("actions.changeRole"),
                            disabled: !props.canManageUsers || !canChangeRole || !user.email || availableRoles.length === 0 || isAssigningRoleUserID === user.id,
                            onSelect: () => openRoleForm(user),
                          },
                          active
                            ? {
                                label: t("actions.deactivate"),
                                disabled: !props.canManageUsers || !canDeactivate,
                                onSelect: () => void changeUserStatus(user.id, false),
                              }
                            : {
                                label: t("actions.activate"),
                                disabled: !props.canManageUsers,
                                onSelect: () => void changeUserStatus(user.id, true),
                              },
                        ]}
                      />
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>

        <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
          <p>
            {t("pagination.showingResults", { from, to, total })}
          </p>
          <div className="flex items-center gap-2">
            <span>{t("pagination.rowsPerPage")}</span>
            <Badge variant="outline">{PAGE_SIZE}</Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void loadUsers(Math.max(0, offset - PAGE_SIZE), query.trim())}
              disabled={!canPrev || isLoading}
            >
              {t("pagination.prev")}
            </Button>
            <Badge variant="secondary">{Math.floor(offset / PAGE_SIZE) + 1}</Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void loadUsers(offset + PAGE_SIZE, query.trim())}
              disabled={!canNext || isLoading}
            >
              {t("pagination.next")}
            </Button>
          </div>
        </div>
      </section>

      <Sheet open={roleFormOpen} onOpenChange={setRoleFormOpen}>
        <SheetContent side="right" className="w-[94vw] max-w-lg">
          <SheetHeader>
            <SheetTitle>{t("roleForm.title")}</SheetTitle>
            <SheetDescription>
              {selectedUserForRole?.email ?? selectedUserForRole?.mobile ?? selectedUserForRole?.id ?? ""}
            </SheetDescription>
          </SheetHeader>
          <form className="mt-6 space-y-4" onSubmit={submitRoleForm}>
            <div className="max-h-[56vh] space-y-3 overflow-auto pr-1">
              {availableRoles.map((role) => (
                <label key={role.id} className="block cursor-pointer rounded-md border p-3">
                  <div className="flex items-start gap-3">
                    <input
                      type="radio"
                      name="role"
                      value={role.name}
                      checked={selectedRoleName === role.name}
                      onChange={(event) => setSelectedRoleName(event.target.value)}
                      className="mt-1"
                    />
                    <div>
                      <p className="text-sm font-semibold">{role.display_name || role.name}</p>
                      <p className="text-xs text-muted-foreground">{role.name}</p>
                      {role.description?.trim() ? (
                        <p className="mt-1 text-xs text-muted-foreground">{role.description}</p>
                      ) : null}
                    </div>
                  </div>
                </label>
              ))}
            </div>
            <Button
              type="submit"
              className="w-full"
              disabled={!props.canManageUsers || !selectedUserForRole || !selectedRoleName || isAssigningRoleUserID === selectedUserForRole?.id}
            >
              {isAssigningRoleUserID === selectedUserForRole?.id ? t("roleForm.submitting") : t("roleForm.submit")}
            </Button>
          </form>
        </SheetContent>
      </Sheet>
    </div>
  );
}

function guessDisplayName(identity: string) {
  const value = identity.trim();
  if (!value) return "";
  if (value.includes("@")) {
    return value.split("@")[0];
  }
  return value;
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

function hasTenantAdminRole(user: TenantAdminUser): boolean {
  if (user.role?.trim().toLowerCase() === "tenant_admin") {
    return true;
  }
  return (user.roles ?? []).some((role) => role.trim().toLowerCase() === "tenant_admin");
}

