"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  AlertTriangle,
  Plus,
  RefreshCw,
  Trash2,
  Users,
} from "lucide-react";

import { usePermissions } from "@/app/components/permissions-context";
import { AppCard, AppSelectField, CheckboxOptionList, CommercialPlanUpgradeDialog, DateTimeValue, EmptyState, FilterCombobox, FormSheet, IdentifierField, MetricCard, OperationNotice, PageHeader, RowActionsMenu, SearchField, StatusBadge as AppStatusBadge, TableEmptyRow, TableLoadingRows, TablePagination, TypedConfirmDialog } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { canAnyPermission } from "@/lib/authz";
import { cn } from "@/lib/utils";

type Team = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  org_unit_id?: string | null;
  workspace_type?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};
type TeamMember = {
  id?: string;
  user_id: string;
  user_display_name?: string;
  user_email?: string;
  user_mobile?: string;
  role: string;
  team_member_permissions?: string[];
  metadata?: Record<string, unknown>;
  updated_at?: string;
};
type OrgUnit = { id: string; alias: string; display_name?: string };
type AccessPrincipalCandidate = { principal_id: string; display_name: string; secondary?: string };
type MemberCandidateOption = { userID: string; label: string };
type ListResponse<T> = { data?: T[]; total?: number };
type Banner = { kind: "success" | "error" | "warning"; message: string };
type EffectiveEntitlements = { features?: Record<string, unknown> } | null;

type TeamForm = {
  alias: string;
  displayName: string;
  orgUnitID: string;
  workspaceType: string;
  description: string;
};

const PERMISSIONS = ["/key/info", "/key/list", "/key/generate", "/key/update", "/key/delete", "/key/regenerate", "/key/block", "/key/unblock", "/key/health"] as const;
const DEFAULT_PERMISSIONS = ["/key/info", "/key/health"];
const WORKSPACE_TYPES = ["production", "staging", "development", "research", "internal", "general"] as const;
const WORKSPACE_TYPE_SET = new Set<string>(WORKSPACE_TYPES);
const PAGE_SIZE = 25;
const FEATURE_TEAMS_ENABLED = "teams.enabled";

const emptyForm: TeamForm = {
  alias: "",
  displayName: "",
  orgUnitID: "",
  workspaceType: "production",
  description: "",
};

