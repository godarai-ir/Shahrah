"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import {
  AlertTriangle,
  ChevronLeft,
  KeyRound,
  MoreVertical,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
  Trash2,
  UserCircle2,
  Users,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";

type Team = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  org_unit_id?: string | null;
  workspace_type?: string;
  budget_id?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};
type TeamMember = {
  id?: string;
  user_id: string;
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

type TeamForm = {
  alias: string;
  displayName: string;
  orgUnitID: string;
  workspaceType: string;
  allowedModels: string;
  budgetID: string;
  costCenter: string;
  description: string;
  maxBudget: string;
  softBudget: string;
  budgetDuration: string;
  tpmLimit: string;
  rpmLimit: string;
};

const PERMISSIONS = ["/key/info", "/key/list", "/key/generate", "/key/update", "/key/delete", "/key/regenerate", "/key/block", "/key/unblock", "/key/health"];
const DEFAULT_PERMISSIONS = ["/key/info", "/key/health"];
const WORKSPACE_TYPES = ["production", "staging", "development", "research", "internal"];
const PAGE_SIZE = 25;

const emptyForm: TeamForm = {
  alias: "",
  displayName: "",
  orgUnitID: "",
  workspaceType: "production",
  allowedModels: "",
  budgetID: "",
  costCenter: "",
  description: "",
  maxBudget: "",
  softBudget: "",
  budgetDuration: "30d",
  tpmLimit: "",
  rpmLimit: "",
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
  const tabT = useTranslations("accessManagement");
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
  const [deleteConfirmation, setDeleteConfirmation] = useState("");

  const canManage = ["tenant_admin", "proxy_admin", "org_admin", "team_admin"].includes(currentUserRole);
  const canDelete = ["tenant_admin", "proxy_admin", "org_admin"].includes(currentUserRole);

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
      if (sortBy === "spend") return getSpend(b) - getSpend(a);
      if (sortBy === "members") return getMemberCount(b) - getMemberCount(a);
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });
  }, [teams, orgFilter, workspaceFilter, statusFilter, sortBy]);

  const memberCandidateOptions = useMemo<MemberCandidateOption[]>(() => {
    return users.map((user) => ({
      userID: user.principal_id,
      label: formatMemberCandidateLabel(user),
    }));
  }, [users]);
  const memberLabelByUserID = useMemo(() => {
    return new Map(memberCandidateOptions.map((item) => [item.userID, item.label]));
  }, [memberCandidateOptions]);

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
      setSelectedTeam(team);
      setMembers(memberPayload.data ?? []);
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
    if (!budgetValid(form, selectedTeam)) return setBanner({ kind: "warning", message: t("messages.budgetCannotDrop") });
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

  async function handleDelete(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedTeam) return;
    if (deleteConfirmation.trim() !== selectedTeam.alias) return setBanner({ kind: "error", message: t("messages.deleteConfirmAlias") });
    if (getActiveKeys(selectedTeam) > 0 || members.length > 0 || getVirtualUsers(selectedTeam) > 0) {
      return setBanner({ kind: "warning", message: t("messages.deleteBlocked") });
    }
    setIsSaving(true);
    try {
      await requestJSON(`/api/teams/${encodeURIComponent(selectedTeam.id)}`, { method: "DELETE" });
      setDeleteOpen(false);
      setDeleteConfirmation("");
      setBanner({ kind: "success", message: t("messages.deleteSuccess") });
      await loadTeams(0, query.trim());
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.deleteError")) });
    } finally {
      setIsSaving(false);
    }
  }

  const summary = selectedTeam ? buildSummary(selectedTeam, members) : null;
  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + teams.length, total);

  return (
    <div className="space-y-5">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <span>{t("breadcrumb.tenant")}</span>
            <ChevronLeft className="h-4 w-4" />
            <span>{tabT("title")}</span>
            <ChevronLeft className="h-4 w-4" />
            <span className="text-foreground">{t("title")}</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          <nav className="border-b">
            <ul className="flex flex-wrap items-center gap-5 pb-2">
              <li><Link className="inline-flex items-center gap-2 border-b-2 border-transparent px-1 py-2 text-muted-foreground hover:text-foreground" href="/users"><UserCircle2 className="h-4 w-4" /><span className="text-sm font-semibold">{tabT("tabs.users")}</span></Link></li>
              <li><Link className="inline-flex items-center gap-2 border-b-2 border-indigo-500 px-1 py-2 text-foreground" href="/teams" aria-current="page"><Users className="h-4 w-4" /><span className="text-sm font-semibold">{tabT("tabs.teams")}</span></Link></li>
              <li><span className="inline-flex items-center gap-2 px-1 py-2 text-sm text-muted-foreground"><ShieldCheck className="h-4 w-4" />{tabT("tabs.roles")}</span></li>
              <li><span className="inline-flex items-center gap-2 px-1 py-2 text-sm text-muted-foreground"><KeyRound className="h-4 w-4" />{tabT("tabs.personalAccessTokens")}</span></li>
            </ul>
          </nav>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="outline" className="gap-2" onClick={() => void loadTeams(offset, query.trim())} disabled={isLoading}>
            <RefreshCw className={cn("h-4 w-4", isLoading && "animate-spin")} />{t("actions.refresh")}
          </Button>
          <Sheet open={createOpen} onOpenChange={setCreateOpen}>
            <Button className="gap-2" onClick={() => setCreateOpen(true)} disabled={!canManage}><Plus className="h-4 w-4" />{t("actions.new")}</Button>
            <SheetContent side="right" className="w-[94vw] max-w-xl overflow-y-auto">
              <SheetHeader><SheetTitle>{t("create.title")}</SheetTitle><SheetDescription>{t("create.description")}</SheetDescription></SheetHeader>
              <TeamForm t={t} form={form} setForm={setForm} orgUnits={orgUnits} onSubmit={handleCreate} isSaving={isSaving} submitLabel={t("create.submit")} />
            </SheetContent>
          </Sheet>
        </div>
      </header>

      {banner ? <Alert variant={banner.kind === "error" ? "destructive" : banner.kind}><AlertTitle>{t(`alerts.${banner.kind}`)}</AlertTitle><AlertDescription>{banner.message}</AlertDescription></Alert> : null}

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(360px,0.8fr)]">
        <div className="space-y-4 rounded-xl border bg-card p-4">
          <div className="flex flex-wrap gap-2">
            <div className="relative"><Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" /><Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t("filters.search")} className="w-full pl-9 sm:w-64" /></div>
            <select value={orgFilter} onChange={(e) => setOrgFilter(e.target.value)} className="h-10 rounded-md border border-input bg-background px-3 text-sm"><option value="all">{t("filters.orgAll")}</option><option value="root">{t("table.tenantRoot")}</option>{orgUnits.map((org) => <option key={org.id} value={org.id}>{org.display_name || org.alias}</option>)}</select>
            <select value={workspaceFilter} onChange={(e) => setWorkspaceFilter(e.target.value)} className="h-10 rounded-md border border-input bg-background px-3 text-sm"><option value="all">{t("filters.workspaceAll")}</option>{WORKSPACE_TYPES.map((type) => <option key={type} value={type}>{t(`workspace.${type}`)}</option>)}</select>
            <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="h-10 rounded-md border border-input bg-background px-3 text-sm"><option value="all">{t("filters.statusAll")}</option><option value="active">{t("status.active")}</option><option value="blocked">{t("status.blocked")}</option><option value="budget_warning">{t("status.budgetWarning")}</option><option value="budget_exceeded">{t("status.budgetExceeded")}</option></select>
            <select value={sortBy} onChange={(e) => setSortBy(e.target.value)} className="h-10 rounded-md border border-input bg-background px-3 text-sm"><option value="recent">{t("sort.recent")}</option><option value="name">{t("sort.name")}</option><option value="spend">{t("sort.spend")}</option><option value="members">{t("sort.members")}</option></select>
          </div>

          <div className="overflow-x-auto">
            <Table>
              <TableHeader><TableRow><TableHead>{t("table.team")}</TableHead><TableHead>{t("table.orgUnit")}</TableHead><TableHead>{t("table.workspaceType")}</TableHead><TableHead>{t("table.membersAdmins")}</TableHead><TableHead>{t("table.budgetSpend")}</TableHead><TableHead>{t("table.models")}</TableHead><TableHead>{t("table.limits")}</TableHead><TableHead>{t("table.status")}</TableHead><TableHead>{t("table.updated")}</TableHead><TableHead className="w-10 text-right"> </TableHead></TableRow></TableHeader>
              <TableBody>
                {filtered.length === 0 ? <TableRow><TableCell colSpan={10} className="py-10 text-center"><div className="mx-auto max-w-md space-y-2"><Users className="mx-auto h-9 w-9 text-muted-foreground" /><p className="font-medium">{t("empty.title")}</p><p className="text-sm text-muted-foreground">{t("empty.description")}</p></div></TableCell></TableRow> : filtered.map((team) => (
                  <TableRow key={team.id} className={cn("cursor-pointer", selectedTeamID === team.id && "bg-muted/50")} onClick={() => setSelectedTeamID(team.id)}>
                    <TableCell><div className="min-w-[180px]"><p className="truncate font-medium">{getTeamName(team)}</p><p className="truncate text-xs text-muted-foreground">{team.alias}</p></div></TableCell>
                    <TableCell>{getOrgUnitName(team, orgUnits, t)}</TableCell>
                    <TableCell><Badge variant="outline">{t(`workspace.${getWorkspaceType(team)}`)}</Badge></TableCell>
                    <TableCell>{`${getMemberCount(team)}/${getAdminCount(team)}`}</TableCell>
                    <TableCell>{`${formatMoney(getSpend(team))} / ${formatMoney(getMaxBudget(team))}`}</TableCell>
                    <TableCell>{getModelSummary(team, t)}</TableCell>
                    <TableCell>{`${formatLimit(getTpm(team))} / ${formatLimit(getRpm(team))}`}</TableCell>
                    <TableCell><StatusBadge status={getStatus(team)} t={t} /></TableCell>
                    <TableCell>{formatDate(team.updated_at)}</TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon"><MoreVertical className="h-4 w-4" /></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => { setForm(formFromTeam(team)); setEditOpen(true); }} disabled={!canManage}>{t("actions.edit")}</DropdownMenuItem>
                        <DropdownMenuItem onClick={() => { setSelectedTeamID(team.id); setMembersOpen(true); void loadUsers(); }} disabled={!canManage}>{t("actions.members")}</DropdownMenuItem>
                        <DropdownMenuItem onClick={() => { setSelectedTeamID(team.id); setDeleteOpen(true); }} disabled={!canDelete} className="text-destructive"><Trash2 className="mr-2 h-4 w-4" />{t("actions.delete")}</DropdownMenuItem>
                      </DropdownMenuContent></DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          <div className="flex items-center justify-between text-sm text-muted-foreground"><p>{t("pagination.showing", { from, to, total })}</p><div className="flex gap-2"><Button variant="outline" size="sm" onClick={() => void loadTeams(Math.max(0, offset - PAGE_SIZE), query.trim())} disabled={offset <= 0 || isLoading}>{t("pagination.prev")}</Button><Button variant="outline" size="sm" onClick={() => void loadTeams(offset + PAGE_SIZE, query.trim())} disabled={offset + teams.length >= total || isLoading}>{t("pagination.next")}</Button></div></div>
        </div>

        <div className="space-y-4">
          {!selectedTeam || !summary ? <Card><CardContent className="py-10 text-center text-sm text-muted-foreground">{t("detail.empty")}</CardContent></Card> : <>
            <div className="grid grid-cols-2 gap-3">
              <Kpi label={t("kpi.spend")} value={formatMoney(summary.spend)} />
              <Kpi label={t("kpi.remaining")} value={formatMoney(summary.remaining)} />
              <Kpi label={t("kpi.membersAdmins")} value={`${summary.members}/${summary.admins}`} />
              <Kpi label={t("kpi.activeKeys")} value={String(summary.activeKeys)} />
            </div>
            <Card><CardHeader><CardTitle>{t("detail.guardrails")}</CardTitle></CardHeader><CardContent className="space-y-2 text-sm"><p>{t("detail.budget", { spend: formatMoney(summary.spend), max: formatMoney(summary.maxBudget) })}</p><p>{t("detail.softBudget", { value: formatMoney(summary.softBudget) })}</p><p>{t("detail.duration", { value: summary.duration || t("common.notSet") })}</p><p>{t("detail.limits", { tpm: formatLimit(summary.tpm), rpm: formatLimit(summary.rpm) })}</p></CardContent></Card>
            <Card><CardHeader><CardTitle>{t("detail.access")}</CardTitle></CardHeader><CardContent className="space-y-2 text-sm"><p>{t("detail.models", { value: getModelSummary(selectedTeam, t) })}</p><p>{t("detail.orgInheritance", { value: getOrgUnitName(selectedTeam, orgUnits, t) })}</p><p>{t("detail.permissions", { value: summarizePermissions(members, t) })}</p></CardContent></Card>
            <Card><CardHeader><CardTitle>{t("detail.members")}</CardTitle></CardHeader><CardContent className="space-y-2 text-sm">{detailLoading ? <p>{t("detail.loading")}</p> : members.length === 0 ? <p>{t("detail.noMembers")}</p> : members.map((member) => <div key={member.user_id} className="flex items-center justify-between gap-2 rounded border p-2"><span className="truncate">{getMemberDisplayLabel(member.user_id, memberLabelByUserID, t("members.unknownUser"))}</span><Badge variant={member.role === "team_admin" ? "default" : "secondary"}>{member.role}</Badge></div>)}</CardContent></Card>
          </>}
        </div>
      </section>

      <Sheet open={editOpen} onOpenChange={setEditOpen}><SheetContent side="right" className="w-[94vw] max-w-xl overflow-y-auto"><SheetHeader><SheetTitle>{t("edit.title")}</SheetTitle><SheetDescription>{t("edit.description")}</SheetDescription></SheetHeader><TeamForm t={t} form={form} setForm={setForm} orgUnits={orgUnits} onSubmit={handleEdit} isSaving={isSaving} submitLabel={t("edit.submit")} /></SheetContent></Sheet>

      <Sheet open={membersOpen} onOpenChange={setMembersOpen}><SheetContent side="right" className="w-[94vw] max-w-xl overflow-y-auto"><SheetHeader><SheetTitle>{t("members.title")}</SheetTitle><SheetDescription>{t("members.description")}</SheetDescription></SheetHeader>
        <form onSubmit={addMember} className="mt-6 space-y-3">
          <div className="grid gap-2"><Label htmlFor="team-member-user">{t("members.user")}</Label><Input id="team-member-user" value={memberUserInput} onChange={(e) => setMemberUserInput(e.target.value)} list="member-candidates" placeholder={t("members.userPlaceholder")} /><datalist id="member-candidates">{memberCandidateOptions.map((user) => <option key={user.userID} value={user.label} />)}</datalist></div>
          <div className="grid gap-2"><Label htmlFor="team-member-role">{t("members.role")}</Label><select id="team-member-role" className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={memberRole} onChange={(e) => setMemberRole(e.target.value)}><option value="team_member">{t("members.teamMember")}</option><option value="team_admin">{t("members.teamAdmin")}</option></select></div>
          {memberRole === "team_member" ? <div className="grid gap-2"><Label>{t("members.permissions")}</Label><div className="grid gap-2">{PERMISSIONS.map((perm) => <label key={perm} className="flex items-center gap-2 text-sm"><input type="checkbox" checked={memberPerms.includes(perm)} onChange={(e) => setMemberPerms((prev) => e.target.checked ? [...prev, perm] : prev.filter((item) => item !== perm))} /><span>{perm}</span></label>)}</div></div> : null}
          <Button type="submit" className="w-full" disabled={!canManage || isSaving}>{t("members.add")}</Button>
        </form>
        <div className="mt-6 space-y-2">{members.map((member) => <div key={member.user_id} className="rounded border p-3"><div className="flex items-center justify-between"><span className="truncate text-sm">{getMemberDisplayLabel(member.user_id, memberLabelByUserID, t("members.unknownUser"))}</span><Badge variant={member.role === "team_admin" ? "default" : "secondary"}>{member.role}</Badge></div><div className="mt-2 flex flex-wrap gap-2"><Button size="sm" variant="outline" onClick={() => void updateMember(member, member.role === "team_admin" ? "team_member" : "team_admin", member.team_member_permissions ?? DEFAULT_PERMISSIONS)} disabled={!canManage || isSaving}>{member.role === "team_admin" ? t("members.makeTeamMember") : t("members.makeTeamAdmin")}</Button><Button size="sm" variant="outline" className="text-destructive" onClick={() => void removeMember(member.user_id)} disabled={!canManage || isSaving}>{t("members.remove")}</Button></div></div>)}</div>
      </SheetContent></Sheet>

      <Sheet open={deleteOpen} onOpenChange={setDeleteOpen}><SheetContent side="right" className="w-[94vw] max-w-md"><SheetHeader><SheetTitle>{t("delete.title")}</SheetTitle><SheetDescription>{t("delete.description")}</SheetDescription></SheetHeader>
        {selectedTeam ? <form className="mt-6 space-y-4" onSubmit={handleDelete}><Alert variant="warning"><AlertTriangle className="h-4 w-4" /><AlertTitle>{t("delete.guardTitle")}</AlertTitle><AlertDescription>{t("delete.guardDescription")}</AlertDescription></Alert><div className="grid gap-2"><Label htmlFor="delete-alias">{t("delete.confirm", { alias: selectedTeam.alias })}</Label><Input id="delete-alias" value={deleteConfirmation} onChange={(e) => setDeleteConfirmation(e.target.value)} placeholder={selectedTeam.alias} /></div><Button type="submit" variant="destructive" disabled={!canDelete || isSaving} className="w-full">{t("delete.submit")}</Button></form> : null}
      </SheetContent></Sheet>
    </div>
  );
}

function TeamForm({ t, form, setForm, orgUnits, onSubmit, isSaving, submitLabel }: { t: ReturnType<typeof useTranslations>; form: TeamForm; setForm: (value: TeamForm) => void; orgUnits: OrgUnit[]; onSubmit: (event: FormEvent<HTMLFormElement>) => void; isSaving: boolean; submitLabel: string; }) {
  const max = toNumber(form.maxBudget);
  const soft = toNumber(form.softBudget);
  const softError = max > 0 && soft > max;
  const update = <K extends keyof TeamForm>(key: K, value: TeamForm[K]) => setForm({ ...form, [key]: value });
  return (
    <form className="mt-6 space-y-4" onSubmit={onSubmit}>
      <div className="grid gap-2"><Label htmlFor="team-alias">{t("form.alias")}</Label><Input id="team-alias" value={form.alias} onChange={(e) => update("alias", e.target.value)} required /></div>
      <div className="grid gap-2"><Label htmlFor="team-display">{t("form.displayName")}</Label><Input id="team-display" value={form.displayName} onChange={(e) => update("displayName", e.target.value)} /></div>
      <div className="grid gap-2"><Label htmlFor="team-org">{t("form.orgUnit")}</Label><select id="team-org" className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={form.orgUnitID} onChange={(e) => update("orgUnitID", e.target.value)}><option value="">{t("table.tenantRoot")}</option>{orgUnits.map((org) => <option key={org.id} value={org.id}>{org.display_name || org.alias}</option>)}</select></div>
      <div className="grid gap-2"><Label htmlFor="team-workspace">{t("form.workspaceType")}</Label><select id="team-workspace" className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={form.workspaceType} onChange={(e) => update("workspaceType", e.target.value)}>{WORKSPACE_TYPES.map((type) => <option key={type} value={type}>{t(`workspace.${type}`)}</option>)}</select></div>
      <div className="grid gap-2"><Label htmlFor="team-models">{t("form.allowedModels")}</Label><Input id="team-models" value={form.allowedModels} onChange={(e) => update("allowedModels", e.target.value)} placeholder={t("form.allowedModelsPlaceholder")} /></div>
      <div className="grid gap-2"><Label htmlFor="team-budget-id">{t("form.budgetId")}</Label><Input id="team-budget-id" value={form.budgetID} onChange={(e) => update("budgetID", e.target.value)} /></div>
      <div className="grid gap-2"><Label htmlFor="team-cost">{t("form.costCenter")}</Label><Input id="team-cost" value={form.costCenter} onChange={(e) => update("costCenter", e.target.value)} /></div>
      <div className="grid gap-2"><Label htmlFor="team-description">{t("form.description")}</Label><Input id="team-description" value={form.description} onChange={(e) => update("description", e.target.value)} /></div>
      <div className="grid gap-2 sm:grid-cols-2"><div className="grid gap-2"><Label htmlFor="team-max">{t("form.maxBudget")}</Label><Input id="team-max" type="number" min="0" value={form.maxBudget} onChange={(e) => update("maxBudget", e.target.value)} /></div><div className="grid gap-2"><Label htmlFor="team-soft">{t("form.softBudget")}</Label><Input id="team-soft" type="number" min="0" value={form.softBudget} onChange={(e) => update("softBudget", e.target.value)} /></div></div>
      <div className="grid gap-2"><Label htmlFor="team-duration">{t("form.budgetDuration")}</Label><select id="team-duration" className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={form.budgetDuration} onChange={(e) => update("budgetDuration", e.target.value)}><option value="30d">30d</option><option value="7d">7d</option><option value="1d">1d</option></select></div>
      <div className="grid gap-2 sm:grid-cols-2"><div className="grid gap-2"><Label htmlFor="team-tpm">{t("form.tpmLimit")}</Label><Input id="team-tpm" type="number" min="0" value={form.tpmLimit} onChange={(e) => update("tpmLimit", e.target.value)} /></div><div className="grid gap-2"><Label htmlFor="team-rpm">{t("form.rpmLimit")}</Label><Input id="team-rpm" type="number" min="0" value={form.rpmLimit} onChange={(e) => update("rpmLimit", e.target.value)} /></div></div>
      {softError ? <Alert variant="warning"><AlertTitle>{t("alerts.warning")}</AlertTitle><AlertDescription>{t("messages.softBudgetConflict")}</AlertDescription></Alert> : null}
      <Button type="submit" className="w-full" disabled={isSaving || softError}>{isSaving ? t("actions.saving") : submitLabel}</Button>
    </form>
  );
}

function buildTeamPayload(form: TeamForm, metadata?: Record<string, unknown>) {
  return {
    alias: form.alias.trim(),
    display_name: form.displayName.trim(),
    org_unit_id: form.orgUnitID.trim() || null,
    workspace_type: form.workspaceType,
    allowed_models: modelsFromInput(form.allowedModels),
    budget_id: form.budgetID.trim() || undefined,
    metadata: {
      ...(metadata ?? {}),
      description: form.description.trim(),
      cost_center: form.costCenter.trim(),
      max_budget: toNumber(form.maxBudget),
      soft_budget: toNumber(form.softBudget),
      budget_duration: form.budgetDuration,
      tpm_limit: toNumber(form.tpmLimit),
      rpm_limit: toNumber(form.rpmLimit),
      allowed_models: modelsFromInput(form.allowedModels),
    },
  };
}

function validateForm(form: TeamForm, teams: Team[], t: ReturnType<typeof useTranslations>, currentID?: string) {
  const alias = form.alias.trim();
  if (!alias) return t("messages.aliasRequired");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{1,62}$/.test(alias)) return t("messages.aliasPattern");
  if (teams.some((team) => team.id !== currentID && team.alias.toLowerCase() === alias.toLowerCase())) return t("messages.aliasUnique");
  if (toNumber(form.maxBudget) > 0 && toNumber(form.softBudget) > toNumber(form.maxBudget)) return t("messages.softBudgetConflict");
  return null;
}

