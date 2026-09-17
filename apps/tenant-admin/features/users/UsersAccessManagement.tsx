"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";
import { Info, Plus, UsersRound } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppEmailField } from "@/components/app-ui/AppEmailField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { DataTableBulkBar } from "@/components/app-ui/DataTableBulkBar";
import { DataTableToolbar } from "@/components/app-ui/DataTableToolbar";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { SearchField } from "@/components/app-ui/SearchField";
import { TablePagination } from "@/components/app-ui/TablePagination";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Progress } from "@/components/ui/progress";
import { DataTable } from "@/components/ui/data-table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useCalendarLocale } from "@/hooks/use-locale-format";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { cn } from "@/lib/utils";
import {
  assignUserRole,
  getUserEntitlements,
  inviteUser,
  listUserRoles,
  listUsers,
  setUserActive,
  type EffectiveEntitlements,
  type RoleDefinition,
  type TenantAdminUser,
} from "./api";

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
  const calendarLocale = useCalendarLocale();

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
  const [bulkRoleUsers, setBulkRoleUsers] = useState<TenantAdminUser[] | null>(null);
  const [isBulkAssigning, setIsBulkAssigning] = useState(false);
  const [bulkNotice, setBulkNotice] = useState<{ kind: "success" | "warning" | "error"; message: string } | null>(null);
  const bulkClearRef = useRef<() => void>(() => {});
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
      getUserEntitlements().then((response) => {
        if (isMounted) setEntitlements(response);
      }),
    ]).finally(() => {
      if (isMounted) setHasLoadedEntitlements(true);
    });

    return () => {
      isMounted = false;
    };
  }, []);

  const roleFormPending = isBulkAssigning || (selectedUserForRole !== null && isAssigningRoleUserID === selectedUserForRole.id);
  const noticeLabels = { success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") };
  const identityLabel = mobileAuthUIEnabled ? t("table.emailMobile") : t("table.email");

  const columns: ColumnDef<TenantAdminUser>[] = [
    {
      id: "member",
      header: t("table.member"),
      meta: { label: t("table.member"), headerClassName: "w-[28%]" },
      cell: ({ row }) => {
        const displayName = guessDisplayName(resolveIdentity(row.original));
        return (
          <div className="flex min-w-0 items-center gap-2">
            <Avatar className="size-6">
              <AvatarFallback className="text-xs">{displayName.slice(0, 1).toUpperCase()}</AvatarFallback>
            </Avatar>
            <span className="truncate font-medium">{displayName || "-"}</span>
          </div>
        );
      },
    },
    {
      id: "identity",
      header: identityLabel,
      meta: { label: identityLabel, headerClassName: "w-[28%]" },
      cell: ({ row }) => {
        const identity = resolveIdentity(row.original);
        return <span dir="ltr" className="block truncate font-mono text-xs" title={identity}>{identity}</span>;
      },
    },
    {
      id: "role",
      header: t("table.role"),
      meta: { label: t("table.role") },
      cell: ({ row }) => (
        <Badge variant={isUserActive(row.original) ? "secondary" : "outline"}>
          {resolveRoleDisplayName(row.original.roles?.[0] ?? row.original.role)}
        </Badge>
      ),
    },
    {
      id: "status",
      header: t("table.status"),
      meta: { label: t("table.status") },
      cell: ({ row }) => {
        const active = isUserActive(row.original);
        return (
          <span className="inline-flex items-center gap-1.5">
            <span aria-hidden="true" className={cn("size-2 rounded-full", active ? "bg-success" : "bg-warning")} />
            {active ? t("status.active") : t("status.deactivated")}
          </span>
        );
      },
    },
    {
      id: "lastActivity",
      header: t("table.lastActivity"),
      meta: { optional: true, label: t("table.lastActivity") },
      cell: ({ row }) => (
        <DateTimeValue
          value={row.original.last_login_at}
          locale={calendarLocale}
          fallback={t("table.neverActive")}
          className="text-muted-foreground"
        />
      ),
    },
    {
      id: "actions",
      header: () => <span className="sr-only">{t("actions.changeRole")}</span>,
      meta: { headerClassName: "w-14", cellClassName: "w-14 text-end" },
      cell: ({ row }) => {
        const user = row.original;
        const active = isUserActive(user);
        const isSelf = user.id === props.currentUserID;
        const canDeactivate = !isSelf && !hasTenantAdminRole(user);
        return (
          <RowActionsMenu
            label={t("actions.changeRole")}
            disabled={isSelf || isUpdatingUserID === user.id}
            actions={[
              {
                label: t("actions.changeRole"),
                disabled: !props.canManageUsers || isSelf || !user.email || availableRoles.length === 0 || isAssigningRoleUserID === user.id,
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
        );
      },
    },
  ];

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

      const data = await listUsers(search);
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
      const data = await listUserRoles();
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
      const response = await inviteUser({ email, role: inviteRole });
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
      await setUserActive(userID, activate);
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
      await assignUserRole(email, roleName);
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
    if (bulkRoleUsers) {
      await submitBulkRoleForm();
      return;
    }
    if (!selectedUserForRole || !selectedRoleName) return;
    const succeeded = await assignRole(selectedUserForRole, selectedRoleName);
    if (!succeeded) return;
    setRoleFormOpen(false);
    setSelectedUserForRole(null);
  }

  function openBulkRoleForm(selected: TenantAdminUser[], clearSelection: () => void) {
    if (!props.canManageUsers || selected.length === 0) return;
    bulkClearRef.current = clearSelection;
    setBulkRoleUsers(selected);
    setSelectedUserForRole(null);
    setSelectedRoleName("");
    setRoleError(null);
    setRoleFormOpen(true);
  }

  // No bulk endpoint yet: assign one user at a time and report the totals.
  async function submitBulkRoleForm() {
    if (!props.canManageUsers || !bulkRoleUsers || !selectedRoleName) return;
    const eligible = bulkRoleUsers.filter((user) => user.id !== props.currentUserID && (user.email ?? "").trim());
    const skipped = bulkRoleUsers.length - eligible.length;
    setIsBulkAssigning(true);
    let succeeded = 0;
    let failed = 0;
    for (const user of eligible) {
      try {
        await assignUserRole((user.email ?? "").trim(), selectedRoleName);
        succeeded += 1;
      } catch {
        failed += 1;
      }
    }
    setIsBulkAssigning(false);
    setRoleFormOpen(false);
    setBulkRoleUsers(null);
    bulkClearRef.current();
    const summary = t("bulk.result", { succeeded, failed });
    setBulkNotice({
      kind: failed === 0 && skipped === 0 ? "success" : succeeded === 0 ? "error" : "warning",
      message: skipped > 0 ? `${summary} ${t("bulk.skipped", { count: skipped })}` : summary,
    });
    await loadUsers(offset, query.trim());
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
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
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

        <OperationNotice notice={bulkNotice} labels={noticeLabels} />

        <DataTable
          tableId="users"
          columns={columns}
          data={users}
          state={isLoading && users.length === 0 ? "loading" : "ready"}
          skeletonRows={4}
          emptyLabel={t("table.noUsersFound")}
          noResultsSlot={<EmptyState variant="table" title={t("table.noUsersFound")} description={t("usersPage.emptyDescription")} />}
          enableRowSelection={props.canManageUsers}
          getRowId={(user) => user.id}
          toolbar={(
            <DataTableToolbar
              search={(
                <SearchField
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder={t("searchPlaceholder")}
                />
              )}
            />
          )}
          bulkBar={(ids, clear) => (
            <DataTableBulkBar
              count={ids.length}
              onClear={clear}
              actions={[
                {
                  key: "change-role",
                  label: t("actions.changeRole"),
                  disabled: isBulkAssigning || availableRoles.length === 0,
                  onSelect: () => openBulkRoleForm(users.filter((user) => ids.includes(user.id)), clear),
                },
              ]}
            />
          )}
        />

        <TablePagination
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

      <FormSheet
        open={inviteOpen}
        onOpenChange={handleInviteOpenChange}
        title={t("inviteTitle")}
        description={t("inviteDescription")}
        size="sm"
        pending={isInviting}
        closeLabel={t("actions.close")}
        footer={(
          <FormSheetActions
            secondary={<Button type="button" variant="outline" disabled={isInviting} onClick={() => handleInviteOpenChange(false)}>{t("actions.cancel")}</Button>}
            primary={<Button type="submit" form="invite-user-form" disabled={isInviting || !inviteRole || availableRoles.length === 0}>{isInviting ? t("button.inviting") : t("button.createInvite")}</Button>}
          />
        )}
      >
          <form id="invite-user-form" className="flex flex-col gap-4" onSubmit={handleInvite}>
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

      <FormSheet
        open={roleFormOpen}
        onOpenChange={(open) => { setRoleFormOpen(open); if (!open) { setRoleError(null); setBulkRoleUsers(null); } }}
        title={bulkRoleUsers ? t("bulk.changeRoleTitle") : t("roleForm.title")}
        description={bulkRoleUsers
          ? t("bulk.changeRoleDescription", { count: bulkRoleUsers.length })
          : selectedUserForRole?.email ?? (mobileAuthUIEnabled ? selectedUserForRole?.mobile : undefined) ?? selectedUserForRole?.id ?? ""}
        size="md"
        pending={roleFormPending}
        closeLabel={t("actions.close")}
        footer={(
          <FormSheetActions
            secondary={<Button type="button" variant="outline" disabled={roleFormPending} onClick={() => { setRoleFormOpen(false); setRoleError(null); setBulkRoleUsers(null); }}>{t("actions.cancel")}</Button>}
            primary={<Button type="submit" form="assign-user-role-form" disabled={!props.canManageUsers || !(selectedUserForRole || bulkRoleUsers) || !selectedRoleName || roleFormPending}>{roleFormPending ? t("roleForm.submitting") : t("roleForm.submit")}</Button>}
          />
        )}
      >
          <OperationNotice notice={roleError ? { kind: "error", message: roleError } : null} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />
          <form id="assign-user-role-form" className="flex flex-col gap-4" onSubmit={submitRoleForm}>
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
          </form>
      </FormSheet>
    </div>
  );
}

function resolveIdentity(user: TenantAdminUser) {
  return user.email || (mobileAuthUIEnabled ? user.mobile : "") || user.id;
}

function isUserActive(user: TenantAdminUser) {
  return user.active && user.role !== "deactivated";
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

