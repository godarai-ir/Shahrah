"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Plus,
  Search,
} from "lucide-react";

import { AppCard, AppSelectField, CommercialPlanUpgradeDialog, EmptyState, PageHeader, RowActionsMenu } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
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

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
} | null;

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
const USERS_FEATURE_KEY = "users.max";

export default function UsersAccessManagement(props: UsersAccessManagementProps) {
  const t = useTranslations("accessManagement");

  const [users, setUsers] = useState<TenantAdminUser[]>(props.initialUsers);
  const [total, setTotal] = useState(props.initialTotal);
  const [query, setQuery] = useState("");
  const [offset, setOffset] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [isInviting, setIsInviting] = useState(false);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
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
    let isMounted = true;

    void Promise.allSettled([
      loadRoles(),
      requestJSON<EffectiveEntitlements>("/api/entitlements").then((response) => {
        if (isMounted) setEntitlements(response);
      }),
    ]).finally(() => {
      if (isMounted) setHasLoadedEntitlements(true);
    });

    return () => {
      isMounted = false;
    };
  }, []);

  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + users.length, total);
  const canPrev = offset > 0;
  const canNext = offset + users.length < total;
  const usersLimit = coerceNumericEntitlement(entitlements?.features?.[USERS_FEATURE_KEY]);
  const hasReachedUsersLimit = usersLimit !== null && total >= usersLimit;

  function handleOpenInvite() {
    if (hasReachedUsersLimit) {
      setUpgradeDialogOpen(true);
      return;
    }
    setInviteError(null);
    setInviteOpen(true);
  }

  function handleInviteOpenChange(open: boolean) {
    setInviteOpen(open);
    if (!open) setInviteError(null);
  }

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
      const nextRoles = (data.data ?? []).filter((role) => !hiddenRolesInUI.has(role.name));
      setAvailableRoles(nextRoles);
      setInviteRole((currentRole) => (
        nextRoles.some((role) => role.name === currentRole)
          ? currentRole
          : resolveDefaultInviteRole(nextRoles)
      ));
    } catch {
      // Ignore role loading failure for users table.
    }
  }

  async function handleInvite(event: FormEvent<HTMLFormElement>) {
    if (!props.canManageUsers) return;
    event.preventDefault();
    setInviteError(null);
    setLastInviteLink(null);

    const email = inviteEmail.trim();
    if (!email) {
      setInviteError(t("messages.provideEmail"));
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
      setInviteRole(resolveDefaultInviteRole(availableRoles));
      setInviteOpen(false);
      await loadUsers(0, query.trim());
    } catch (error) {
      setInviteError(normalizeError(error, t("messages.inviteError")));
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
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("usersPage.title")}
        description={t("usersPage.description")}
        actions={
          <Button type="button" onClick={handleOpenInvite} disabled={!props.canManageUsers || !hasLoadedEntitlements}>
            <Plus data-icon="inline-start" />
            {t("inviteUser")}
          </Button>
        }
      />
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

      <AppCard title={t("allUsers")} description={t("usersPage.listDescription")} contentClassName="flex min-w-0 flex-col gap-4" className="min-w-0 shadow-none">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <InputGroup className="w-full sm:max-w-md">
            <InputGroupAddon><Search aria-hidden="true" /></InputGroupAddon>
            <InputGroupInput
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={t("searchPlaceholder")}
              aria-label={t("searchPlaceholder")}
            />
          </InputGroup>
          <p className="text-sm text-muted-foreground">{t("usersPage.usage", { used: total, limit: usersLimit ?? t("usersPage.unlimited") })}</p>
        </div>

        <div className="overflow-x-auto"><Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("table.emailMobile")}</TableHead>
              <TableHead>{t("table.displayName")}</TableHead>
              <TableHead>{t("table.role")}</TableHead>
              <TableHead className="w-16 text-right"> </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && users.length === 0 ? Array.from({ length: 4 }).map((_, index) => (
              <TableRow key={index}><TableCell colSpan={4}><Skeleton className="h-10 w-full" /></TableCell></TableRow>
            )) : users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="py-6"><EmptyState title={t("table.noUsersFound")} description={t("usersPage.emptyDescription")} className="min-h-40 border-0 bg-transparent" /></TableCell>
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
                        <Avatar className="size-8">
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
        </Table></div>

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
      </AppCard>

      <Sheet open={inviteOpen} onOpenChange={handleInviteOpenChange}>
        <SheetContent side="left" className="w-full max-w-md overflow-y-auto sm:w-3/4">
          <SheetHeader>
            <SheetTitle>{t("inviteTitle")}</SheetTitle>
            <SheetDescription>{t("inviteDescription")}</SheetDescription>
          </SheetHeader>
          <form className="mt-6 flex flex-col gap-4" onSubmit={handleInvite}>
            {inviteError ? (
              <Alert variant="destructive">
                <AlertTitle>{t("alerts.error")}</AlertTitle>
                <AlertDescription>{inviteError}</AlertDescription>
              </Alert>
            ) : null}
            <div className="flex flex-col gap-2">
              <Label htmlFor="invite-email">{t("form.email")}</Label>
              <Input id="invite-email" type="email" value={inviteEmail} onChange={(event) => setInviteEmail(event.target.value)} placeholder={t("form.emailPlaceholder")} />
            </div>
            <AppSelectField
              label={t("form.role")}
              ariaLabel={t("form.role")}
              value={inviteRole}
              onValueChange={setInviteRole}
              triggerClassName="h-10 w-full"
              options={availableRoles.map((role) => ({
                value: role.name,
                label: role.display_name?.trim() || resolveRoleDisplayName(role.name),
              }))}
              disabled={availableRoles.length === 0}
            />
            <Button type="submit" disabled={isInviting || !inviteRole || availableRoles.length === 0} className="w-full">
              {isInviting ? t("button.inviting") : t("button.createInvite")}
            </Button>
          </form>
        </SheetContent>
      </Sheet>

      <CommercialPlanUpgradeDialog
        open={upgradeDialogOpen}
        onOpenChange={setUpgradeDialogOpen}
        title={t("usersPage.upgradeDialog.title")}
        description={t("usersPage.upgradeDialog.description", { limit: usersLimit ?? 0 })}
        cancelLabel={t("usersPage.upgradeDialog.cancel")}
        actionLabel={t("usersPage.upgradeDialog.action")}
      />

      <Sheet open={roleFormOpen} onOpenChange={setRoleFormOpen}>
        <SheetContent side="right" className="w-full max-w-lg overflow-y-auto sm:w-3/4">
          <SheetHeader>
            <SheetTitle>{t("roleForm.title")}</SheetTitle>
            <SheetDescription>
              {selectedUserForRole?.email ?? selectedUserForRole?.mobile ?? selectedUserForRole?.id ?? ""}
            </SheetDescription>
          </SheetHeader>
          <form className="mt-6 flex flex-col gap-4" onSubmit={submitRoleForm}>
            <RadioGroup value={selectedRoleName} onValueChange={setSelectedRoleName} className="flex max-h-96 flex-col gap-3 overflow-auto pe-1">
              {availableRoles.map((role) => (
                <Label key={role.id} htmlFor={`user-role-${role.id}`} className="block cursor-pointer rounded-md border p-3 font-normal">
                  <div className="flex items-start gap-3">
                    <RadioGroupItem id={`user-role-${role.id}`} value={role.name} className="mt-1" />
                    <div>
                      <p className="text-sm font-semibold">{role.display_name || role.name}</p>
                      <p className="text-xs text-muted-foreground">{role.name}</p>
                      {role.description?.trim() ? (
                        <p className="mt-1 text-xs text-muted-foreground">{role.description}</p>
                      ) : null}
                    </div>
                  </div>
                </Label>
              ))}
            </RadioGroup>
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

function resolveDefaultInviteRole(roles: RoleDefinition[]): string {
  return roles.find((role) => role.name === "member")?.name ?? roles[0]?.name ?? "";
}

function coerceNumericEntitlement(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
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

