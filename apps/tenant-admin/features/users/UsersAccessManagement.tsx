"use client";

import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";

import { AppEmailField } from "@/components/app-ui/AppEmailField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { DataTableBulkBar } from "@/components/app-ui/DataTableBulkBar";
import { DataTableToolbar } from "@/components/app-ui/DataTableToolbar";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { FilterField } from "@/components/app-ui/FilterField";
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
import { DataTable } from "@/components/ui/data-table";
import { useCalendarLocale } from "@/hooks/use-locale-format";
import { listTenantKeys, type TenantKey } from "@/lib/api/access-keys";
import type { PermissionItem } from "@/lib/contracts/roles";
import { memberDisplayName, resolveIdentity } from "@/lib/access/identity";
import { hiddenRolesInUI, resolveRoleDisplayName } from "@/lib/access/role-display";
import type { TenantAdminUser } from "@/lib/contracts/tenant-users";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { cn } from "@/lib/utils";

import { MemberDetailSheet } from "./member-detail-sheet";
import {
  filterMembers,
  hasTenantAdminRole,
  isUserActive,
  listUserRoleNames,
} from "./member-filters";
import { parseUsersURLState, patchUsersURL, type MemberStatusFilter } from "./url-state";
import {
  assignUserRole,
  getUserEntitlements,
  inviteUser,
  updateUserProfile,
  listUserRoles,
  listUsers,
  setUserActive,
  type EffectiveEntitlements,
  type RoleDefinition,
} from "./api";

type Banner = {
  kind: "success" | "error";
  message: string;
};

type UsersAccessManagementProps = {
  initialUsers: TenantAdminUser[];
  initialTotal: number;
  initialError: string | null;
  initialPermissionCatalog: PermissionItem[];
  currentUserID: string;
  canManageUsers: boolean;
};

const PAGE_SIZE = 25;
const USERS_FEATURE_KEY = "users.max";
/**
 * The control plane has no role or status filter on the user list, so the whole
 * roster is fetched once and filtered here. 200 is the server's own page cap.
 */
const ROSTER_FETCH_LIMIT = 200;