function budgetValid(form: TeamForm, team: Team) {
  const next = toNumber(form.maxBudget);
  const spend = getSpend(team);
  return next <= 0 || next >= spend;
}

function formFromTeam(team: Team): TeamForm {
  return {
    alias: team.alias,
    displayName: getTeamName(team),
    orgUnitID: team.org_unit_id ?? "",
    workspaceType: getWorkspaceType(team),
    allowedModels: getModels(team).join(","),
    budgetID: team.budget_id ?? "",
    costCenter: toText(team.metadata?.cost_center),
    description: toText(team.metadata?.description),
    maxBudget: numberToInput(getMaxBudget(team)),
    softBudget: numberToInput(getSoftBudget(team)),
    budgetDuration: toText(team.metadata?.budget_duration) || "30d",
    tpmLimit: numberToInput(getTpm(team)),
    rpmLimit: numberToInput(getRpm(team)),
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
  const byLabel = options.find((option) => option.label === normalizedInput);
  if (byLabel) return byLabel.userID;
  if (looksLikeUserID(normalizedInput)) return normalizedInput;
  return "";
}

function looksLikeUserID(value: string) {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value.trim());
}

function getMemberDisplayLabel(userID: string, memberLabelByUserID: Map<string, string>, unknownUserLabel: string) {
  const userLabel = memberLabelByUserID.get(userID);
  if (userLabel) return userLabel;
  if (looksLikeUserID(userID)) return unknownUserLabel;
  return userID;
}