export default function TeamManagement({
  initialTeams,
  initialTotal,
  initialError,
  initialOrgUnits,
  currentUserRole,
}: {
  initialTeams: Team[];
  initialTotal: number;
  initialError: string | null;
  initialOrgUnits: OrgUnit[];
  currentUserRole: string;
}) {
  const t = useTranslations("teams");
  const locale = useLocale();
  const permissions = usePermissions() ?? [];
  const [teams, setTeams] = useState(initialTeams);
  const [total, setTotal] = useState(initialTotal);
  const [orgUnits] = useState(initialOrgUnits);
  const [query, setQuery] = useState("");
  const [orgFilter, setOrgFilter] = useState("all");
  const [workspaceFilter, setWorkspaceFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [sortBy, setSortBy] = useState("recent");
  const [offset, setOffset] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
  const [isUpgradeDialogOpen, setIsUpgradeDialogOpen] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(initialError ? { kind: "error", message: initialError } : null);

  const [selectedTeamID, setSelectedTeamID] = useState<string | null>(initialTeams[0]?.id ?? null);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(initialTeams[0] ?? null);
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [users, setUsers] = useState<AccessPrincipalCandidate[]>([]);
  const [detailLoading, setDetailLoading] = useState(false);

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [membersOpen, setMembersOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [form, setForm] = useState<TeamForm>(emptyForm);
  const [memberUserInput, setMemberUserInput] = useState("");
  const [memberRole, setMemberRole] = useState("team_member");
  const [memberPerms, setMemberPerms] = useState<string[]>(DEFAULT_PERMISSIONS);

  const canManage = ["tenant_admin", "proxy_admin", "org_admin", "team_admin"].includes(currentUserRole)
    || canAnyPermission(permissions, ["team:manage"]);
  const canCreateTeam = canAnyPermission(permissions, ["team:create"]);
  const canDelete = canAnyPermission(permissions, ["team:delete"]);
  const isTeamsFeatureEnabled = hasLoadedEntitlements
    ? readBooleanFeature(entitlements?.features?.[FEATURE_TEAMS_ENABLED], false)
    : null;

  function handleCreateTeam() {
    if (isTeamsFeatureEnabled === false) {
      setIsUpgradeDialogOpen(true);
      return;
    }
    setForm(emptyForm);
    setCreateOpen(true);
  }

  useEffect(() => {
    let isMounted = true;

    void requestJSON<EffectiveEntitlements>("/api/entitlements")
      .then((response) => {
        if (isMounted) setEntitlements(response);
      })
      .catch(() => {
        if (isMounted) setEntitlements(null);
      })
      .finally(() => {
        if (isMounted) setHasLoadedEntitlements(true);
      });

    return () => {
      isMounted = false;
    };
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      setOffset(0);
      void loadTeams(0, query.trim());
    }, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  useEffect(() => {
    if (selectedTeamID) void loadTeamDetails(selectedTeamID);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedTeamID]);

  const filtered = useMemo(() => {
    const list = teams.filter((team) => {
      if (orgFilter !== "all" && (team.org_unit_id ?? "root") !== orgFilter) return false;
      if (workspaceFilter !== "all" && getWorkspaceType(team) !== workspaceFilter) return false;
      if (statusFilter !== "all" && getStatus(team) !== statusFilter) return false;
      return true;
    });
    return [...list].sort((a, b) => {
      if (sortBy === "name") return getTeamName(a).localeCompare(getTeamName(b));
      if (sortBy === "members") return getMemberCount(b) - getMemberCount(a);
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });
  }, [teams, orgFilter, workspaceFilter, statusFilter, sortBy]);

  const memberCandidateOptions = useMemo<MemberCandidateOption[]>(() => {
    const memberIDs = new Set(members.map((member) => member.user_id));
    return users.filter((user) => !memberIDs.has(user.principal_id)).map((user) => ({
      userID: user.principal_id,
      label: formatMemberCandidateLabel(user),
    }));
  }, [members, users]);
  const memberLabelByUserID = useMemo(() => {
    return new Map(users.map((user) => [user.principal_id, formatMemberCandidateLabel(user)]));
  }, [users]);

  async function loadTeams(nextOffset: number, q: string) {
    setIsLoading(true);
    try {
      const params = new URLSearchParams();
      params.set("limit", String(PAGE_SIZE));
      params.set("offset", String(nextOffset));
      if (q) params.set("query", q);
      const payload = await requestJSON<ListResponse<Team>>(`/api/teams?${params.toString()}`);
      const next = payload.data ?? [];
      setTeams(next);
      setTotal(payload.total ?? next.length);
      setOffset(nextOffset);
      if (!selectedTeamID || !next.some((item) => item.id === selectedTeamID)) {
        setSelectedTeamID(next[0]?.id ?? null);
      }
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.loadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadTeamDetails(teamID: string) {
    setDetailLoading(true);
    try {
      const [team, memberPayload] = await Promise.all([
        requestJSON<Team>(`/api/teams/${encodeURIComponent(teamID)}`),
        requestJSON<ListResponse<TeamMember>>(`/api/teams/${encodeURIComponent(teamID)}/members`),
      ]);
      let userPayload: ListResponse<AccessPrincipalCandidate> = { data: [] };
      try {
        // Keep member labels visible for read-only users when principal search is permitted.
        userPayload = await requestJSON<ListResponse<AccessPrincipalCandidate>>("/api/access-principals?kind=user&limit=100");
      } catch {
        // Ignore lookup failures so team details and member list still render.
      }
      setSelectedTeam(team);
      setMembers(memberPayload.data ?? []);
      setUsers(userPayload.data ?? []);
      setTeams((prev) => prev.map((item) => (item.id === team.id ? team : item)));
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.detailLoadError")) });
    } finally {
      setDetailLoading(false);
    }
  }

  async function loadUsers() {
    const payload = await requestJSON<ListResponse<AccessPrincipalCandidate>>("/api/access-principals?kind=user&limit=100");
    setUsers(payload.data ?? []);
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateForm(form, teams, t);
    if (validation) return setBanner({ kind: "error", message: validation });
    setIsSaving(true);
    try {
      const created = await requestJSON<Team>("/api/teams", { method: "POST", body: JSON.stringify(buildTeamPayload(form)) });
      setCreateOpen(false);
      setForm(emptyForm);
      setBanner({ kind: "success", message: t("messages.createSuccess") });
      await loadTeams(0, query.trim());
      setSelectedTeamID(created.id);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.createError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function handleEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedTeam) return;
    const validation = validateForm(form, teams, t, selectedTeam.id);
    if (validation) return setBanner({ kind: "error", message: validation });
    setIsSaving(true);
    try {
      const updated = await requestJSON<Team>(`/api/teams/${encodeURIComponent(selectedTeam.id)}`, {
        method: "PATCH",
        body: JSON.stringify(buildTeamPayload(form, selectedTeam.metadata)),
      });
      setSelectedTeam(updated);
      setTeams((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      setEditOpen(false);
      setBanner({ kind: "success", message: t("messages.updateSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.updateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function addMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedTeam) return;
    const resolvedUserID = resolveMemberUserID(memberUserInput, memberCandidateOptions);
    if (!resolvedUserID) return setBanner({ kind: "error", message: t("messages.selectUser") });
    setIsSaving(true);
    try {
      await requestJSON(`/api/teams/${encodeURIComponent(selectedTeam.id)}/members`, {
        method: "POST",
        body: JSON.stringify({
          user_id: resolvedUserID,
          role: memberRole,
          team_member_permissions: memberRole === "team_member" ? memberPerms : PERMISSIONS,
          metadata: {},
        }),
      });
      setMemberUserInput("");
      setMemberRole("team_member");
      setMemberPerms(DEFAULT_PERMISSIONS);
      await loadTeamDetails(selectedTeam.id);
      setBanner({ kind: "success", message: t("messages.memberAddSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.memberAddError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function updateMember(member: TeamMember, role: string, perms: string[]) {
    if (!selectedTeam) return;
    setIsSaving(true);
    try {
      await requestJSON(`/api/teams/${encodeURIComponent(selectedTeam.id)}/members/${encodeURIComponent(member.user_id)}`, {
        method: "PATCH",
        body: JSON.stringify({
          role,
          team_member_permissions: role === "team_member" ? perms : PERMISSIONS,
        }),
      });
      await loadTeamDetails(selectedTeam.id);
      setBanner({ kind: "success", message: t("messages.memberUpdateSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.memberUpdateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function removeMember(userID: string) {
    if (!selectedTeam) return;
    setIsSaving(true);
    try {
      await requestJSON(`/api/teams/${encodeURIComponent(selectedTeam.id)}/members/${encodeURIComponent(userID)}`, { method: "DELETE" });
      await loadTeamDetails(selectedTeam.id);
      setBanner({ kind: "success", message: t("messages.memberRemoveSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.memberRemoveError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function handleDelete() {
    if (!selectedTeam) return;
    setIsSaving(true);
    try {
      await requestJSON(`/api/teams/${encodeURIComponent(selectedTeam.id)}`, { method: "DELETE" });
      setBanner({ kind: "success", message: t("messages.deleteSuccess") });
      await loadTeams(0, query.trim());
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.deleteError")) });
      throw error;
    } finally {
      setIsSaving(false);
    }
  }

  const summary = selectedTeam ? buildSummary(selectedTeam, members) : null;
  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + teams.length, total);

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("title")}
        description={t("subtitle")}
        actions={
          <>
            <Button variant="outline" onClick={() => void loadTeams(offset, query.trim())} disabled={isLoading}>
              {isLoading ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
              {t("actions.refresh")}
            </Button>
            <Button
              type="button"
              onClick={handleCreateTeam}
              disabled={!canCreateTeam || !hasLoadedEntitlements}
            >
              <Plus data-icon="inline-start" />
              {t("actions.new")}
            </Button>
          </>
        }
      />

      <OperationNotice notice={banner} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} />

      <section className="grid min-w-0 gap-6 2xl:grid-cols-[minmax(0,1.2fr)_minmax(22rem,0.8fr)]">
        <AppCard className="min-w-0 shadow-none" contentClassName="flex min-w-0 flex-col gap-4">
          <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-[minmax(14rem,1fr)_repeat(4,auto)]">
            <SearchField containerClassName="min-w-0" value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t("filters.search")} />
            <AppSelectField ariaLabel={t("filters.orgAll")} value={orgFilter} onValueChange={setOrgFilter} triggerClassName="h-10 w-full" options={[{ value: "all", label: t("filters.orgAll") }, { value: "root", label: t("table.tenantRoot") }, ...orgUnits.map((org) => ({ value: org.id, label: org.display_name || org.alias }))]} />
            <AppSelectField ariaLabel={t("filters.workspaceAll")} value={workspaceFilter} onValueChange={setWorkspaceFilter} triggerClassName="h-10 w-full" options={[{ value: "all", label: t("filters.workspaceAll") }, ...WORKSPACE_TYPES.map((type) => ({ value: type, label: t(`workspace.${type}`) }))]} />
            <AppSelectField ariaLabel={t("filters.statusAll")} value={statusFilter} onValueChange={setStatusFilter} triggerClassName="h-10 w-full" options={[{ value: "all", label: t("filters.statusAll") }, { value: "active", label: t("status.active") }, { value: "blocked", label: t("status.blocked") }]} />
            <AppSelectField ariaLabel={t("filters.sortLabel")} value={sortBy} onValueChange={setSortBy} triggerClassName="h-10 w-full" options={[{ value: "recent", label: t("sort.recent") }, { value: "name", label: t("sort.name") }, { value: "members", label: t("sort.members") }]} />
          </div>

          <div className="overflow-x-auto">
            <Table>
              <TableHeader><TableRow><TableHead>{t("table.team")}</TableHead><TableHead>{t("table.orgUnit")}</TableHead><TableHead>{t("table.workspaceType")}</TableHead><TableHead>{t("table.status")}</TableHead><TableHead>{t("table.updated")}</TableHead><TableHead className="w-10 text-end"> </TableHead></TableRow></TableHeader>
              <TableBody>
                {isLoading && filtered.length === 0 ? <TableLoadingRows colSpan={6} label={t("table.loading")} /> : filtered.length === 0 ? <TableEmptyRow colSpan={6} icon={Users} title={t("empty.title")} description={t("empty.description")} primaryAction={canCreateTeam ? <Button type="button" onClick={handleCreateTeam} disabled={!hasLoadedEntitlements}><Plus data-icon="inline-start" />{t("actions.new")}</Button> : undefined} /> : filtered.map((team) => (
                  <TableRow key={team.id} className={cn("cursor-pointer", selectedTeamID === team.id && "bg-muted/50")} onClick={() => setSelectedTeamID(team.id)}>
                    <TableCell><div className="min-w-44"><p className="truncate font-medium">{getTeamName(team)}</p><p className="truncate text-xs text-muted-foreground">{team.alias}</p></div></TableCell>
                    <TableCell>{getOrgUnitName(team, orgUnits, t)}</TableCell>
                    <TableCell><Badge variant="outline">{t(`workspace.${getWorkspaceType(team)}`)}</Badge></TableCell>
                    <TableCell>{getStatus(team) === "blocked" ? <AppStatusBadge status="error">{t("status.blocked")}</AppStatusBadge> : <AppStatusBadge status="active">{t("status.active")}</AppStatusBadge>}</TableCell>
                    <TableCell><DateTimeValue value={team.updated_at} locale={locale} format={{ month: "short", day: "numeric" }} /></TableCell>
                    <TableCell className="text-end">
                      <RowActionsMenu
                        label={t("table.actions")}
                        actions={[
                          { label: t("actions.edit"), disabled: !canManage, onSelect: () => { setForm(formFromTeam(team)); setEditOpen(true); } },
                          { label: t("actions.members"), disabled: !canManage, onSelect: () => { setSelectedTeamID(team.id); setMembersOpen(true); void loadUsers(); } },
                          ...(canDelete ? [{ label: t("actions.delete"), destructive: true, icon: <Trash2 />, onSelect: () => { setSelectedTeamID(team.id); setDeleteOpen(true); } }] : []),
                        ]}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          <TablePagination
            label={t("title")}
            summary={t("pagination.showing", { from, to, total })}
            previousLabel={t("pagination.prev")}
            nextLabel={t("pagination.next")}
            canPrevious={offset > 0}
            canNext={offset + teams.length < total}
            disabled={isLoading}
            onPrevious={() => void loadTeams(Math.max(0, offset - PAGE_SIZE), query.trim())}
            onNext={() => void loadTeams(offset + PAGE_SIZE, query.trim())}
          />
        </AppCard>

        <div className="flex min-w-0 flex-col gap-4">
          {!selectedTeam || !summary ? <EmptyState icon={Users} title={t("detail.empty")} className="min-h-56" /> : <>
            <div className="grid grid-cols-2 gap-3">
              <MetricCard label={t("kpi.membersAdmins")} value={`${summary.members}/${summary.admins}`} density="compact" className="shadow-none" />
              <MetricCard label={t("kpi.activeKeys")} value={String(summary.activeKeys)} density="compact" className="shadow-none" />
            </div>
            <AppCard title={t("detail.access")} contentClassName="flex flex-col gap-2 text-sm">
              <p>{t("detail.orgInheritance", { value: getOrgUnitName(selectedTeam, orgUnits, t) })}</p><p>{t("detail.permissions", { value: summarizePermissions(members, t) })}</p>
            </AppCard>
            <AppCard title={t("detail.members")} contentClassName="flex flex-col gap-2 text-sm">
              {detailLoading ? <div className="flex flex-col gap-2" aria-label={t("detail.loading")}><Skeleton className="h-12 w-full" /><Skeleton className="h-12 w-full" /></div> : members.length === 0 ? <p className="text-muted-foreground">{t("detail.noMembers")}</p> : members.map((member) => <div key={member.user_id} className="flex items-center justify-between gap-2 rounded-md border border-border p-2"><span className="truncate" dir="auto">{getMemberDisplayLabel(member, memberLabelByUserID, t("members.unknownUser"))}</span><Badge variant={member.role === "team_admin" ? "default" : "secondary"}>{member.role}</Badge></div>)}
            </AppCard>
          </>}
        </div>
      </section>

      <FormSheet open={createOpen} onOpenChange={setCreateOpen} title={t("create.title")} description={t("create.description")} pending={isSaving}>
        <TeamForm t={t} form={form} setForm={setForm} orgUnits={orgUnits} onSubmit={handleCreate} isSaving={isSaving} submitLabel={t("create.submit")} />
      </FormSheet>

      <FormSheet open={editOpen} onOpenChange={setEditOpen} title={t("edit.title")} description={t("edit.description")} pending={isSaving}>
        <TeamForm t={t} form={form} setForm={setForm} orgUnits={orgUnits} onSubmit={handleEdit} isSaving={isSaving} submitLabel={t("edit.submit")} />
      </FormSheet>

      <Sheet open={membersOpen} onOpenChange={setMembersOpen}><SheetContent side="left" className="w-full max-w-xl overflow-y-auto sm:w-3/4"><SheetHeader><SheetTitle>{t("members.title")}</SheetTitle><SheetDescription>{t("members.description")}</SheetDescription></SheetHeader>
        <form onSubmit={addMember} className="mt-6 flex flex-col gap-3">
          <FieldGroup className="gap-3">
            <Field>
              <FieldLabel htmlFor="team-member-user">{t("members.user")}</FieldLabel>
              <FilterCombobox id="team-member-user" value={memberUserInput} onValueChange={setMemberUserInput} options={memberCandidateOptions.map((user) => ({ value: user.userID, label: user.label }))} placeholder={t("members.userPlaceholder")} emptyText={t("members.userEmpty")} className="h-10" />
              <FieldDescription>{t("members.userHelp")}</FieldDescription>
            </Field>
            <AppSelectField label={t("members.role")} value={memberRole} onValueChange={setMemberRole} triggerClassName="h-10" options={[{ value: "team_member", label: t("members.teamMember") }, { value: "team_admin", label: t("members.teamAdmin") }]} />
          </FieldGroup>
          {memberRole === "team_member" ? <div className="flex flex-col gap-3"><div className="flex justify-end"><Button type="button" variant="ghost" size="sm" onClick={() => setMemberPerms(DEFAULT_PERMISSIONS)}>{t("members.resetPermissions")}</Button></div><CheckboxOptionList label={t("members.permissions")} description={t("members.permissionsHelp")} value={memberPerms} onValueChange={setMemberPerms} emptyText={t("members.permissions")} options={PERMISSIONS.map((perm) => ({ value: perm, label: t(`members.permissionOptions.${permissionTranslationKey(perm)}.label`), description: t(`members.permissionOptions.${permissionTranslationKey(perm)}.description`) }))} /></div> : <Alert><AlertTitle>{t("members.teamAdminAccessTitle")}</AlertTitle><AlertDescription>{t("members.teamAdminAccessDescription")}</AlertDescription></Alert>}
          <Button type="submit" className="w-full" disabled={!canManage || isSaving}>{t("members.add")}</Button>
        </form>
        <div className="mt-6 flex flex-col gap-2">{members.map((member) => <div key={member.user_id} className="rounded-md border border-border p-3"><div className="flex items-center justify-between gap-2"><span className="truncate text-sm" dir="auto">{getMemberDisplayLabel(member, memberLabelByUserID, t("members.unknownUser"))}</span><Badge variant={member.role === "team_admin" ? "default" : "secondary"}>{member.role}</Badge></div><div className="mt-2 flex flex-wrap gap-2"><Button size="sm" variant="outline" onClick={() => void updateMember(member, member.role === "team_admin" ? "team_member" : "team_admin", member.team_member_permissions ?? DEFAULT_PERMISSIONS)} disabled={!canManage || isSaving}>{member.role === "team_admin" ? t("members.makeTeamMember") : t("members.makeTeamAdmin")}</Button><Button size="sm" variant="outline" className="text-destructive" onClick={() => void removeMember(member.user_id)} disabled={!canManage || isSaving}>{t("members.remove")}</Button></div></div>)}</div>
      </SheetContent></Sheet>

      {selectedTeam ? (
        <TypedConfirmDialog
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          title={t("delete.title")}
          description={t("delete.description")}
          expectedValue={selectedTeam.alias}
          inputLabel={t("delete.confirm", { alias: selectedTeam.alias })}
          cancelLabel={t("actions.cancel")}
          confirmLabel={t("delete.submit")}
          pendingLabel={t("actions.saving")}
          onConfirm={handleDelete}
          disabled={!canDelete || getActiveKeys(selectedTeam) > 0 || members.length > 0 || getVirtualUsers(selectedTeam) > 0}
          warning={<Alert variant="warning"><AlertTriangle aria-hidden="true" /><AlertTitle>{t("delete.guardTitle")}</AlertTitle><AlertDescription>{t("delete.guardDescription")}</AlertDescription></Alert>}
        />
      ) : null}

      <CommercialPlanUpgradeDialog
        open={isUpgradeDialogOpen}
        onOpenChange={setIsUpgradeDialogOpen}
        title={t("upgradeDialog.title")}
        description={t("upgradeDialog.description")}
        cancelLabel={t("upgradeDialog.cancel")}
        actionLabel={t("upgradeDialog.action")}
      />
    </div>
  );
}

function TeamForm({ t, form, setForm, orgUnits, onSubmit, isSaving, submitLabel }: { t: ReturnType<typeof useTranslations>; form: TeamForm; setForm: (value: TeamForm) => void; orgUnits: OrgUnit[]; onSubmit: (event: FormEvent<HTMLFormElement>) => void; isSaving: boolean; submitLabel: string; }) {
  const update = <K extends keyof TeamForm>(key: K, value: TeamForm[K]) => setForm({ ...form, [key]: value });
  return (
    <form className="mt-6 flex flex-col gap-4" onSubmit={onSubmit}>
      <FieldGroup>
        <Field><FieldLabel htmlFor="team-display">{t("form.displayName")}</FieldLabel><Input id="team-display" value={form.displayName} onChange={(e) => update("displayName", e.target.value)} /></Field>
        <IdentifierField id="team-alias" label={t("form.alias")} value={form.alias} onChange={(e) => update("alias", e.target.value)} required />
        <AppSelectField label={t("form.orgUnit")} value={form.orgUnitID || "root"} onValueChange={(value) => update("orgUnitID", value === "root" ? "" : value)} triggerClassName="h-10" options={[{ value: "root", label: t("table.tenantRoot") }, ...orgUnits.map((org) => ({ value: org.id, label: org.display_name || org.alias }))]} />
        <AppSelectField label={t("form.workspaceType")} value={form.workspaceType} onValueChange={(value) => update("workspaceType", value)} triggerClassName="h-10" options={WORKSPACE_TYPES.map((type) => ({ value: type, label: t(`workspace.${type}`) }))} />
        <Field><FieldLabel htmlFor="team-description">{t("form.description")}</FieldLabel><Input id="team-description" value={form.description} onChange={(e) => update("description", e.target.value)} /></Field>
      </FieldGroup>
      <Button type="submit" className="w-full" disabled={isSaving}>{isSaving ? t("actions.saving") : submitLabel}</Button>
    </form>
  );
}

function buildTeamPayload(form: TeamForm, metadata?: Record<string, unknown>) {
  return {
    alias: form.alias.trim(),
    display_name: form.displayName.trim(),
    org_unit_id: form.orgUnitID.trim() || null,
    workspace_type: form.workspaceType,
    metadata: {
      ...withoutLegacyTeamMetadata(metadata),
      description: form.description.trim(),
    },
  };
}

function withoutLegacyTeamMetadata(metadata?: Record<string, unknown>) {
  return Object.fromEntries(
    Object.entries(metadata ?? {}).filter(([key]) => key !== "allowed_models" && key !== "cost_center"),
  );
}

function validateForm(form: TeamForm, teams: Team[], t: ReturnType<typeof useTranslations>, currentID?: string) {
  const alias = form.alias.trim();
  if (!alias) return t("messages.aliasRequired");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{1,62}$/.test(alias)) return t("messages.aliasPattern");
  if (teams.some((team) => team.id !== currentID && team.alias.toLowerCase() === alias.toLowerCase())) return t("messages.aliasUnique");
  return null;
}

function formFromTeam(team: Team): TeamForm {
  return {
    alias: team.alias,
    displayName: getTeamName(team),
    orgUnitID: team.org_unit_id ?? "",
    workspaceType: getWorkspaceType(team),
    description: toText(team.metadata?.description),
  };
}

function formatMemberCandidateLabel(candidate: AccessPrincipalCandidate) {
  const secondary = candidate.secondary?.trim() ?? "";
  const display = candidate.display_name.trim();
  if (secondary && !looksLikeUserID(secondary)) return secondary;
  if (!looksLikeUserID(display)) return display;
  if (secondary) return secondary;
  return display;
}

function resolveMemberUserID(input: string, options: MemberCandidateOption[]) {
  const normalizedInput = input.trim();
  if (!normalizedInput) return "";
  const byID = options.find((option) => option.userID === normalizedInput);
  if (byID) return byID.userID;
  const byLabel = options.find((option) => option.label === normalizedInput);
  if (byLabel) return byLabel.userID;
  if (looksLikeUserID(normalizedInput)) return normalizedInput;
  return "";
}

function looksLikeUserID(value: string) {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value.trim());
}

function getMemberDisplayLabel(member: TeamMember, memberLabelByUserID: Map<string, string>, unknownUserLabel: string) {
  const userLabel = member.user_email?.trim() || member.user_mobile?.trim() || member.user_display_name?.trim() || memberLabelByUserID.get(member.user_id);
  if (userLabel) return userLabel;
  if (looksLikeUserID(member.user_id)) return unknownUserLabel;
  return member.user_id;
}

function buildSummary(team: Team, members: TeamMember[]) {
  return {
    members: members.length || getMemberCount(team),
    admins: members.filter((item) => item.role === "team_admin").length || getAdminCount(team),
    activeKeys: getActiveKeys(team),
  };
}

function getTeamName(team: Team) { return team.display_name || toText(team.metadata?.display_name) || team.alias; }
function getWorkspaceType(team: Team) {
  const value = (team.workspace_type || toText(team.metadata?.workspace_type) || "production").trim().toLowerCase();
  return WORKSPACE_TYPE_SET.has(value) ? value : "general";
}
function getMemberCount(team: Team) {
  const metadata = team.metadata ?? {};
  return firstPositiveCount(
    metadata.member_count,
    metadata.members_count,
    metadata.memberCount,
    metadata.membersCount,
    metadata.team_member_count,
    metadata.team_members_count,
    metadata.members,
    metadata.team_members,
    metadata.users,
    (team as Team & { member_count?: unknown }).member_count,
    (team as Team & { members_count?: unknown }).members_count,
    (team as Team & { memberCount?: unknown }).memberCount,
    (team as Team & { membersCount?: unknown }).membersCount
  );
}
function getAdminCount(team: Team) {
  const metadata = team.metadata ?? {};
  return firstPositiveCount(
    metadata.admin_count,
    metadata.admins_count,
    metadata.adminCount,
    metadata.adminsCount,
    metadata.team_admin_count,
    metadata.team_admins_count,
    metadata.admins,
    metadata.team_admins,
    (team as Team & { admin_count?: unknown }).admin_count,
    (team as Team & { admins_count?: unknown }).admins_count,
    (team as Team & { adminCount?: unknown }).adminCount,
    (team as Team & { adminsCount?: unknown }).adminsCount
  );
}
function getActiveKeys(team: Team) { return toNumber(team.metadata?.active_keys_count ?? team.metadata?.active_key_count); }
function getVirtualUsers(team: Team) { return toNumber(team.metadata?.virtual_users_count); }
function getStatus(team: Team): "active" | "blocked" {
  if (Boolean(team.metadata?.blocked)) return "blocked";
  return "active";
}
function getOrgUnitName(team: Team, orgUnits: OrgUnit[], t: ReturnType<typeof useTranslations>) {
  if (!team.org_unit_id) return t("table.tenantRoot");
  const org = orgUnits.find((item) => item.id === team.org_unit_id);
  return org ? org.display_name || org.alias : t("table.orgUnknown");
}
function summarizePermissions(members: TeamMember[], t: ReturnType<typeof useTranslations>) {
  const permissions = members.flatMap((m) => m.team_member_permissions ?? []);
  const unique = new Set(permissions);
  return unique.size === 0 ? t("common.notSet") : `${unique.size}`;
}
function toNumber(value: unknown) { if (typeof value === "number" && Number.isFinite(value)) return value; if (typeof value === "string") { const parsed = Number(value.replace(/,/g, "")); return Number.isFinite(parsed) ? parsed : 0; } return 0; }
function toCount(value: unknown) {
  if (Array.isArray(value)) return value.length;
  return toNumber(value);
}
function firstPositiveCount(...values: unknown[]) {
  for (const value of values) {
    const count = toCount(value);
    if (count > 0) return count;
  }
  return 0;
}
function toText(value: unknown) { if (typeof value === "string") return value; if (typeof value === "number") return String(value); return ""; }
function permissionTranslationKey(permission: (typeof PERMISSIONS)[number]) {
  return permission.slice("/key/".length);
}

function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}


async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
  return payload as T;
}
function parsePayload(text: string): unknown { if (!text.trim()) return {}; try { return JSON.parse(text) as unknown; } catch { return { message: text }; } }
function extractMessage(payload: unknown): string | null { if (typeof payload === "string") return payload; if (!payload || typeof payload !== "object") return null; const candidate = payload as { error?: { message?: string } | string; message?: string }; if (typeof candidate.error === "string") return candidate.error; return candidate.error?.message ?? candidate.message ?? null; }
function normalizeError(error: unknown, fallback: string) { if (error instanceof Error) return error.message; return fallback; }