export default function UsersAccessManagement(props: UsersAccessManagementProps) {
  const t = useTranslations("accessManagement");
  const calendarLocale = useCalendarLocale();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const queryString = searchParams.toString();
  const urlState = useMemo(() => parseUsersURLState(new URLSearchParams(queryString)), [queryString]);

  const [users, setUsers] = useState<TenantAdminUser[]>(props.initialUsers);
  const [total, setTotal] = useState(props.initialTotal);
  const [requestedPage, setRequestedPage] = useState(0);
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
  const [keys, setKeys] = useState<TenantKey[]>([]);
  const [hasKeyData, setHasKeyData] = useState(false);
  const [roleFormOpen, setRoleFormOpen] = useState(false);
  const [selectedUserForRole, setSelectedUserForRole] = useState<TenantAdminUser | null>(null);
  const [selectedRoleName, setSelectedRoleName] = useState("");
  const [roleError, setRoleError] = useState<string | null>(null);
  const [bulkRoleUsers, setBulkRoleUsers] = useState<TenantAdminUser[] | null>(null);
  const [isBulkAssigning, setIsBulkAssigning] = useState(false);
  const [bulkNotice, setBulkNotice] = useState<{ kind: "success" | "warning" | "error"; message: string } | null>(null);
  const bulkClearRef = useRef<() => void>(() => {});
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteFirstName, setInviteFirstName] = useState("");
  const [inviteLastName, setInviteLastName] = useState("");
  const [inviteRole, setInviteRole] = useState("member");
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError ? { kind: "error", message: props.initialError } : null,
  );

  const patchURL = useCallback(
    (patch: Record<string, string | null | undefined>) => {
      const next = patchUsersURL(new URLSearchParams(queryString), patch);
      router.replace(next ? `${pathname}?${next}` : pathname, { scroll: false });
    },
    [pathname, queryString, router],
  );

  useEffect(() => {
    void loadUsers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    let isMounted = true;

    void Promise.allSettled([
      loadRoles(),
      getUserEntitlements().then((response) => {
        if (isMounted) setEntitlements(response);
      }),
      // Team admins and above can read the tenant key list; anyone else gets an
      // error here and the drawer simply hides its key metrics.
      listTenantKeys().then((response) => {
        if (!isMounted) return;
        setKeys(response.data ?? []);
        setHasKeyData(true);
      }),
    ]).finally(() => {
      if (isMounted) setHasLoadedEntitlements(true);
    });

    return () => {
      isMounted = false;
    };
  }, []);

  const roleFormPending =
    isBulkAssigning || (selectedUserForRole !== null && isAssigningRoleUserID === selectedUserForRole.id);
  const noticeLabels = {
    success: t("alerts.success"),
    error: t("alerts.error"),
    warning: t("alerts.error"),
    info: t("alerts.success"),
  };
  const identityLabel = mobileAuthUIEnabled ? t("table.emailMobile") : t("table.email");

  const filteredUsers = useMemo(() => filterMembers(users, urlState), [users, urlState]);
  // Clamped rather than reset on filter change, so narrowing the results cannot
  // strand the table on a page that no longer exists.
  const lastPage = Math.max(0, Math.ceil(filteredUsers.length / PAGE_SIZE) - 1);
  const page = Math.min(requestedPage, lastPage);
  const pageUsers = useMemo(
    () => filteredUsers.slice(page * PAGE_SIZE, page * PAGE_SIZE + PAGE_SIZE),
    [filteredUsers, page],
  );
  const openMember = useMemo(
    () => users.find((user) => user.id === urlState.member) ?? null,
    [users, urlState.member],
  );

  const columns: ColumnDef<TenantAdminUser>[] = [
    {
      id: "member",
      header: t("table.member"),
      meta: { label: t("table.member"), headerClassName: "w-[28%]" },
      cell: ({ row }) => {
        const displayName = memberDisplayName(row.original);
        return (
          <button
            type="button"
            className="flex min-w-0 items-center gap-2 rounded text-start outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring"
            onClick={() => patchURL({ member: row.original.id })}
          >
            <Avatar className="size-6">
              <AvatarFallback className="text-xs">{displayName.slice(0, 1).toUpperCase()}</AvatarFallback>
            </Avatar>
            <span className="truncate font-medium">{displayName || "-"}</span>
          </button>
        );
      },
    },
    {
      id: "identity",
      header: identityLabel,
      meta: { label: identityLabel, headerClassName: "w-[28%]" },
      cell: ({ row }) => {
        const identity = resolveIdentity(row.original);
        return (
          <span dir="ltr" className="block truncate font-mono text-xs" title={identity}>
            {identity}
          </span>
        );
      },
    },
    {
      id: "role",
      header: t("table.role"),
      meta: { label: t("table.role") },
      cell: ({ row }) => {
        const roleNames = listUserRoleNames(row.original);
        if (roleNames.length === 0) {
          return <Badge variant="outline">{t("member.noAccess")}</Badge>;
        }
        const extra = roleNames.length - 1;
        return (
          <span className="flex min-w-0 items-center gap-1">
            <Badge variant={isUserActive(row.original) ? "secondary" : "outline"} className="truncate">
              {resolveRoleDisplayName(roleNames[0], availableRoles, t)}
            </Badge>
            {extra > 0 ? (
              <Badge
                variant="outline"
                className="shrink-0 tabular-nums"
                title={roleNames.slice(1).map((role) => resolveRoleDisplayName(role, availableRoles, t)).join("، ")}
              >
                {t("table.moreRoles", { count: extra })}
              </Badge>
            ) : null}
          </span>
        );
      },
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
      header: () => <span className="sr-only">{t("actions.openMenu")}</span>,
      meta: { headerClassName: "w-14", cellClassName: "w-14 text-end" },
      cell: ({ row }) => {
        const user = row.original;
        const active = isUserActive(user);
        const isSelf = user.id === props.currentUserID;
        const canDeactivate = !isSelf && !hasTenantAdminRole(user);
        return (
          <RowActionsMenu
            label={t("actions.openMenu")}
            disabled={isSelf || isUpdatingUserID === user.id}
            actions={[
              {
                label: t("member.viewProfile"),
                disabled: false,
                onSelect: () => patchURL({ member: user.id }),
              },
              {
                label: t("actions.changeRole"),
                disabled:
                  !props.canManageUsers ||
                  isSelf ||
                  !user.email ||
                  availableRoles.length === 0 ||
                  isAssigningRoleUserID === user.id,
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

  const matchCount = filteredUsers.length;
  const from = matchCount === 0 ? 0 : page * PAGE_SIZE + 1;
  const to = Math.min(page * PAGE_SIZE + pageUsers.length, matchCount);
  const canPrev = page > 0;
  const canNext = (page + 1) * PAGE_SIZE < matchCount;
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

  async function loadUsers() {
    setIsLoading(true);
    try {
      const search = new URLSearchParams();
      search.set("limit", String(ROSTER_FETCH_LIMIT));
      search.set("offset", "0");

      const data = await listUsers(search);
      setUsers(data.data ?? []);
      setTotal(data.total ?? 0);
      setBanner(null);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.loadUsersError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadRoles() {
    try {
      const data = await listUserRoles();
      const nextRoles = (data.data ?? []).filter((role) => !hiddenRolesInUI.has(role.name));
      setAvailableRoles(nextRoles);
      setInviteRole((currentRole) =>
        nextRoles.some((role) => role.name === currentRole) ? currentRole : resolveDefaultInviteRole(nextRoles),
      );
    } catch {
      // Ignore role loading failure for users table.
    }
  }

  async function handleInvite(event: FormEvent<HTMLFormElement>) {
    if (!props.canManageUsers) return;
    event.preventDefault();
    setInviteError(null);

    const email = inviteEmail.trim();
    if (!email) {
      setInviteError(t("messages.provideEmail"));
      return;
    }

    setIsInviting(true);
    try {
      await inviteUser({
        email,
        role: inviteRole,
        first_name: inviteFirstName.trim(),
        last_name: inviteLastName.trim(),
      });
      showFeedbackToast({
        kind: "success",
        title: t("alerts.success"),
        description: t("messages.invitationSuccess"),
      });
      setInviteEmail("");
      setInviteFirstName("");
      setInviteLastName("");
      setInviteRole(resolveDefaultInviteRole(availableRoles));
      setInviteOpen(false);
      await loadUsers();
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
      await loadUsers();
    } catch (error) {
      showFeedbackToast({
        kind: "error",
        title: t("alerts.error"),
        description: normalizeError(error, t("messages.updateStatusError")),
      });
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
      showFeedbackToast({
        kind: "success",
        title: t("alerts.success"),
        description: t("messages.roleAssignmentSuccess"),
      });
      await loadUsers();
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
    setSelectedUserForRole(user);
    setSelectedRoleName(listUserRoleNames(user)[0] ?? "");
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
    reportBulkResult(succeeded, failed, skipped);
    await loadUsers();
  }

  // Same one-at-a-time shape as the role assignment above.
  async function runBulkStatusChange(selected: TenantAdminUser[], activate: boolean, clearSelection: () => void) {
    if (!props.canManageUsers || selected.length === 0) return;
    const eligible = selected.filter(
      (user) => user.id !== props.currentUserID && (activate || !hasTenantAdminRole(user)),
    );
    const skipped = selected.length - eligible.length;
    setIsBulkAssigning(true);
    let succeeded = 0;
    let failed = 0;
    for (const user of eligible) {
      try {
        await setUserActive(user.id, activate);
        succeeded += 1;
      } catch {
        failed += 1;
      }
    }
    setIsBulkAssigning(false);
    clearSelection();
    reportBulkResult(succeeded, failed, skipped);
    await loadUsers();
  }

  function reportBulkResult(succeeded: number, failed: number, skipped: number) {
    const summary = t("bulk.result", { succeeded, failed });
    setBulkNotice({
      kind: failed === 0 && skipped === 0 ? "success" : succeeded === 0 ? "error" : "warning",
      message: skipped > 0 ? `${summary} ${t("bulk.skipped", { count: skipped })}` : summary,
    });
  }

  return (
    <div className="flex min-w-0 flex-col gap-4">
      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
        <Button
          type="button"
          className="w-full sm:w-auto"
          onClick={handleOpenInvite}
          disabled={!props.canManageUsers || !hasLoadedEntitlements}
        >
          <Plus data-icon="inline-start" />
          {t("inviteUser")}
        </Button>
      </div>

      <OperationNotice notice={bulkNotice} labels={noticeLabels} />

      <DataTable
        tableId="users"
        columns={columns}
        data={pageUsers}
        state={isLoading && users.length === 0 ? "loading" : "ready"}
        skeletonRows={4}
        emptyLabel={t("table.noUsersFound")}
        noResultsSlot={
          <EmptyState
            variant="table"
            title={t("table.noUsersFound")}
            description={t("usersPage.emptyDescription")}
          />
        }
        enableRowSelection={props.canManageUsers}
        getRowId={(user) => user.id}
        toolbar={
          <DataTableToolbar
            search={
              <SearchField
                value={urlState.q}
                onChange={(event) => patchURL({ q: event.target.value })}
                placeholder={t("searchPlaceholder")}
              />
            }
            filters={
              <>
                <FilterField label={t("filters.role")} hideLabel>
                  <AppSelectField
                    value={urlState.role || "all"}
                    onValueChange={(value) => patchURL({ role: value === "all" ? null : value })}
                    ariaLabel={t("filters.role")}
                    triggerClassName="h-8 w-40 text-xs"
                    options={[
                      { value: "all", label: t("filters.allRoles") },
                      ...availableRoles.map((role) => ({
                        value: role.name,
                        label: resolveRoleDisplayName(role.name, availableRoles, t),
                      })),
                    ]}
                  />
                </FilterField>
                <FilterField label={t("filters.status")} hideLabel>
                  <AppSelectField
                    value={urlState.status}
                    onValueChange={(value) => patchURL({ status: value as MemberStatusFilter })}
                    ariaLabel={t("filters.status")}
                    triggerClassName="h-8 w-36 text-xs"
                    options={[
                      { value: "all", label: t("filters.allStatuses") },
                      { value: "active", label: t("status.active") },
                      { value: "deactivated", label: t("status.deactivated") },
                    ]}
                  />
                </FilterField>
              </>
            }
          />
        }
        bulkBar={(ids, clear) => {
          const selected = users.filter((user) => ids.includes(user.id));
          return (
            <DataTableBulkBar
              count={ids.length}
              onClear={clear}
              actions={[
                {
                  key: "change-role",
                  label: t("actions.changeRole"),
                  disabled: isBulkAssigning || availableRoles.length === 0,
                  onSelect: () => openBulkRoleForm(selected, clear),
                },
                {
                  key: "activate",
                  label: t("actions.activate"),
                  disabled: isBulkAssigning,
                  onSelect: () => void runBulkStatusChange(selected, true, clear),
                },
                {
                  key: "deactivate",
                  label: t("actions.deactivate"),
                  destructive: true,
                  disabled: isBulkAssigning,
                  onSelect: () => void runBulkStatusChange(selected, false, clear),
                },
              ]}
            />
          );
        }}
      />

      <TablePagination
        label={t("usersPage.title")}
        summary={t("pagination.showingResults", { from, to, total: matchCount })}
        pageSizeLabel={t("pagination.rowsPerPage")}
        pageSize={PAGE_SIZE}
        page={page + 1}
        previousLabel={t("pagination.prev")}
        nextLabel={t("pagination.next")}
        canPrevious={canPrev}
        canNext={canNext}
        disabled={isLoading}
        onPrevious={() => setRequestedPage(Math.max(0, page - 1))}
        onNext={() => setRequestedPage(Math.min(lastPage, page + 1))}
      />

      <MemberDetailSheet
        member={openMember}
        roles={availableRoles}
        permissionCatalog={props.initialPermissionCatalog}
        keys={keys}
        hasKeyData={hasKeyData}
        canManageUsers={props.canManageUsers}
        isSelf={openMember?.id === props.currentUserID}
        onOpenChange={(open) => {
          if (!open) patchURL({ member: null });
        }}
        onSaveName={async (member, firstName, lastName) => {
          await updateUserProfile(member.id, { first_name: firstName, last_name: lastName });
          showFeedbackToast({
            kind: "success",
            title: t("alerts.success"),
            description: t("messages.nameUpdated"),
          });
          await loadUsers();
        }}
        onChangeRole={(member) => {
          patchURL({ member: null });
          openRoleForm(member);
        }}
        onToggleActive={(member, activate) => {
          patchURL({ member: null });
          void changeUserStatus(member.id, activate);
        }}
      />

      <FormSheet
        open={inviteOpen}
        onOpenChange={handleInviteOpenChange}
        title={t("inviteTitle")}
        description={t("inviteDescription")}
        size="sm"
        pending={isInviting}
        closeLabel={t("actions.close")}
        footer={
          <FormSheetActions
            secondary={
              <Button type="button" variant="outline" disabled={isInviting} onClick={() => handleInviteOpenChange(false)}>
                {t("actions.cancel")}
              </Button>
            }
            primary={
              <Button
                type="submit"
                form="invite-user-form"
                disabled={isInviting || !inviteRole || availableRoles.length === 0}
              >
                {isInviting ? t("button.inviting") : t("button.createInvite")}
              </Button>
            }
          />
        }
      >
        <form id="invite-user-form" className="flex flex-col gap-4" onSubmit={handleInvite}>
          {inviteError ? (
            <Alert variant="destructive">
              <AlertTitle>{t("alerts.error")}</AlertTitle>
              <AlertDescription>{inviteError}</AlertDescription>
            </Alert>
          ) : null}
          <div className="grid gap-4 sm:grid-cols-2">
            <AppTextField
              id="invite-first-name"
              label={t("form.firstName")}
              value={inviteFirstName}
              onChange={(event) => setInviteFirstName(event.target.value)}
              autoComplete="given-name"
              maxLength={100}
            />
            <AppTextField
              id="invite-last-name"
              label={t("form.lastName")}
              value={inviteLastName}
              onChange={(event) => setInviteLastName(event.target.value)}
              autoComplete="family-name"
              maxLength={100}
            />
          </div>
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
              label: resolveRoleDisplayName(role.name, availableRoles, t),
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
        onOpenChange={(open) => {
          setRoleFormOpen(open);
          if (!open) {
            setRoleError(null);
            setBulkRoleUsers(null);
          }
        }}
        title={bulkRoleUsers ? t("bulk.changeRoleTitle") : t("roleForm.title")}
        description={
          bulkRoleUsers
            ? t("bulk.changeRoleDescription", { count: bulkRoleUsers.length })
            : selectedUserForRole?.email ??
              (mobileAuthUIEnabled ? selectedUserForRole?.mobile : undefined) ??
              selectedUserForRole?.id ??
              ""
        }
        size="md"
        pending={roleFormPending}
        closeLabel={t("actions.close")}
        footer={
          <FormSheetActions
            secondary={
              <Button
                type="button"
                variant="outline"
                disabled={roleFormPending}
                onClick={() => {
                  setRoleFormOpen(false);
                  setRoleError(null);
                  setBulkRoleUsers(null);
                }}
              >
                {t("actions.cancel")}
              </Button>
            }
            primary={
              <Button
                type="submit"
                form="assign-user-role-form"
                disabled={
                  !props.canManageUsers ||
                  !(selectedUserForRole || bulkRoleUsers) ||
                  !selectedRoleName ||
                  roleFormPending
                }
              >
                {roleFormPending ? t("roleForm.submitting") : t("roleForm.submit")}
              </Button>
            }
          />
        }
      >
        <OperationNotice notice={roleError ? { kind: "error", message: roleError } : null} labels={noticeLabels} />
        <form id="assign-user-role-form" className="flex flex-col gap-4" onSubmit={submitRoleForm}>
          <RadioGroup
            value={selectedRoleName}
            onValueChange={setSelectedRoleName}
            className="flex max-h-96 flex-col gap-3 overflow-auto pe-1"
          >
            {availableRoles.map((role) => (
              <Label
                key={role.id}
                htmlFor={`user-role-${role.id}`}
                className="block cursor-pointer rounded-md border p-3 font-normal"
              >
                <div className="flex items-start gap-3">
                  <RadioGroupItem id={`user-role-${role.id}`} value={role.name} className="mt-1" />
                  <div>
                    <p className="text-sm font-semibold">{resolveRoleDisplayName(role.name, availableRoles, t)}</p>
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