function buildSummary(team: Team, members: TeamMember[]) {
  const maxBudget = getMaxBudget(team);
  const spend = getSpend(team);
  return {
    maxBudget,
    softBudget: getSoftBudget(team),
    spend,
    remaining: maxBudget > 0 ? Math.max(0, maxBudget - spend) : 0,
    duration: toText(team.metadata?.budget_duration),
    members: members.length || getMemberCount(team),
    admins: members.filter((item) => item.role === "team_admin").length || getAdminCount(team),
    activeKeys: getActiveKeys(team),
    tpm: getTpm(team),
    rpm: getRpm(team),
  };
}

function getTeamName(team: Team) { return team.display_name || toText(team.metadata?.display_name) || team.alias; }
function getWorkspaceType(team: Team) { return team.workspace_type || toText(team.metadata?.workspace_type) || "production"; }
function getMemberCount(team: Team) { return toNumber(team.metadata?.member_count ?? team.metadata?.members_count); }
function getAdminCount(team: Team) { return toNumber(team.metadata?.admin_count ?? team.metadata?.admins_count); }
function getSpend(team: Team) { return toNumber(team.metadata?.spend ?? team.metadata?.current_spend ?? team.metadata?.total_spend); }
function getMaxBudget(team: Team) { return toNumber(team.metadata?.max_budget); }
function getSoftBudget(team: Team) { return toNumber(team.metadata?.soft_budget); }
function getTpm(team: Team) { return toNumber(team.metadata?.tpm_limit); }
function getRpm(team: Team) { return toNumber(team.metadata?.rpm_limit); }
function getActiveKeys(team: Team) { return toNumber(team.metadata?.active_keys_count ?? team.metadata?.active_key_count); }
function getVirtualUsers(team: Team) { return toNumber(team.metadata?.virtual_users_count); }
function getModels(team: Team) {
  const fromTop = Array.isArray(team.metadata?.allowed_models) ? team.metadata?.allowed_models : [];
  const normalized = Array.isArray(fromTop) ? fromTop.map((item) => String(item).trim()).filter(Boolean) : [];
  return normalized;
}
function getModelSummary(team: Team, t: ReturnType<typeof useTranslations>) {
  const models = getModels(team);
  if (models.length === 0) return t("common.notSet");
  if (models.length === 1) return models[0];
  return t("table.modelsSummary", { count: models.length, first: models[0] });
}
function getStatus(team: Team): "active" | "blocked" | "budget_warning" | "budget_exceeded" {
  if (Boolean(team.metadata?.blocked)) return "blocked";
  const max = getMaxBudget(team);
  const spend = getSpend(team);
  const soft = getSoftBudget(team);
  if (max > 0 && spend >= max) return "budget_exceeded";
  if ((soft > 0 && spend >= soft) || (max > 0 && spend / max >= 0.8)) return "budget_warning";
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
function modelsFromInput(value: string) { return value.split(",").map((item) => item.trim()).filter(Boolean); }
function toNumber(value: unknown) { if (typeof value === "number" && Number.isFinite(value)) return value; if (typeof value === "string") { const parsed = Number(value.replace(/,/g, "")); return Number.isFinite(parsed) ? parsed : 0; } return 0; }
function toText(value: unknown) { if (typeof value === "string") return value; if (typeof value === "number") return String(value); return ""; }
function numberToInput(value: number) { return value > 0 ? String(value) : ""; }
function formatMoney(value: number) { if (value <= 0) return "-"; return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 }).format(value); }
function formatLimit(value: number) { if (value <= 0) return "-"; return new Intl.NumberFormat("en-US", { notation: "compact", maximumFractionDigits: 1 }).format(value); }
function formatDate(value: string) { if (!value) return "-"; return new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric" }).format(new Date(value)); }

function StatusBadge({ status, t }: { status: ReturnType<typeof getStatus>; t: ReturnType<typeof useTranslations> }) {
  if (status === "blocked") return <Badge variant="destructive">{t("status.blocked")}</Badge>;
  if (status === "budget_warning") return <Badge className="bg-warning text-warning-foreground">{t("status.budgetWarning")}</Badge>;
  if (status === "budget_exceeded") return <Badge variant="destructive">{t("status.budgetExceeded")}</Badge>;
  return <Badge className="bg-success text-success-foreground">{t("status.active")}</Badge>;
}

function Kpi({ label, value }: { label: string; value: string }) {
  return <Card><CardHeader className="pb-2"><CardTitle className="text-xs text-muted-foreground">{label}</CardTitle></CardHeader><CardContent><p className="truncate text-lg font-semibold">{value}</p></CardContent></Card>;
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
