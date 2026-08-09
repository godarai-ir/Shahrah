"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Info, Plus, UsersRound } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppEmailField } from "@/components/app-ui/AppEmailField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { SearchField } from "@/components/app-ui/SearchField";
import { TablePagination } from "@/components/app-ui/TablePagination";
import { TableEmptyRow, TableLoadingRows } from "@/components/app-ui/TableStateRows";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Progress } from "@/components/ui/progress";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
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
const USERS_COLUMN_CLASS = {
  identity: "w-[44%]",
  displayName: "hidden w-[28%] @md:table-cell",
  role: "w-[36%]",
  actions: "w-14 text-end",
} as const;

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
  const [roleError, setRoleError] = useState<string | null>(null);
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
  const usersUsagePercent = usersLimit && usersLimit > 0 ? Math.min(100, (total / usersLimit) * 100) : 0;

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
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.invitationSuccess") });
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
    try {
      await requestJSON(`/api/users/${encodeURIComponent(userID)}/${activate ? "activate" : "deactivate"}`, {
        method: "POST",
      });
      showFeedbackToast({
        kind: "success",
        title: t("alerts.success"),
        description: activate ? t("messages.activateSuccess") : t("messages.deactivateSuccess"),
      });
      await loadUsers(offset, query.trim());
    } catch (error) {
      showFeedbackToast({ kind: "error", title: t("alerts.error"), description: normalizeError(error, t("messages.updateStatusError")) });
    } finally {
      setIsUpdatingUserID(null);
    }
  }

  async function assignRole(user: TenantAdminUser, roleName: string) {
    if (!props.canManageUsers) return false;
    const email = (user.email ?? "").trim();
    if (!email) {
      setRoleError(t("messages.roleAssignmentNeedsEmail"));
      return false;
    }
    setIsAssigningRoleUserID(user.id);
    setRoleError(null);
    try {
      await requestJSON("/api/users/roles", {
        method: "PATCH",
        body: JSON.stringify({
          email,
          roles: [roleName],
          resourceType: "tenant",
        }),
      });
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.roleAssignmentSuccess") });
      await loadUsers(offset, query.trim());
      return true;
    } catch (error) {
      setRoleError(normalizeError(error, t("messages.roleAssignmentError")));
      return false;
    } finally {
      setIsAssigningRoleUserID(null);
    }
  }

  function openRoleForm(user: TenantAdminUser) {
    if (!props.canManageUsers) return;
    const currentRole = (user.roles?.[0] ?? user.role ?? "").trim();
    setSelectedUserForRole(user);
    setSelectedRoleName(currentRole);
    setRoleError(null);
    setRoleFormOpen(true);
  }

  async function submitRoleForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedUserForRole || !selectedRoleName) return;
    const succeeded = await assignRole(selectedUserForRole, selectedRoleName);
    if (!succeeded) return;
    setRoleFormOpen(false);
    setSelectedUserForRole(null);
  }

  function resolveRoleDisplayName(roleName: string) {
    const normalized = roleName.trim().toLowerCase();
    const localizedRoleKey = {
      billing_manager: "billingManager",
      member: "member",
      org_admin: "orgAdmin",
      readonly_member: "readOnlyMember",
      team_admin: "teamAdmin",
      tenant_admin: "tenantAdmin",
    }[normalized];
    if (localizedRoleKey) {
      return t(`roles.${localizedRoleKey}`);
    }
    const match = availableRoles.find((item) => item.name.trim().toLowerCase() === normalized);
    if (match?.display_name?.trim()) {
      return match.display_name;
    }
    return roleName;
  }

  return (
    <div className="flex flex-col gap-6">
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

      <AppCard contentClassName="flex min-w-0 flex-col gap-4" className="min-w-0 shadow-none">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <SearchField
              containerClassName="w-full sm:max-w-md"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={t("searchPlaceholder")}
          />
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <Button type="button" className="w-full sm:w-auto" onClick={handleOpenInvite} disabled={!props.canManageUsers || !hasLoadedEntitlements}>
              <Plus data-icon="inline-start" />
              {t("inviteUser")}
            </Button>
            <TooltipProvider delayDuration={150}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div tabIndex={0} className="flex min-w-fit items-center gap-2 rounded-lg border border-border bg-muted/40 px-2.5 py-1.5 outline-none focus-visible:ring-2 focus-visible:ring-ring">
                    <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                      <UsersRound className="size-4" aria-hidden="true" />
                    </span>
                    <span className="text-sm font-semibold">{t("usersPage.usage", { used: total, limit: usersLimit ?? t("usersPage.unlimited") })}</span>
                    {usersLimit !== null ? <Progress dir="ltr" value={usersUsagePercent} className="w-16" aria-label={t("usersPage.progressLabel", { value: Math.round(usersUsagePercent) })} /> : null}
                    <Info className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                  </div>
                </TooltipTrigger>
                <TooltipContent className="max-w-xs">
                  <p>{usersLimit === null ? t("usersPage.unlimitedTooltip") : t("usersPage.limitTooltip", { limit: usersLimit })}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>

        <div className="@container/users overflow-hidden rounded-lg border border-border"><Table className="table-fixed text-xs lg:text-sm">
          <TableHeader>
            <TableRow>
              <TableHead className={USERS_COLUMN_CLASS.identity}>{mobileAuthUIEnabled ? t("table.emailMobile") : t("table.email")}</TableHead>
              <TableHead className={USERS_COLUMN_CLASS.displayName}>{t("table.displayName")}</TableHead>
              <TableHead className={USERS_COLUMN_CLASS.role}>{t("table.role")}</TableHead>
              <TableHead className={USERS_COLUMN_CLASS.actions}> </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && users.length === 0 ? <TableLoadingRows colSpan={4} label={t("table.loading")} /> : users.length === 0 ? (
              <TableEmptyRow colSpan={4} title={t("table.noUsersFound")} description={t("usersPage.emptyDescription")} />
            ) : (
              users.map((user) => {
                const identity = user.email || (mobileAuthUIEnabled ? user.mobile : "") || user.id;
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
                    <TableCell className={`${USERS_COLUMN_CLASS.identity} min-w-0 font-medium`}>
                      <span className="block truncate" title={identity}>{identity}</span>
                    </TableCell>
                    <TableCell className={USERS_COLUMN_CLASS.displayName}>
                      <div className="flex items-center gap-2">
                        <Avatar className="size-8">
                          <AvatarFallback>{displayName.slice(0, 1).toUpperCase()}</AvatarFallback>
                        </Avatar>
                        <span className="truncate">{displayName || "-"}</span>
                      </div>
                    </TableCell>
                    <TableCell className={USERS_COLUMN_CLASS.role}>
                      <Badge variant={active ? "secondary" : "outline"}>
                        {active ? effectiveRoleLabel : t("status.deactivated")}
                      </Badge>
                    </TableCell>
                    <TableCell className={USERS_COLUMN_CLASS.actions}>
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

        <TablePagination
          className="mt-4"
          label={t("usersPage.title")}
          summary={t("pagination.showingResults", { from, to, total })}
          pageSizeLabel={t("pagination.rowsPerPage")}
          pageSize={PAGE_SIZE}
          page={Math.floor(offset / PAGE_SIZE) + 1}
          previousLabel={t("pagination.prev")}
          nextLabel={t("pagination.next")}
          canPrevious={canPrev}
          canNext={canNext}
          disabled={isLoading}
          onPrevious={() => void loadUsers(Math.max(0, offset - PAGE_SIZE), query.trim())}
          onNext={() => void loadUsers(offset + PAGE_SIZE, query.trim())}
        />
      </AppCard>

      <FormSheet open={inviteOpen} onOpenChange={handleInviteOpenChange} title={t("inviteTitle")} description={t("inviteDescription")} size="sm" pending={isInviting} closeLabel={t("actions.close")}>
          <form className="flex flex-col gap-4" onSubmit={handleInvite}>
            {inviteError ? (
              <Alert variant="destructive">
                <AlertTitle>{t("alerts.error")}</AlertTitle>
                <AlertDescription>{inviteError}</AlertDescription>
              </Alert>
            ) : null}
            <AppEmailField
              id="invite-email"
              label={t("form.email")}
              value={inviteEmail}
              onChange={(event) => setInviteEmail(event.target.value)}
              placeholder={t("form.emailPlaceholder")}
              invalidMessage={t("messages.invalidEmail")}
              requiredMessage={t("messages.provideEmail")}
              required
            />
            <AppSelectField
              label={t("form.role")}
              ariaLabel={t("form.role")}
              value={inviteRole}
              onValueChange={setInviteRole}
              triggerClassName="h-10 w-full"
              options={availableRoles.map((role) => ({
                value: role.name,
                label: resolveRoleDisplayName(role.name),
              }))}
              disabled={availableRoles.length === 0}
            />
            <Button type="submit" disabled={isInviting || !inviteRole || availableRoles.length === 0} className="w-full">
              {isInviting ? t("button.inviting") : t("button.createInvite")}
            </Button>
          </form>
      </FormSheet>

      <CommercialPlanUpgradeDialog
        open={upgradeDialogOpen}
        onOpenChange={setUpgradeDialogOpen}
        title={t("usersPage.upgradeDialog.title")}
        description={t("usersPage.upgradeDialog.description", { limit: usersLimit ?? 0 })}
        cancelLabel={t("usersPage.upgradeDialog.cancel")}
        actionLabel={t("usersPage.upgradeDialog.action")}
      />

      <FormSheet open={roleFormOpen} onOpenChange={(open) => { setRoleFormOpen(open); if (!open) setRoleError(null); }} title={t("roleForm.title")} description={selectedUserForRole?.email ?? (mobileAuthUIEnabled ? selectedUserForRole?.mobile : undefined) ?? selectedUserForRole?.id ?? ""} size="md" pending={isAssigningRoleUserID === selectedUserForRole?.id} closeLabel={t("actions.close")}>
          <OperationNotice notice={roleError ? { kind: "error", message: roleError } : null} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />
          <form className="flex flex-col gap-4" onSubmit={submitRoleForm}>
            <RadioGroup value={selectedRoleName} onValueChange={setSelectedRoleName} className="flex max-h-96 flex-col gap-3 overflow-auto pe-1">
              {availableRoles.map((role) => (
                <Label key={role.id} htmlFor={`user-role-${role.id}`} className="block cursor-pointer rounded-md border p-3 font-normal">
                  <div className="flex items-start gap-3">
                    <RadioGroupItem id={`user-role-${role.id}`} value={role.name} className="mt-1" />
                    <div>
                      <p className="text-sm font-semibold">{resolveRoleDisplayName(role.name)}</p>
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
      </FormSheet>
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

