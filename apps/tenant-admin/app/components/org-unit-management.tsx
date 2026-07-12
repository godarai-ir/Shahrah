"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  Building2,
  CalendarClock,
  Edit3,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
  Trash2,
  Users,
  WalletCards,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { AppCard, AppSelectField, CommercialPlanUpgradeDialog, CompactValue, DefinitionList, DefinitionListItem, EmptyState, FilterCombobox, HelpLink, IdentifierField, MetricCard, OperationNotice, PageHeader, RowActionsMenu, TablePagination } from "@/components/app-ui";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
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
import { appendDisplaySuffix, formatCompactNumber } from "@/lib/compact-number";
import { getCurrentJalaliMonthRange } from "@/app/billing/_shared";
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

type UsagePeriod = "current_month" | "last_30_days" | "last_7_days";

type OrgUnitUsageSummary = {
  request_count: number;
  cost_irr: number;
};

type OrgUnitUsageSummaryResponse = {
  data?: Partial<OrgUnitUsageSummary>;
};

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
} | null;

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
const FEATURE_ORG_UNITS_ENABLED = "org_units.enabled";

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
  const locale = useLocale();
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
  const [usagePeriod, setUsagePeriod] = useState<UsagePeriod>("current_month");
  const [usageSummary, setUsageSummary] = useState<OrgUnitUsageSummary | null>(null);
  const [usageError, setUsageError] = useState<string | null>(null);
  const [isUsageLoading, setIsUsageLoading] = useState(false);
  const [usageReloadKey, setUsageReloadKey] = useState(0);

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [memberOpen, setMemberOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [memberUserID, setMemberUserID] = useState("");
  const [memberRole, setMemberRole] = useState("org_admin");
  const [deleteConfirmation, setDeleteConfirmation] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
  const [isUpgradeDialogOpen, setIsUpgradeDialogOpen] = useState(false);

  const canManage = ["tenant_admin", "proxy_admin", "org_admin"].includes(currentUserRole) || canAnyPermission(permissions, ["org_unit:manage"]);
  const canCreateOrgUnit = canManage || canAnyPermission(permissions, ["org_unit:create"]);
  const canDeleteOrgUnit = canManage || canAnyPermission(permissions, ["org_unit:delete"]);
  const isOrgUnitsFeatureEnabled = hasLoadedEntitlements
    ? readBooleanFeature(entitlements?.features?.[FEATURE_ORG_UNITS_ENABLED], false)
    : null;

  useEffect(() => {
    let isMounted = true;

    async function loadEntitlements() {
      try {
        const response = await requestJSON<EffectiveEntitlements>("/api/entitlements");
        if (isMounted) setEntitlements(response);
      } catch {
        if (isMounted) setEntitlements(null);
      } finally {
        if (isMounted) setHasLoadedEntitlements(true);
      }
    }

    void loadEntitlements();
    return () => {
      isMounted = false;
    };
  }, []);

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
    const controller = new AbortController();
    const range = getUsagePeriodRange(usagePeriod);
    const search = new URLSearchParams({
      org_unit_id: selectedOrgUnitID,
      from: range.from.toISOString(),
      to: range.to.toISOString(),
    });

    async function loadUsage() {
      setIsUsageLoading(true);
      setUsageError(null);
      try {
        const payload = await requestJSON<OrgUnitUsageSummaryResponse>(`/api/logs/summary?${search.toString()}`, { signal: controller.signal });
        setUsageSummary({
          request_count: toNumber(payload.data?.request_count),
          cost_irr: toNumber(payload.data?.cost_irr),
        });
      } catch (error: unknown) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setUsageSummary(null);
        setUsageError(normalizeError(error, t("usage.loadError")));
      } finally {
        if (!controller.signal.aborted) setIsUsageLoading(false);
      }
    }

    void loadUsage();

    return () => controller.abort();
  }, [selectedOrgUnitID, t, usagePeriod, usageReloadKey]);

  useEffect(() => {
    if (!selectedOrgUnitID) return;
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
      const teamCount = teamCountByOrgUnit[orgUnit.id] ?? getTeamCount(orgUnit);
      if (statusFilter === "all") return true;
      if (statusFilter === "with_teams") return teamCount > 0;
      if (statusFilter === "empty") return teamCount === 0;
      return true;
    });

    return [...filtered].sort((a, b) => {
      if (sortBy === "name") return getOrgName(a).localeCompare(getOrgName(b));
      if (sortBy === "teams") {
        const bCount = teamCountByOrgUnit[b.id] ?? getTeamCount(b);
        const aCount = teamCountByOrgUnit[a.id] ?? getTeamCount(a);
        return bCount - aCount;
      }
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
  const memberUserOptions = useMemo(
    () => users.slice(0, 100).map((user) => ({
      value: user.principal_id,
      label: user.secondary ? `${user.display_name || t("members.user")} — ${user.secondary}` : user.display_name || t("members.user"),
    })),
    [t, users],
  );
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
    if (isOrgUnitsFeatureEnabled === false) {
      setIsUpgradeDialogOpen(true);
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
    void loadUsers();
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("title")}
        titleAddon={<HelpLink href={docsHref} />}
        description={t("subtitle")}
        actions={
          <>
          <Button
            variant="outline"
            type="button"
            onClick={() => void loadOrgUnits(offset, query.trim())}
            disabled={isLoading}
          >
            {isLoading ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
            {t("actions.refresh")}
          </Button>
          <Sheet open={createOpen} onOpenChange={setCreateOpen}>
            <Button type="button" onClick={openCreate} disabled={!canCreateOrgUnit || !hasLoadedEntitlements}>
              <Plus data-icon="inline-start" />
              {t("actions.new")}
            </Button>
            <SheetContent side="left" className="w-full max-w-xl overflow-y-auto sm:w-3/4">
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

      <OperationNotice notice={banner} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} />

      <section className="grid min-w-0 gap-6 2xl:grid-cols-[minmax(0,3fr)_minmax(22rem,1fr)]">
        <AppCard
          className="min-w-0"
          contentClassName="flex min-w-0 flex-col gap-4"
        >
          <div className="grid w-full gap-2 sm:grid-cols-[minmax(14rem,1fr)_auto_auto] 2xl:flex 2xl:flex-wrap">
              <div className="relative min-w-0">
                <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder={t("filters.searchPlaceholder")}
                  className="w-full ps-9 2xl:w-64"
                />
              </div>
              <AppSelectField
                value={statusFilter}
                onValueChange={setStatusFilter}
                ariaLabel={t("filters.filterLabel")}
                triggerClassName="h-10 w-full min-w-36"
                options={[
                  { value: "all", label: t("filters.all") },
                  { value: "with_teams", label: t("filters.withTeams") },
                  { value: "empty", label: t("filters.noTeams") },
                ]}
              />
              <AppSelectField
                value={sortBy}
                onValueChange={setSortBy}
                ariaLabel={t("filters.sortLabel")}
                triggerClassName="h-10 w-full min-w-32"
                options={[
                  { value: "updated_at", label: t("sort.recent") },
                  { value: "name", label: t("sort.name") },
                  { value: "teams", label: t("sort.teams") },
                ]}
              />
          </div>

          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("table.orgUnit")}</TableHead>
                  <TableHead>{t("table.teams")}</TableHead>
                  <TableHead>{t("table.members")}</TableHead>
                  <TableHead>{t("table.updated")}</TableHead>
                  <TableHead className="w-12 text-end"> </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {visibleOrgUnits.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="py-10 text-center">
                      <EmptyState
                        variant="table"
                        icon={Building2}
                        title={t("empty.title")}
                        description={t("empty.description")}
                        primaryAction={
                          canCreateOrgUnit ? <Button type="button" onClick={openCreate} disabled={!hasLoadedEntitlements}>
                          <Plus data-icon="inline-start" />
                          {t("empty.action")}
                        </Button> : null
                        }
                      />
                    </TableCell>
                  </TableRow>
                ) : (
                  visibleOrgUnits.map((orgUnit) => {
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
                        <TableCell className="text-sm text-muted-foreground">{formatDate(orgUnit.updated_at)}</TableCell>
                        <TableCell className="text-end" onClick={(event) => event.stopPropagation()}>
                          <RowActionsMenu
                            label={t("actions.openMenu")}
                            actions={[
                              { label: t("actions.edit"), icon: <Edit3 />, disabled: !canManage, onSelect: () => openEdit(orgUnit) },
                              { label: t("actions.manageMembers"), icon: <Users />, disabled: !canManage, onSelect: openMembers },
                              {
                                label: t("actions.delete"),
                                icon: <Trash2 />,
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

          <TablePagination
            label={t("title")}
            summary={t("pagination.showing", { from, to, total })}
            pageSize={t("pagination.rows", { count: PAGE_SIZE })}
            page={Math.floor(offset / PAGE_SIZE) + 1}
            previousLabel={t("pagination.prev")}
            nextLabel={t("pagination.next")}
            canPrevious={canPrev}
            canNext={canNext}
            disabled={isLoading}
            onPrevious={() => void loadOrgUnits(Math.max(0, offset - PAGE_SIZE), query.trim())}
            onNext={() => void loadOrgUnits(offset + PAGE_SIZE, query.trim())}
          />
        </AppCard>

        <aside className="flex min-w-0 flex-col gap-4">
          {selectedOrgUnit && selectedSummary ? (
            <>
              <div className="grid grid-cols-2 gap-3">
                <MetricCard icon={Building2} label={t("kpi.teams")} value={String(selectedSummary.teams)} />
                <MetricCard icon={ShieldCheck} label={t("kpi.membersAdmins")} value={`${selectedSummary.members}/${selectedSummary.admins}`} />
              </div>

              <AppCard
                title={
                  <span className="flex items-center gap-2">
                    <WalletCards aria-hidden="true" />
                    <span>{t("usage.title")}</span>
                  </span>
                }
                description={t("usage.description")}
                actions={
                  <AppSelectField
                    value={usagePeriod}
                    onValueChange={(value) => setUsagePeriod(value as UsagePeriod)}
                    options={[
                      { value: "current_month", label: t("usage.periods.current_month") },
                      { value: "last_30_days", label: t("usage.periods.last_30_days") },
                      { value: "last_7_days", label: t("usage.periods.last_7_days") },
                    ]}
                    ariaLabel={t("usage.periodLabel")}
                    triggerClassName="h-9"
                    disabled={isUsageLoading}
                  />
                }
                contentClassName="flex flex-col gap-4"
              >
                  {isUsageLoading ? (
                    <div className="grid gap-4 sm:grid-cols-2" aria-label={t("usage.loading")}>
                      <Skeleton className="h-24 w-full" />
                      <Skeleton className="h-24 w-full" />
                    </div>
                  ) : usageError ? (
                    <Alert variant="warning">
                      <AlertTitle>{t("usage.unavailableTitle")}</AlertTitle>
                      <AlertDescription className="flex flex-col items-start gap-3">
                        <span>{t("usage.unavailableDescription")}</span>
                        <Button type="button" variant="outline" size="sm" onClick={() => setUsageReloadKey((value) => value + 1)}>
                          <RefreshCw data-icon="inline-start" />
                          {t("usage.retry")}
                        </Button>
                      </AlertDescription>
                    </Alert>
                  ) : usageSummary ? (
                    <>
                      <dl className="grid grid-cols-2 gap-3">
                        <div className="rounded-lg border border-border p-4">
                          <dt className="text-sm text-muted-foreground">{t("usage.cost")}</dt>
                          <dd className="mt-2 text-2xl font-bold tracking-tight">
                            <CompactValue
                              dir="ltr"
                              display={appendDisplaySuffix(formatCompactNumber(usageSummary.cost_irr, { locale }), t("usage.currencySuffix"))}
                            />
                          </dd>
                        </div>
                        <div className="rounded-lg border border-border p-4">
                          <dt className="text-sm text-muted-foreground">{t("usage.requests")}</dt>
                          <dd className="mt-2 text-2xl font-bold tracking-tight">
                            <CompactValue dir="ltr" display={formatCompactNumber(usageSummary.request_count, { locale })} />
                          </dd>
                        </div>
                      </dl>
                      {usageSummary.request_count === 0 ? (
                        <p className="text-sm text-muted-foreground">{t("usage.empty")}</p>
                      ) : null}
                      <p className="text-xs leading-5 text-muted-foreground">{t("usage.attributionNote")}</p>
                    </>
                  ) : null}
              </AppCard>

              <AppCard
                title={
                  <span className="flex items-center gap-2">
                    <Building2 aria-hidden="true" />
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
                    <Users aria-hidden="true" />
                    <span>{t("members.title")}</span>
                  </span>
                }
                contentClassName="flex flex-col gap-3"
              >
                  {isDetailLoading ? (
                    <div className="flex flex-col gap-3" aria-label={t("members.loading")}>
                      <Skeleton className="h-14 w-full" />
                      <Skeleton className="h-14 w-full" />
                    </div>
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
                    <CalendarClock aria-hidden="true" />
                    <span>{t("audit.title")}</span>
                  </span>
                }
                contentClassName="flex flex-col gap-3 text-sm"
              >
                  <DefinitionList columns={2}>
                    <DefinitionListItem label={t("audit.created")} value={formatDateTime(selectedOrgUnit.created_at)} />
                    <DefinitionListItem label={t("audit.updated")} value={formatDateTime(selectedOrgUnit.updated_at)} />
                  </DefinitionList>
                  <pre dir="ltr" className="max-h-40 overflow-auto rounded-lg bg-muted p-3 text-left text-xs text-muted-foreground">
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
        <SheetContent side="left" className="w-full max-w-xl overflow-y-auto sm:w-3/4">
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
        <SheetContent side="left" className="w-full max-w-xl overflow-y-auto sm:w-3/4">
          <SheetHeader>
            <SheetTitle>{t("members.manageTitle")}</SheetTitle>
            <SheetDescription>
              {t("members.manageDescription")}
            </SheetDescription>
          </SheetHeader>

          <form className="mt-6 flex flex-col gap-4" onSubmit={handleAddMember}>
            <div className="flex flex-col gap-2">
              <Label htmlFor="member-user">{t("members.user")}</Label>
              <FilterCombobox
                id="member-user"
                options={memberUserOptions}
                value={memberUserID}
                onValueChange={setMemberUserID}
                placeholder={t("members.userPlaceholder")}
                emptyText={t("common.notSet")}
                className="h-10"
              />
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

          <div className="mt-6 flex flex-col gap-3">
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
        <SheetContent side="left" className="w-full max-w-md sm:w-3/4">
          <SheetHeader>
            <SheetTitle>{t("delete.title")}</SheetTitle>
            <SheetDescription>
              {t("delete.description")}
            </SheetDescription>
          </SheetHeader>
          {selectedOrgUnit ? (
            <form className="mt-6 flex flex-col gap-4" onSubmit={handleDelete}>
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

      <CommercialPlanUpgradeDialog open={isUpgradeDialogOpen} onOpenChange={setIsUpgradeDialogOpen} title={t("upgradeDialog.title")} description={t("upgradeDialog.description")} cancelLabel={t("upgradeDialog.cancel")} actionLabel={t("upgradeDialog.action")} />
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
    <form className="mt-6 flex flex-col gap-5" onSubmit={onSubmit}>
      <IdentifierField id="org-alias" label={t("form.alias")} value={form.alias} onChange={(event) => update("alias", event.target.value)} placeholder={t("form.aliasPlaceholder")} required />
      <div className="grid gap-2">
        <Label htmlFor="org-display-name">{t("form.displayName")}</Label>
        <Input
          id="org-display-name"
          value={form.displayName}
          onChange={(event) => update("displayName", event.target.value)}
          placeholder={t("form.displayNamePlaceholder")}
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="org-description">{t("form.description")}</Label>
        <Input
          id="org-description"
          value={form.description}
          onChange={(event) => update("description", event.target.value)}
          placeholder={t("form.descriptionPlaceholder")}
        />
      </div>

      <Button type="submit" disabled={isSaving} className="w-full">
        {isSaving ? t("actions.saving") : submitLabel}
      </Button>
    </form>
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
      <div className="flex flex-col gap-2 text-sm text-muted-foreground">
        <p>{teamCount > 0 ? t("teams.summaryCount", { count: teamCount }) : t("teams.emptyMetadata")}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
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

function buildSummary(orgUnit: OrgUnit, members: OrgUnitMember[]) {
  return {
    teams: getTeamCount(orgUnit),
    members: members.length || getMemberCount(orgUnit),
    admins: members.filter((member) => member.role === "org_admin").length || getAdminCount(orgUnit),
  };
}

function getOrgName(orgUnit: OrgUnit) {
  return orgUnit.display_name || toText(orgUnit.metadata?.display_name) || orgUnit.alias;
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

function getUsagePeriodRange(period: UsagePeriod, now = new Date()) {
  const to = new Date(now);
  if (period === "last_7_days") {
    return { from: new Date(to.getTime() - 7 * 24 * 60 * 60 * 1000), to };
  }
  if (period === "last_30_days") {
    return { from: new Date(to.getTime() - 30 * 24 * 60 * 60 * 1000), to };
  }
  return {
    ...getCurrentJalaliMonthRange(to),
  };
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

function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}
