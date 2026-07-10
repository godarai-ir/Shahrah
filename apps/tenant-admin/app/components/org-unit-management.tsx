"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Building2,
  CalendarClock,
  Edit3,
  ExternalLink,
  Gauge,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
  Trash2,
  Users,
  WalletCards,
  type LucideIcon,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { AppCard, AppSelectField, EmptyState, IdentifierField, PageHeader, RowActionsMenu, StatusBadge } from "@/components/app-ui";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { usePermissions } from "@/app/components/permissions-context";
import { canAnyPermission } from "@/lib/authz";

type OrgUnit = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

type OrgUnitMember = {
  id: string;
  tenant_id: string;
  org_unit_id: string;
  user_id: string;
  role: string;
  created_at: string;
  updated_at: string;
};

type AccessPrincipalCandidate = {
  type: "user" | "team";
  principal_id: string;
  display_name: string;
  secondary?: string;
};

type OrgUnitListResponse = {
  data?: OrgUnit[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

type MemberListResponse = {
  data?: OrgUnitMember[];
};

type UserListResponse = {
  data?: AccessPrincipalCandidate[];
};

type TeamListResponse = {
  data?: TeamSummary[];
  total?: number;
};

type TeamSummary = {
  id: string;
  alias: string;
  display_name?: string;
  workspace_type?: string;
  org_unit_id?: string | null;
  metadata?: Record<string, unknown>;
};

type Banner = {
  kind: "success" | "error" | "warning";
  message: string;
};

type FormState = {
  alias: string;
  displayName: string;
  description: string;
};

type OrgUnitManagementProps = {
  initialOrgUnits: OrgUnit[];
  initialTotal: number;
  initialError: string | null;
  currentUserRole: string;
  docsHref: string;
};

const PAGE_SIZE = 25;

const emptyForm: FormState = {
  alias: "",
  displayName: "",
  description: "",
};

export default function OrgUnitManagement({
  initialOrgUnits,
  initialTotal,
  initialError,
  currentUserRole,
  docsHref,
}: OrgUnitManagementProps) {
  const t = useTranslations("orgUnits");
  const permissions = usePermissions() ?? [];
  const [orgUnits, setOrgUnits] = useState<OrgUnit[]>(initialOrgUnits);
  const [total, setTotal] = useState(initialTotal);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [sortBy, setSortBy] = useState("updated_at");
  const [offset, setOffset] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [teamCountByOrgUnit, setTeamCountByOrgUnit] = useState<Record<string, number>>({});
  const [teamsByOrgUnit, setTeamsByOrgUnit] = useState<Record<string, TeamSummary[]>>({});
  const [banner, setBanner] = useState<Banner | null>(
    initialError ? { kind: "error", message: initialError } : null,
  );

  const [selectedOrgUnitID, setSelectedOrgUnitID] = useState<string | null>(
    initialOrgUnits[0]?.id ?? null,
  );
  const [selectedOrgUnit, setSelectedOrgUnit] = useState<OrgUnit | null>(
    initialOrgUnits[0] ?? null,
  );
  const [members, setMembers] = useState<OrgUnitMember[]>([]);
  const [users, setUsers] = useState<AccessPrincipalCandidate[]>([]);
  const [isDetailLoading, setIsDetailLoading] = useState(false);

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [memberOpen, setMemberOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [memberUserID, setMemberUserID] = useState("");
  const [memberUserQuery, setMemberUserQuery] = useState("");
  const [isMemberUserPickerOpen, setIsMemberUserPickerOpen] = useState(false);
  const [memberRole, setMemberRole] = useState("org_admin");
  const [deleteConfirmation, setDeleteConfirmation] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const canManage = ["tenant_admin", "proxy_admin", "org_admin"].includes(currentUserRole) || canAnyPermission(permissions, ["org_unit:manage"]);
  const canCreateOrgUnit = canManage || canAnyPermission(permissions, ["org_unit:create"]);
  const canDeleteOrgUnit = canManage || canAnyPermission(permissions, ["org_unit:delete"]);

  useEffect(() => {
    const timer = setTimeout(() => {
      setOffset(0);
      void loadOrgUnits(0, query.trim());
    }, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  useEffect(() => {
    if (!selectedOrgUnitID) {
      return;
    }
    void loadOrgUnitDetails(selectedOrgUnitID);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedOrgUnitID]);

  useEffect(() => {
    void loadTeamCounts();
     
  }, []);

  function resolveTeamCount(orgUnit: OrgUnit) {
    return teamCountByOrgUnit[orgUnit.id] ?? getTeamCount(orgUnit);
  }

  function resolveTeams(orgUnit: OrgUnit) {
    return teamsByOrgUnit[orgUnit.id] ?? [];
  }

  const visibleOrgUnits = useMemo(() => {
    const filtered = orgUnits.filter((orgUnit) => {
      if (statusFilter === "all") return true;
      if (statusFilter === "with_teams") return resolveTeamCount(orgUnit) > 0;
      if (statusFilter === "empty") return resolveTeamCount(orgUnit) === 0;
      return getBudgetStatus(orgUnit).kind === statusFilter;
    });

    return [...filtered].sort((a, b) => {
      if (sortBy === "name") return getOrgName(a).localeCompare(getOrgName(b));
      if (sortBy === "spend") return getSpend(b) - getSpend(a);
      if (sortBy === "teams") return resolveTeamCount(b) - resolveTeamCount(a);
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });
  }, [orgUnits, sortBy, statusFilter, teamCountByOrgUnit]);

  const selectedSummary = selectedOrgUnit
    ? { ...buildSummary(selectedOrgUnit, members), teams: resolveTeamCount(selectedOrgUnit) }
    : null;
  const userNameByID = useMemo(() => {
    const map = new Map<string, string>();
    users.forEach((user) => {
      map.set(user.principal_id, user.display_name || user.secondary || t("members.user"));
    });
    return map;
  }, [t, users]);
  const memberUserOptions = useMemo(() => {
    const queryValue = memberUserQuery.trim().toLowerCase();
    if (!queryValue) return users.slice(0, 50);
    return users
      .filter((user) => {
        const displayName = (user.display_name || "").toLowerCase();
        const secondary = (user.secondary || "").toLowerCase();
        return displayName.includes(queryValue) || secondary.includes(queryValue);
      })
      .slice(0, 50);
  }, [memberUserQuery, users]);
  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + orgUnits.length, total);
  const canPrev = offset > 0;
  const canNext = offset + orgUnits.length < total;

  async function loadOrgUnits(nextOffset: number, q: string) {
    setIsLoading(true);
    try {
      const search = new URLSearchParams();
      search.set("limit", String(PAGE_SIZE));
      search.set("offset", String(nextOffset));
      if (q) search.set("query", q);

      const data = await requestJSON<OrgUnitListResponse>(`/api/org-units?${search.toString()}`);
      const nextItems = data.data ?? [];
      setOrgUnits(nextItems);
      setTotal(data.total ?? nextItems.length);
      setOffset(nextOffset);
      setBanner(null);
      if (!selectedOrgUnitID || !nextItems.some((item) => item.id === selectedOrgUnitID)) {
        setSelectedOrgUnitID(nextItems[0]?.id ?? null);
        setSelectedOrgUnit(nextItems[0] ?? null);
        setMembers([]);
      }
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.loadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadOrgUnitDetails(orgUnitID: string) {
    setIsDetailLoading(true);
    try {
      const [orgUnit, memberPayload] = await Promise.all([
        requestJSON<OrgUnit>(`/api/org-units/${encodeURIComponent(orgUnitID)}`),
        requestJSON<MemberListResponse>(`/api/org-units/${encodeURIComponent(orgUnitID)}/members`),
      ]);
      setSelectedOrgUnit(orgUnit);
      setMembers(memberPayload.data ?? []);
      setOrgUnits((items) => items.map((item) => (item.id === orgUnit.id ? orgUnit : item)));
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.detailLoadError")) });
    } finally {
      setIsDetailLoading(false);
    }
  }

  async function loadUsers() {
    try {
      const payload = await requestJSON<UserListResponse>("/api/access-principals?kind=user&limit=100");
      setUsers(payload.data ?? []);
    } catch {
      // Keep members management usable even when principal-search permissions are not granted.
      setUsers([]);
    }
  }

  async function loadTeamCounts() {
    try {
      const counts: Record<string, number> = {};
      const grouped: Record<string, TeamSummary[]> = {};
      const limit = 200;
      let nextOffset = 0;
      let safety = 0;

      while (safety < 50) {
        const payload = await requestJSON<TeamListResponse>(`/api/teams?limit=${limit}&offset=${nextOffset}`);
        const teams = payload.data ?? [];
        teams.forEach((team) => {
          if (!team.org_unit_id) return;
          counts[team.org_unit_id] = (counts[team.org_unit_id] ?? 0) + 1;
          grouped[team.org_unit_id] = [...(grouped[team.org_unit_id] ?? []), team];
        });

        const totalTeams = payload.total ?? teams.length;
        nextOffset += teams.length;
        safety += 1;

        if (teams.length === 0 || nextOffset >= totalTeams) {
          break;
        }
      }

      setTeamCountByOrgUnit(counts);
      setTeamsByOrgUnit(grouped);
    } catch {
      // Keep metadata-based fallback when team count aggregation fails.
    }
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateForm(form, orgUnits, t);
    if (validation) {
      setBanner({ kind: "error", message: validation });
      return;
    }

    setIsSaving(true);
    try {
      const created = await requestJSON<OrgUnit>("/api/org-units", {
        method: "POST",
        body: JSON.stringify(buildPayload(form)),
      });
      setBanner({ kind: "success", message: t("messages.createSuccess") });
      setCreateOpen(false);
      setForm(emptyForm);
      await loadOrgUnits(0, query.trim());
      setSelectedOrgUnitID(created.id);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.createError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedOrgUnit) return;
    const validation = validateForm(form, orgUnits, t, selectedOrgUnit.id);
    if (validation) {
      setBanner({ kind: "error", message: validation });
      return;
    }

    setIsSaving(true);
    try {
      const updated = await requestJSON<OrgUnit>(`/api/org-units/${encodeURIComponent(selectedOrgUnit.id)}`, {
        method: "PATCH",
        body: JSON.stringify(buildPayload(form, selectedOrgUnit.metadata)),
      });
      setSelectedOrgUnit(updated);
      setOrgUnits((items) => items.map((item) => (item.id === updated.id ? updated : item)));
      setBanner({ kind: "success", message: t("messages.updateSuccess") });
      setEditOpen(false);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.updateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function handleAddMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedOrgUnit) return;
    if (!memberUserID.trim()) {
      setBanner({ kind: "error", message: t("messages.selectUser") });
      return;
    }
    if (members.some((item) => item.user_id === memberUserID.trim())) {
      setBanner({ kind: "error", message: t("messages.duplicateMember") });
      return;
    }

    setIsSaving(true);
    try {
      await requestJSON(`/api/org-units/${encodeURIComponent(selectedOrgUnit.id)}/members`, {
        method: "POST",
        body: JSON.stringify({ user_id: memberUserID.trim(), role: memberRole }),
      });
      await loadOrgUnitDetails(selectedOrgUnit.id);
      setMemberUserID("");
      setMemberUserQuery("");
      setIsMemberUserPickerOpen(false);
      setMemberRole("org_admin");
      setBanner({ kind: "success", message: t("messages.memberAddSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.memberAddError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function updateMemberRole(userID: string, role: string) {
    if (!selectedOrgUnit) return;
    setIsSaving(true);
    try {
      await requestJSON(`/api/org-units/${encodeURIComponent(selectedOrgUnit.id)}/members/${encodeURIComponent(userID)}`, {
        method: "PATCH",
        body: JSON.stringify({ role }),
      });
      await loadOrgUnitDetails(selectedOrgUnit.id);
      setBanner({ kind: "success", message: t("messages.memberUpdateSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.memberUpdateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function removeMember(userID: string) {
    if (!selectedOrgUnit) return;
    setIsSaving(true);
    try {
      await requestJSON(`/api/org-units/${encodeURIComponent(selectedOrgUnit.id)}/members/${encodeURIComponent(userID)}`, {
        method: "DELETE",
      });
      await loadOrgUnitDetails(selectedOrgUnit.id);
      setBanner({ kind: "success", message: t("messages.memberRemoveSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.memberRemoveError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function handleDelete(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedOrgUnit) return;
    if (getTeamCount(selectedOrgUnit) > 0) {
      setBanner({ kind: "warning", message: t("messages.deleteBlocked") });
      return;
    }
    if (deleteConfirmation.trim() !== selectedOrgUnit.alias) {
      setBanner({ kind: "error", message: t("messages.deleteConfirmAlias") });
      return;
    }

    setIsSaving(true);
    try {
      await requestJSON(`/api/org-units/${encodeURIComponent(selectedOrgUnit.id)}`, {
        method: "DELETE",
      });
      setBanner({ kind: "success", message: t("messages.deleteSuccess") });
      setDeleteOpen(false);
      setDeleteConfirmation("");
      setSelectedOrgUnitID(null);
      await loadOrgUnits(0, query.trim());
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.deleteError")) });
    } finally {
      setIsSaving(false);
    }
  }

  function openCreate() {
    if (!canCreateOrgUnit) {
      setBanner({
        kind: "error",
        message: t("messages.permissionDenied"),
      });
      return;
    }
    setForm(emptyForm);
    setCreateOpen(true);
  }

  function openEdit(orgUnit: OrgUnit) {
    setForm(formFromOrgUnit(orgUnit));
    setEditOpen(true);
  }

  function openMembers() {
    if (!canManage) return;
    setMemberOpen(true);
    setMemberUserID("");
    setMemberUserQuery("");
    setIsMemberUserPickerOpen(false);
    void loadUsers();
  }

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow={`${t("breadcrumb.tenant")} / ${t("title")}`}
        title={
          <span className="inline-flex flex-wrap items-center gap-2">
            <span>{t("title")}</span>
            <a
              href={docsHref}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-1 rounded-md border border-border bg-background px-2 py-1 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              {t("docsLink")}
              <ExternalLink className="size-3.5" aria-hidden="true" />
            </a>
          </span>
        }
        description={t("subtitle")}
        actions={
          <>
          <Button
            variant="outline"
            className="gap-2"
            type="button"
            onClick={() => void loadOrgUnits(offset, query.trim())}
            disabled={isLoading}
          >
            <RefreshCw className={cn("h-4 w-4", isLoading && "animate-spin")} />
            {t("actions.refresh")}
          </Button>
          <Sheet open={createOpen} onOpenChange={setCreateOpen}>
            <Button className="gap-2" type="button" onClick={openCreate} disabled={!canCreateOrgUnit}>
              <Plus className="h-4 w-4" />
              {t("actions.new")}
            </Button>
            <SheetContent side="left" className="w-[94vw] max-w-xl overflow-y-auto">
              <SheetHeader>
                <SheetTitle>{t("create.title")}</SheetTitle>
                <SheetDescription>
                  {t("create.description")}
                </SheetDescription>
              </SheetHeader>
              <OrgUnitForm form={form} setForm={setForm} onSubmit={handleCreate} isSaving={isSaving} submitLabel={t("create.submit")} />
            </SheetContent>
          </Sheet>
          </>
        }
      />

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : banner.kind}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : banner.kind === "warning" ? t("alerts.warning") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.15fr)_minmax(360px,0.85fr)]">
        <AppCard
          title={t("list.title")}
          description={t("list.description")}
          contentClassName="space-y-4"
          actions={
            <div className="flex flex-wrap items-center gap-2">
              <div className="relative">
                <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder={t("filters.searchPlaceholder")}
                  className="w-full ps-9 sm:w-64"
                />
              </div>
              <AppSelectField
                value={statusFilter}
                onValueChange={setStatusFilter}
                ariaLabel={t("filters.filterLabel")}
                triggerClassName="h-10 min-w-36"
                options={[
                  { value: "all", label: t("filters.all") },
                  { value: "normal", label: t("status.normal") },
                  { value: "warning", label: t("status.warning") },
                  { value: "exceeded", label: t("status.exceeded") },
                  { value: "with_teams", label: t("filters.withTeams") },
                  { value: "empty", label: t("filters.noTeams") },
                ]}
              />
              <AppSelectField
                value={sortBy}
                onValueChange={setSortBy}
                ariaLabel={t("filters.sortLabel")}
                triggerClassName="h-10 min-w-32"
                options={[
                  { value: "updated_at", label: t("sort.recent") },
                  { value: "name", label: t("sort.name") },
                  { value: "spend", label: t("sort.spend") },
                  { value: "teams", label: t("sort.teams") },
                ]}
              />
            </div>
          }
        >

          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("table.orgUnit")}</TableHead>
                  <TableHead>{t("table.teams")}</TableHead>
                  <TableHead>{t("table.members")}</TableHead>
                  <TableHead>{t("table.budget")}</TableHead>
                  <TableHead>{t("table.spend")}</TableHead>
                  <TableHead>{t("table.limits")}</TableHead>
                  <TableHead>{t("table.status")}</TableHead>
                  <TableHead>{t("table.updated")}</TableHead>
                  <TableHead className="w-12 text-end"> </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {visibleOrgUnits.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={9} className="py-10 text-center">
                      <EmptyState
                        icon={Building2}
                        title={t("empty.title")}
                        description={t("empty.description")}
                        className="border-0 bg-transparent"
                        primaryAction={
                          <Button className="gap-2" type="button" onClick={openCreate}>
                          <Plus className="h-4 w-4" />
                          {t("empty.action")}
                        </Button>
                        }
                      />
                    </TableCell>
                  </TableRow>
                ) : (
                  visibleOrgUnits.map((orgUnit) => {
                    const status = getBudgetStatus(orgUnit);
                    const active = orgUnit.id === selectedOrgUnitID;
                    return (
                      <TableRow
                        key={orgUnit.id}
                        className={cn("cursor-pointer", active && "bg-muted/60")}
                        onClick={() => setSelectedOrgUnitID(orgUnit.id)}
                      >
                        <TableCell>
                          <div className="min-w-48">
                            <p className="font-medium">{getOrgName(orgUnit)}</p>
                            <p className="text-xs text-muted-foreground">{orgUnit.alias}</p>
                          </div>
                        </TableCell>
                        <TableCell>{resolveTeamCount(orgUnit)}</TableCell>
                        <TableCell>{getMemberCount(orgUnit, orgUnit.id === selectedOrgUnitID ? members.length : undefined)}</TableCell>
                        <TableCell>{formatMoney(getMaxBudget(orgUnit))}</TableCell>
                        <TableCell>
                          <div className="min-w-28">
                            <p>{formatMoney(getSpend(orgUnit))}</p>
                            <Progress value={status.percent} tone={status.kind} />
                          </div>
                        </TableCell>
                        <TableCell>
                          <span className="text-xs text-muted-foreground">
                            {formatLimit(getTpmLimit(orgUnit))} {t("units.tpm")} / {formatLimit(getRpmLimit(orgUnit))} {t("units.rpm")}
                          </span>
                        </TableCell>
                        <TableCell>
                          <BudgetStatusBadge status={status.kind} />
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">{formatDate(orgUnit.updated_at)}</TableCell>
                        <TableCell className="text-end" onClick={(event) => event.stopPropagation()}>
                          <RowActionsMenu
                            label={t("actions.viewDetails")}
                            disabled={!canManage}
                            actions={[
                              { label: t("actions.viewDetails"), onSelect: () => setSelectedOrgUnitID(orgUnit.id) },
                              { label: t("actions.edit"), onSelect: () => openEdit(orgUnit) },
                              { label: t("actions.manageMembers"), onSelect: openMembers },
                              {
                                label: t("actions.delete"),
                                destructive: true,
                                disabled: !canDeleteOrgUnit,
                                onSelect: () => {
                                  setSelectedOrgUnitID(orgUnit.id);
                                  setDeleteOpen(true);
                                },
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
          </div>

          <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
            <p>{t("pagination.showing", { from, to, total })}</p>
            <div className="flex items-center gap-2">
              <Badge variant="outline">{t("pagination.rows", { count: PAGE_SIZE })}</Badge>
              <Button
                variant="outline"
                size="sm"
                onClick={() => void loadOrgUnits(Math.max(0, offset - PAGE_SIZE), query.trim())}
                disabled={!canPrev || isLoading}
              >
                {t("pagination.prev")}
              </Button>
              <Badge variant="secondary">{Math.floor(offset / PAGE_SIZE) + 1}</Badge>
              <Button
                variant="outline"
                size="sm"
                onClick={() => void loadOrgUnits(offset + PAGE_SIZE, query.trim())}
                disabled={!canNext || isLoading}
              >
                {t("pagination.next")}
              </Button>
            </div>
          </div>
        </AppCard>

        <aside className="space-y-4">
          {selectedOrgUnit && selectedSummary ? (
            <>
              <AppCard contentClassName="space-y-4">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="text-sm text-muted-foreground">{t("detail.breadcrumb")}</p>
                    <h2 className="mt-1 truncate text-xl font-semibold">{getOrgName(selectedOrgUnit)}</h2>
                  </div>
                  <BudgetStatusBadge status={selectedSummary.status.kind} />
                </div>
                <div className="mt-4 flex flex-wrap gap-2">
                  <Button variant="outline" size="sm" className="gap-2" onClick={() => openEdit(selectedOrgUnit)} disabled={!canManage}>
                    <Edit3 className="h-4 w-4" />
                    {t("actions.edit")}
                  </Button>
                  <Button variant="outline" size="sm" className="gap-2" onClick={openMembers} disabled={!canManage}>
                    <Users className="h-4 w-4" />
                    {t("actions.members")}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-2 text-destructive"
                    onClick={() => setDeleteOpen(true)}
                    disabled={!canDeleteOrgUnit}
                  >
                    <Trash2 className="h-4 w-4" />
                    {t("actions.delete")}
                  </Button>
                </div>
              </AppCard>

              <div className="grid gap-3 sm:grid-cols-2">
                <KpiCard icon={WalletCards} label={t("kpi.totalSpend")} value={formatMoney(selectedSummary.spend)} />
                <KpiCard icon={Gauge} label={t("kpi.remaining")} value={formatMoney(selectedSummary.remaining)} />
                <KpiCard icon={Building2} label={t("kpi.teams")} value={String(selectedSummary.teams)} />
                <KpiCard icon={ShieldCheck} label={t("kpi.membersAdmins")} value={`${selectedSummary.members}/${selectedSummary.admins}`} />
              </div>

              <AppCard
                title={
                  <span className="flex items-center gap-2">
                    <WalletCards className="h-4 w-4" aria-hidden="true" />
                    <span>{t("budget.title")}</span>
                  </span>
                }
                contentClassName="space-y-4"
              >
                  <div>
                    <div className="mb-2 flex items-center justify-between text-sm">
                      <span>{t("budget.used", { amount: formatMoney(selectedSummary.spend) })}</span>
                      <span className="text-muted-foreground">{formatMoney(selectedSummary.maxBudget)}</span>
                    </div>
                    <Progress value={selectedSummary.status.percent} tone={selectedSummary.status.kind} />
                  </div>
                  <dl className="grid grid-cols-2 gap-3 text-sm">
                    <InfoItem label={t("budget.soft")} value={formatMoney(selectedSummary.softBudget)} />
                    <InfoItem label={t("budget.duration")} value={selectedSummary.duration || "30d"} />
                    <InfoItem label={t("budget.tpm")} value={formatLimit(selectedSummary.tpm)} />
                    <InfoItem label={t("budget.rpm")} value={formatLimit(selectedSummary.rpm)} />
                  </dl>
                  {selectedSummary.status.kind !== "normal" ? (
                    <Alert variant={selectedSummary.status.kind === "exceeded" ? "destructive" : "warning"}>
                      <AlertTitle>{selectedSummary.status.kind === "exceeded" ? t("budget.exceededTitle") : t("budget.warningTitle")}</AlertTitle>
                      <AlertDescription>
                        {t("budget.warningDescription")}
                      </AlertDescription>
                    </Alert>
                  ) : null}
              </AppCard>

              <AppCard
                title={
                  <span className="flex items-center gap-2">
                    <Building2 className="h-4 w-4" aria-hidden="true" />
                    <span>{t("teams.title")}</span>
                  </span>
                }
              >
                  <MetadataTeams
                    orgUnit={selectedOrgUnit}
                    teamCount={resolveTeamCount(selectedOrgUnit)}
                    teams={resolveTeams(selectedOrgUnit)}
                  />
              </AppCard>

              <AppCard
                title={
                  <span className="flex items-center gap-2">
                    <Users className="h-4 w-4" aria-hidden="true" />
                    <span>{t("members.title")}</span>
                  </span>
                }
                contentClassName="space-y-3"
              >
                  {isDetailLoading ? (
                    <p className="text-sm text-muted-foreground">{t("members.loading")}</p>
                  ) : members.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{t("members.empty")}</p>
                  ) : (
                    members.map((member) => (
                      <div key={member.id} className="flex items-center justify-between gap-3 rounded-lg border p-3">
                        <div className="min-w-0">
                          <p className="truncate text-sm font-medium">{userNameByID.get(member.user_id) || t("members.user")}</p>
                          <p className="text-xs text-muted-foreground">{t("members.added", { date: formatDate(member.created_at) })}</p>
                        </div>
                        <Badge variant={member.role === "org_admin" ? "default" : "secondary"}>{member.role}</Badge>
                      </div>
                    ))
                  )}
              </AppCard>

              <AppCard
                title={
                  <span className="flex items-center gap-2">
                    <CalendarClock className="h-4 w-4" aria-hidden="true" />
                    <span>{t("audit.title")}</span>
                  </span>
                }
                contentClassName="space-y-3 text-sm"
              >
                  <InfoItem label={t("audit.created")} value={formatDateTime(selectedOrgUnit.created_at)} />
                  <InfoItem label={t("audit.updated")} value={formatDateTime(selectedOrgUnit.updated_at)} />
                  <pre className="max-h-40 overflow-auto rounded-lg bg-muted p-3 text-xs text-muted-foreground">
                    {JSON.stringify(selectedOrgUnit.metadata ?? {}, null, 2)}
                  </pre>
              </AppCard>
            </>
          ) : (
            <EmptyState title={t("detail.empty")} />
          )}
        </aside>
      </section>

      <Sheet open={editOpen} onOpenChange={setEditOpen}>
        <SheetContent side="right" className="w-[94vw] max-w-xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{t("edit.title")}</SheetTitle>
            <SheetDescription>
              {t("edit.description")}
            </SheetDescription>
          </SheetHeader>
          <OrgUnitForm form={form} setForm={setForm} onSubmit={handleUpdate} isSaving={isSaving} submitLabel={t("edit.submit")} />
        </SheetContent>
      </Sheet>

      <Sheet open={memberOpen} onOpenChange={setMemberOpen}>
        <SheetContent side="right" className="w-[94vw] max-w-xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{t("members.manageTitle")}</SheetTitle>
            <SheetDescription>
              {t("members.manageDescription")}
            </SheetDescription>
          </SheetHeader>

          <form className="mt-6 space-y-4" onSubmit={handleAddMember}>
            <div className="grid gap-2">
              <Label htmlFor="member-user">{t("members.user")}</Label>
              <div className="relative">
                <Input
                  id="member-user"
                  value={memberUserQuery}
                  onFocus={() => setIsMemberUserPickerOpen(true)}
                  onBlur={() => setTimeout(() => setIsMemberUserPickerOpen(false), 120)}
                  onChange={(event) => {
                    setMemberUserQuery(event.target.value);
                    setMemberUserID("");
                    setIsMemberUserPickerOpen(true);
                  }}
                  placeholder={t("members.user")}
                  autoComplete="off"
                />
                {isMemberUserPickerOpen ? (
                  <div className="absolute z-50 mt-1 max-h-56 w-full overflow-auto rounded-md border bg-popover p-1 shadow-md">
                    {memberUserOptions.length === 0 ? (
                      <p className="px-2 py-1.5 text-sm text-muted-foreground">{t("common.notSet")}</p>
                    ) : (
                      memberUserOptions.map((user) => (
                        <button
                          key={user.principal_id}
                          type="button"
                          className="w-full rounded-sm px-2 py-1.5 text-start text-sm hover:bg-muted"
                          onMouseDown={(event) => event.preventDefault()}
                          onClick={() => {
                            setMemberUserID(user.principal_id);
                            setMemberUserQuery(user.display_name || user.secondary || "");
                            setIsMemberUserPickerOpen(false);
                          }}
                        >
                          {user.display_name || user.secondary || t("members.user")}
                        </button>
                      ))
                    )}
                  </div>
                ) : null}
              </div>
            </div>
            <div>
              <AppSelectField
                label={t("members.role")}
                value={memberRole}
                onValueChange={setMemberRole}
                triggerClassName="h-10"
                options={[
                  { value: "org_admin", label: t("roles.orgAdmin") },
                  { value: "member", label: t("roles.member") },
                ]}
              />
            </div>
            <Button type="submit" disabled={isSaving || !canManage} className="w-full">
              {t("members.add")}
            </Button>
          </form>

          <div className="mt-6 space-y-3">
            {members.map((member) => (
              <div key={member.id} className="rounded-lg border p-3">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{userNameByID.get(member.user_id) || t("members.user")}</p>
                    <p className="text-xs text-muted-foreground">{formatDateTime(member.updated_at)}</p>
                  </div>
                  <Badge variant={member.role === "org_admin" ? "default" : "secondary"}>{member.role}</Badge>
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => void updateMemberRole(member.user_id, member.role === "org_admin" ? "member" : "org_admin")}
                    disabled={isSaving || !canManage}
                  >
                    {member.role === "org_admin" ? t("members.makeMember") : t("members.makeOrgAdmin")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="text-destructive"
                    onClick={() => void removeMember(member.user_id)}
                    disabled={isSaving || !canManage}
                  >
                    {t("members.remove")}
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </SheetContent>
      </Sheet>

      <Sheet open={deleteOpen} onOpenChange={setDeleteOpen}>
        <SheetContent side="right" className="w-[94vw] max-w-md">
          <SheetHeader>
            <SheetTitle>{t("delete.title")}</SheetTitle>
            <SheetDescription>
              {t("delete.description")}
            </SheetDescription>
          </SheetHeader>
          {selectedOrgUnit ? (
            <form className="mt-6 space-y-4" onSubmit={handleDelete}>
              {resolveTeamCount(selectedOrgUnit) > 0 ? (
                <Alert variant="warning">
                  <AlertTitle>{t("delete.childTeamsTitle")}</AlertTitle>
                  <AlertDescription>
                    {t("delete.childTeamsDescription", { name: getOrgName(selectedOrgUnit) })}
                  </AlertDescription>
                </Alert>
              ) : null}
              <div className="grid gap-2">
                <Label htmlFor="delete-confirm">{t("delete.confirmLabel", { alias: selectedOrgUnit.alias })}</Label>
                <Input
                  id="delete-confirm"
                  value={deleteConfirmation}
                  onChange={(event) => setDeleteConfirmation(event.target.value)}
                  placeholder={selectedOrgUnit.alias}
                />
              </div>
              <Button type="submit" variant="destructive" disabled={isSaving || !canDeleteOrgUnit} className="w-full">
                {t("delete.submit")}
              </Button>
            </form>
          ) : null}
        </SheetContent>
      </Sheet>
    </div>
  );
}

function OrgUnitForm({
  form,
  setForm,
  onSubmit,
  isSaving,
  submitLabel,
}: {
  form: FormState;
  setForm: (next: FormState) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  isSaving: boolean;
  submitLabel: string;
}) {
  const t = useTranslations("orgUnits");

  function update<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm({ ...form, [key]: value });
  }

  return (
    <form className="mt-6 space-y-5" onSubmit={onSubmit}>
      <IdentifierField id="org-alias" label={t("form.alias")} value={form.alias} onChange={(event) => update("alias", event.target.value)} placeholder="engineering" required />
      <div className="grid gap-2">
        <Label htmlFor="org-display-name">{t("form.displayName")}</Label>
        <Input
          id="org-display-name"
          value={form.displayName}
          onChange={(event) => update("displayName", event.target.value)}
          placeholder="Engineering"
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="org-description">{t("form.description")}</Label>
        <Input
          id="org-description"
          value={form.description}
          onChange={(event) => update("description", event.target.value)}
          placeholder="Owner scope for platform and product engineering teams"
        />
      </div>

      <Button type="submit" disabled={isSaving} className="w-full">
        {isSaving ? t("actions.saving") : submitLabel}
      </Button>
    </form>
  );
}

function KpiCard({
  icon: Icon,
  label,
  value,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
}) {
  return (
    <AppCard className="min-w-0" contentClassName="pt-4 md:pt-4">
      <div className="flex min-w-0 items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <Icon className="size-4" aria-hidden="true" />
        </div>
        <div className="min-w-0">
          <p className="text-xs text-muted-foreground">{label}</p>
          <p className="truncate text-lg font-semibold text-foreground">{value}</p>
        </div>
      </div>
    </AppCard>
  );
}

function MetadataTeams({
  orgUnit,
  teamCount,
  teams: orgTeams,
}: {
  orgUnit: OrgUnit;
  teamCount: number;
  teams: TeamSummary[];
}) {
  const t = useTranslations("orgUnits");
  const metadataTeams = getMetadataArray(orgUnit.metadata, "teams");

  if (orgTeams.length > 0) {
    return (
      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>نام تیم</TableHead>
              <TableHead>نوع ورک‌اسپیس</TableHead>
              <TableHead>تعداد اعضا</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {orgTeams.map((team) => (
              <TableRow key={team.id}>
                <TableCell className="font-medium">{getTeamName(team)}</TableCell>
                <TableCell>{getTeamWorkspaceType(team)}</TableCell>
                <TableCell>{getTeamMemberCount(team)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  }

  if (metadataTeams.length === 0) {
    return (
      <div className="space-y-2 text-sm text-muted-foreground">
        <p>{teamCount > 0 ? t("teams.summaryCount", { count: teamCount }) : t("teams.emptyMetadata")}</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {metadataTeams.map((team, index) => {
        const record = asRecord(team);
        return (
          <div key={`${toText(record.id) || index}`} className="rounded-lg border p-3">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{toText(record.display_name) || toText(record.alias) || t("teams.fallbackName", { index: index + 1 })}</p>
                <p className="text-xs text-muted-foreground">{toText(record.workspace_type) || "general"}</p>
              </div>
              <Badge variant="secondary">{formatMoney(toNumber(record.spend))}</Badge>
            </div>
            <div className="mt-3 grid grid-cols-2 gap-2 text-xs text-muted-foreground">
              <span>{t("teams.budget", { amount: formatMoney(toNumber(record.max_budget)) })}</span>
              <span>{t("teams.models", { models: toText(record.models) || t("common.inherited") })}</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function getTeamName(team: TeamSummary) {
  return team.display_name || toText(team.metadata?.display_name) || team.alias;
}

function getTeamWorkspaceType(team: TeamSummary) {
  return (team.workspace_type || toText(team.metadata?.workspace_type) || "general").trim().toLowerCase();
}

function getTeamMemberCount(team: TeamSummary) {
  return toNumber(team.metadata?.member_count ?? team.metadata?.members_count);
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="truncate font-medium">{value}</dd>
    </div>
  );
}

function BudgetStatusBadge({ status }: { status: "normal" | "warning" | "exceeded" }) {
  const t = useTranslations("orgUnits");
  if (status === "exceeded") return <StatusBadge status="error">{t("status.exceeded")}</StatusBadge>;
  if (status === "warning") return <StatusBadge status="warning">{t("status.warning")}</StatusBadge>;
  return <StatusBadge status="success">{t("status.normal")}</StatusBadge>;
}

function Progress({ value, tone }: { value: number; tone: "normal" | "warning" | "exceeded" }) {
  return (
    <div className="h-2 overflow-hidden rounded-full bg-muted">
      <div
        className={cn(
          "h-full rounded-full transition-all",
          tone === "exceeded" ? "bg-destructive" : tone === "warning" ? "bg-warning" : "bg-success",
        )}
        style={{ width: `${Math.min(100, Math.max(0, value))}%` }}
      />
    </div>
  );
}

function buildSummary(orgUnit: OrgUnit, members: OrgUnitMember[]) {
  const maxBudget = getMaxBudget(orgUnit);
  const spend = getSpend(orgUnit);
  const remaining = maxBudget > 0 ? Math.max(0, maxBudget - spend) : 0;
  return {
    spend,
    maxBudget,
    softBudget: getSoftBudget(orgUnit),
    duration: toText(orgUnit.metadata?.budget_duration),
    remaining,
    teams: getTeamCount(orgUnit),
    members: members.length || getMemberCount(orgUnit),
    admins: members.filter((member) => member.role === "org_admin").length || getAdminCount(orgUnit),
    tpm: getTpmLimit(orgUnit),
    rpm: getRpmLimit(orgUnit),
    status: getBudgetStatus(orgUnit),
  };
}

function getBudgetStatus(orgUnit: OrgUnit): { kind: "normal" | "warning" | "exceeded"; percent: number } {
  const maxBudget = getMaxBudget(orgUnit);
  const softBudget = getSoftBudget(orgUnit);
  const spend = getSpend(orgUnit);
  if (maxBudget <= 0) return { kind: "normal", percent: 0 };
  const percent = (spend / maxBudget) * 100;
  if (spend > maxBudget) return { kind: "exceeded", percent };
  if ((softBudget > 0 && spend >= softBudget) || percent >= 80) return { kind: "warning", percent };
  return { kind: "normal", percent };
}

function getOrgName(orgUnit: OrgUnit) {
  return orgUnit.display_name || toText(orgUnit.metadata?.display_name) || orgUnit.alias;
}

function getMaxBudget(orgUnit: OrgUnit) {
  return toNumber(orgUnit.metadata?.max_budget ?? orgUnit.metadata?.budget_max ?? orgUnit.metadata?.monthly_budget);
}

function getSoftBudget(orgUnit: OrgUnit) {
  return toNumber(orgUnit.metadata?.soft_budget);
}

function getSpend(orgUnit: OrgUnit) {
  return toNumber(orgUnit.metadata?.spend ?? orgUnit.metadata?.current_spend ?? orgUnit.metadata?.total_spend);
}

function getTpmLimit(orgUnit: OrgUnit) {
  return toNumber(orgUnit.metadata?.tpm_limit);
}

function getRpmLimit(orgUnit: OrgUnit) {
  return toNumber(orgUnit.metadata?.rpm_limit);
}

function getTeamCount(orgUnit: OrgUnit) {
  return toNumber(orgUnit.metadata?.team_count ?? orgUnit.metadata?.teams_count) || getMetadataArray(orgUnit.metadata, "teams").length;
}

function getMemberCount(orgUnit: OrgUnit, loadedCount?: number) {
  if (loadedCount !== undefined) return loadedCount;
  return toNumber(orgUnit.metadata?.member_count ?? orgUnit.metadata?.members_count);
}

function getAdminCount(orgUnit: OrgUnit) {
  return toNumber(orgUnit.metadata?.admin_count ?? orgUnit.metadata?.admins_count);
}

function getChildBudgetTotal(orgUnit: OrgUnit): number {
  const direct = toNumber(orgUnit.metadata?.child_budget_total ?? orgUnit.metadata?.team_budget_total);
  if (direct > 0) return direct;
  return getMetadataArray(orgUnit.metadata, "teams").reduce<number>(
    (sum, team) => sum + toNumber(asRecord(team).max_budget),
    0,
  );
}

function formFromOrgUnit(orgUnit: OrgUnit): FormState {
  return {
    alias: orgUnit.alias,
    displayName: getOrgName(orgUnit),
    description: toText(orgUnit.metadata?.description),
  };
}

function buildPayload(form: FormState, previousMetadata?: Record<string, unknown>) {
  const metadataWithoutCostCenter = { ...(previousMetadata ?? {}) };
  delete metadataWithoutCostCenter.cost_center;
  delete metadataWithoutCostCenter.costCenter;
  const metadata = {
    ...metadataWithoutCostCenter,
    display_name: form.displayName.trim() || form.alias.trim(),
    description: form.description.trim(),
  };

  return {
    alias: form.alias.trim(),
    display_name: form.displayName.trim(),
    metadata,
  };
}

function validateForm(
  form: FormState,
  orgUnits: OrgUnit[],
  t: ReturnType<typeof useTranslations>,
  currentID?: string,
) {
  const alias = form.alias.trim();
  if (!alias) return t("messages.aliasRequired");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{1,62}$/.test(alias)) {
    return t("messages.aliasPattern");
  }
  const duplicate = orgUnits.some((orgUnit) => orgUnit.id !== currentID && orgUnit.alias.toLowerCase() === alias.toLowerCase());
  if (duplicate) return t("messages.aliasUnique");
  return null;
}

function getMetadataArray(metadata: Record<string, unknown> | undefined, key: string): unknown[] {
  const value = metadata?.[key];
  return Array.isArray(value) ? value : [];
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}

function toNumber(value: unknown) {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value.replace(/,/g, ""));
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function toText(value: unknown) {
  if (typeof value === "string") return value;
  if (typeof value === "number") return String(value);
  if (Array.isArray(value)) return value.join(", ");
  return "";
}

function formatMoney(value: number) {
  if (value <= 0) return "-";
  return new Intl.NumberFormat("fa-IR", {
    style: "currency",
    currency: "IRR",
    maximumFractionDigits: 0,
  }).format(value);
}

function formatLimit(value: number) {
  if (value <= 0) return "-";
  return new Intl.NumberFormat("en-US", { notation: "compact", maximumFractionDigits: 1 }).format(value);
}

function formatDate(value: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
    month: "short",
    day: "numeric",
  }).format(date);
}

function formatDateTime(value: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
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
  const candidate = payload as { error?: { message?: string } | string; message?: string };
  if (typeof candidate.error === "string") return candidate.error;
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}
