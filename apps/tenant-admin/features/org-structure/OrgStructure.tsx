"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { AlertTriangle, Building2, Plus, Users } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { FilterCombobox } from "@/components/app-ui/FilterCombobox";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { SearchField } from "@/components/app-ui/SearchField";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { usePermissions } from "@/components/permissions-context";
import { canAnyPermission } from "@/lib/authz";
import { getCurrentJalaliMonthRange } from "@/lib/date/jalali-month";
import {
  addOrgUnitMember,
  createOrgUnit,
  deleteOrgUnit,
  getOrgUnit,
  getOrgUnitEntitlements,
  getOrgUnitUsage,
  listOrgUnitMembers,
  listOrgUnits,
  removeOrgUnitMember,
  updateOrgUnit,
  updateOrgUnitMember,
  type OrgUnit,
  type OrgUnitMember,
} from "@/lib/api/org-units";
import {
  addTeamMember,
  createTeam,
  deleteTeam,
  getTeam,
  getTeamUsage,
  listTeamCostBreakdown,
  listTeamMembers,
  listTeamUserCandidates,
  listTeams,
  removeTeamMember,
  updateTeam,
  updateTeamMember,
  type AccessPrincipalCandidate,
  type Team,
  type TeamMember,
} from "@/lib/api/teams";

import { buildOrgTree, TENANT_ROOT_ID, type OrgTreeRow } from "./org-tree";
import { OrgStructureTree, type OrgStructureTreeSelection } from "./OrgStructureTree";
import { OrgUnitDetail } from "./OrgUnitDetail";
import { TeamDetail } from "./TeamDetail";
import { formatMemberCandidateLabel, TeamMembersSheet } from "./TeamMembersSheet";
import { getOrgUnitName, getTeamName, getWorkspaceType, normalizeError, readBooleanFeature, toText, WORKSPACE_TYPES, type UsagePeriod } from "./helpers";

const EXPANDED_STORAGE_KEY = "org-structure:expanded";
const FEATURE_TEAMS_ENABLED = "teams.enabled";
const FEATURE_ORG_UNITS_ENABLED = "org_units.enabled";

type Banner = { kind: "success" | "error" | "warning"; message: string } | null;
type TeamForm = { alias: string; displayName: string; orgUnitID: string; workspaceType: string; description: string };
type OrgUnitForm = { alias: string; displayName: string; description: string };

const emptyTeamForm: TeamForm = { alias: "", displayName: "", orgUnitID: "", workspaceType: "production", description: "" };
const emptyOrgUnitForm: OrgUnitForm = { alias: "", displayName: "", description: "" };

export default function OrgStructure({
  initialOrgUnits,
  initialTeams,
  initialError,
  currentUserRole,
}: {
  initialOrgUnits: OrgUnit[];
  initialTeams: Team[];
  initialError: string | null;
  currentUserRole: string;
}) {
  const t = useTranslations("orgStructure");
  const tTeams = useTranslations("teams");
  const tOrgUnits = useTranslations("orgUnits");
  const locale = useLocale();
  const permissions = usePermissions() ?? [];
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [orgUnits, setOrgUnits] = useState(initialOrgUnits);
  const [teams, setTeams] = useState(initialTeams);
  const [isLoading, setIsLoading] = useState(false);
  const [query, setQuery] = useState("");
  const [expanded, setExpanded] = useState<Set<string>>(() => {
    if (typeof window === "undefined") return new Set();
    try {
      const raw = window.localStorage.getItem(EXPANDED_STORAGE_KEY);
      if (raw) return new Set(JSON.parse(raw) as string[]);
    } catch {
      // Ignore malformed/blocked storage; the tree just starts collapsed.
    }
    return new Set();
  });
  const [selection, setSelection] = useState<OrgStructureTreeSelection>(() => selectionFromSearchParams(searchParams));

  const [selectedOrgUnit, setSelectedOrgUnit] = useState<OrgUnit | null>(null);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [orgUnitMembers, setOrgUnitMembers] = useState<OrgUnitMember[]>([]);
  const [candidateUsers, setCandidateUsers] = useState<AccessPrincipalCandidate[]>([]);
  const [detailLoading, setDetailLoading] = useState(false);

  const [usagePeriod, setUsagePeriod] = useState<UsagePeriod>("current_month");
  const [selectionCostIRR, setSelectionCostIRR] = useState<number | null>(null);
  const [costByTeamID, setCostByTeamID] = useState<Map<string, number> | null>(null);

  const [banner, setBanner] = useState<Banner>(initialError ? { kind: "error", message: initialError } : null);
  const [overlayNotice, setOverlayNotice] = useState<Banner>(null);
  const [deleteNotice, setDeleteNotice] = useState<Banner>(null);
  const [isSaving, setIsSaving] = useState(false);

  const [entitlements, setEntitlements] = useState<Record<string, unknown> | null>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);

  const [createTeamOpen, setCreateTeamOpen] = useState(false);
  const [editTeamOpen, setEditTeamOpen] = useState(false);
  const [deleteTeamOpen, setDeleteTeamOpen] = useState(false);
  const [teamMembersOpen, setTeamMembersOpen] = useState(false);
  const [teamForm, setTeamForm] = useState<TeamForm>(emptyTeamForm);

  const [createUnitOpen, setCreateUnitOpen] = useState(false);
  const [editUnitOpen, setEditUnitOpen] = useState(false);
  const [deleteUnitOpen, setDeleteUnitOpen] = useState(false);
  const [unitMembersOpen, setUnitMembersOpen] = useState(false);
  const [unitForm, setUnitForm] = useState<OrgUnitForm>(emptyOrgUnitForm);
  const [unitMemberUserID, setUnitMemberUserID] = useState("");
  const [unitMemberRole, setUnitMemberRole] = useState("org_admin");

  const isTeamsFeatureEnabled = hasLoadedEntitlements ? readBooleanFeature(entitlements?.[FEATURE_TEAMS_ENABLED], false) : null;
  const isOrgUnitsFeatureEnabled = hasLoadedEntitlements ? readBooleanFeature(entitlements?.[FEATURE_ORG_UNITS_ENABLED], false) : null;

  const canManageTeam = ["tenant_admin", "proxy_admin", "org_admin", "team_admin"].includes(currentUserRole) || canAnyPermission(permissions, ["team:manage"]);
  const canCreateTeam = canManageTeam || canAnyPermission(permissions, ["team:create"]);
  const canDeleteTeam = canAnyPermission(permissions, ["team:delete"]);
  const canManageOrg = ["tenant_admin", "proxy_admin", "org_admin"].includes(currentUserRole) || canAnyPermission(permissions, ["org_unit:manage"]);
  const canCreateOrgUnit = canManageOrg || canAnyPermission(permissions, ["org_unit:create"]);
  const canDeleteOrgUnit = canManageOrg || canAnyPermission(permissions, ["org_unit:delete"]);

  // Load entitlements once.
  useEffect(() => {
    let isMounted = true;
    void getOrgUnitEntitlements()
      .then((response) => { if (isMounted) setEntitlements(response?.features ?? null); })
      .catch(() => { if (isMounted) setEntitlements(null); })
      .finally(() => { if (isMounted) setHasLoadedEntitlements(true); });
    return () => { isMounted = false; };
  }, []);

  function toggleExpand(id: string) {
    setExpanded((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      try {
        window.localStorage.setItem(EXPANDED_STORAGE_KEY, JSON.stringify([...next]));
      } catch {
        // Best effort only.
      }
      return next;
    });
  }

  function updateSelection(next: OrgStructureTreeSelection) {
    setSelection(next);
    setSelectionCostIRR(null);
    if (!next) {
      setSelectedOrgUnit(null);
      setSelectedTeam(null);
      setTeamMembers([]);
      setOrgUnitMembers([]);
    }
    const params = new URLSearchParams(searchParams.toString());
    if (next) params.set("sel", `${next.kind}:${next.id}`);
    else params.delete("sel");
    const nextQuery = params.toString();
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
  }

  const rows = useMemo<OrgTreeRow[]>(() => {
    if (isOrgUnitsFeatureEnabled === false) {
      return [...teams]
        .sort((a, b) => getTeamName(a).localeCompare(getTeamName(b)))
        .map((team) => ({ kind: "team" as const, id: team.id, team, unitId: null, unitName: null }));
    }
    const effectiveTeams = isTeamsFeatureEnabled === false ? [] : teams;
    return buildOrgTree(orgUnits, effectiveTeams, expanded, query);
  }, [orgUnits, teams, expanded, query, isOrgUnitsFeatureEnabled, isTeamsFeatureEnabled]);

  async function loadOrgUnits() {
    const search = new URLSearchParams({ limit: "200", offset: "0" });
    const payload = await listOrgUnits(search);
    setOrgUnits(payload.data ?? []);
  }

  async function loadTeams() {
    const search = new URLSearchParams({ limit: "200", offset: "0" });
    const payload = await listTeams(search);
    setTeams(payload.data ?? []);
  }

  async function refresh() {
    setIsLoading(true);
    try {
      await Promise.all([loadOrgUnits(), loadTeams()]);
      setBanner(null);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, tTeams("messages.loadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadUsers() {
    try {
      const payload = await listTeamUserCandidates();
      setCandidateUsers(payload.data ?? []);
    } catch {
      setCandidateUsers([]);
    }
  }

  // Load the selected unit or team's detail (fresh record + members).
  useEffect(() => {
    if (!selection) return;
    let isMounted = true;

    async function run() {
      setDetailLoading(true);
      try {
        if (selection!.kind === "team") {
          const [team, memberPayload] = await Promise.all([getTeam(selection!.id), listTeamMembers(selection!.id)]);
          if (!isMounted) return;
          setSelectedTeam(team);
          setTeamMembers(memberPayload.data ?? []);
          setSelectedOrgUnit(null);
          setTeams((prev) => prev.map((item) => (item.id === team.id ? team : item)));
        } else if (selection!.id !== TENANT_ROOT_ID) {
          const [orgUnit, memberPayload] = await Promise.all([getOrgUnit(selection!.id), listOrgUnitMembers(selection!.id)]);
          if (!isMounted) return;
          setSelectedOrgUnit(orgUnit);
          setOrgUnitMembers(memberPayload.data ?? []);
          setSelectedTeam(null);
          setOrgUnits((prev) => prev.map((item) => (item.id === orgUnit.id ? orgUnit : item)));
        } else {
          setSelectedOrgUnit(null);
          setSelectedTeam(null);
        }
        setBanner(null);
      } catch (error) {
        if (isMounted) setBanner({ kind: "error", message: normalizeError(error, tTeams("messages.detailLoadError")) });
      } finally {
        if (isMounted) setDetailLoading(false);
      }
    }

    void run();
    return () => { isMounted = false; };
  }, [selection, tTeams]);

  // Usage for the selected row (single number, existing /api/logs/summary proxy).
  useEffect(() => {
    if (!selection || selection.id === TENANT_ROOT_ID) return;
    const controller = new AbortController();
    const range = getUsagePeriodRange(usagePeriod);
    const params = new URLSearchParams({ from: range.from.toISOString(), to: range.to.toISOString() });
    if (selection.kind === "team") params.set("team_id", selection.id);
    else params.set("org_unit_id", selection.id);

    async function run() {
      try {
        const payload = selection!.kind === "team" ? await getTeamUsage(params, controller.signal) : await getOrgUnitUsage(params, controller.signal);
        setSelectionCostIRR(payload.data?.cost_irr ?? 0);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setSelectionCostIRR(null);
      }
    }

    void run();
    return () => controller.abort();
  }, [selection, usagePeriod]);

  // Best-effort per-team cost breakdown for the tree's "cost this month" column.
  useEffect(() => {
    const controller = new AbortController();
    const range = getUsagePeriodRange(usagePeriod);
    const params = new URLSearchParams({ type: "cost", group_by: "team_id", from: range.from.toISOString(), to: range.to.toISOString() });

    async function run() {
      try {
        const payload = await listTeamCostBreakdown(params, controller.signal);
        const entries = payload.data ?? [];
        if (entries.length === 0) { setCostByTeamID(null); return; }
        const next = new Map<string, number>();
        for (const entry of entries) {
          if (entry.team_id) next.set(entry.team_id, entry.cost_irr ?? 0);
        }
        setCostByTeamID(next.size > 0 ? next : null);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setCostByTeamID(null);
      }
    }

    void run();
    return () => controller.abort();
  }, [usagePeriod]);

  const memberCandidateOptions = useMemo(() => {
    const memberIDs = new Set(teamMembers.map((member) => member.user_id));
    return candidateUsers.filter((user) => !memberIDs.has(user.principal_id)).map((user) => ({ userID: user.principal_id, label: formatMemberCandidateLabel(user) }));
  }, [teamMembers, candidateUsers]);
  const memberLabelByUserID = useMemo(() => new Map(candidateUsers.map((user) => [user.principal_id, formatMemberCandidateLabel(user)])), [candidateUsers]);
  const orgUnitForSelectedTeam = useMemo(() => {
    if (!selectedTeam?.org_unit_id) return null;
    return orgUnits.find((unit) => unit.id === selectedTeam.org_unit_id) ?? null;
  }, [selectedTeam, orgUnits]);
  const teamsForSelectedUnit = useMemo(() => {
    if (!selectedOrgUnit) return [];
    return teams.filter((team) => team.org_unit_id === selectedOrgUnit.id);
  }, [selectedOrgUnit, teams]);

  function handleCreateTeam() {
    if (!canCreateTeam) return;
    if (isTeamsFeatureEnabled === false) { setUpgradeDialogOpen(true); return; }
    setTeamForm(emptyTeamForm);
    setCreateTeamOpen(true);
  }

  function handleCreateOrgUnit() {
    if (!canCreateOrgUnit) {
      showFeedbackToast({ kind: "error", title: tOrgUnits("alerts.error"), description: tOrgUnits("messages.permissionDenied") });
      return;
    }
    if (isOrgUnitsFeatureEnabled === false) { setUpgradeDialogOpen(true); return; }
    setUnitForm(emptyOrgUnitForm);
    setCreateUnitOpen(true);
  }

  async function submitCreateTeam(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateAlias(teamForm.alias, teams.map((team) => team.alias), tTeams);
    if (validation) return setOverlayNotice({ kind: "error", message: validation });
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      const created = await createTeam(buildTeamPayload(teamForm));
      setCreateTeamOpen(false);
      showFeedbackToast({ kind: "success", title: tTeams("alerts.success"), description: tTeams("messages.createSuccess") });
      await loadTeams();
      updateSelection({ kind: "team", id: created.id });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tTeams("messages.createError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function submitEditTeam(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedTeam) return;
    const validation = validateAlias(teamForm.alias, teams.filter((team) => team.id !== selectedTeam.id).map((team) => team.alias), tTeams);
    if (validation) return setOverlayNotice({ kind: "error", message: validation });
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      const updated = await updateTeam(selectedTeam.id, buildTeamPayload(teamForm, selectedTeam.metadata));
      setSelectedTeam(updated);
      setTeams((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      setEditTeamOpen(false);
      showFeedbackToast({ kind: "success", title: tTeams("alerts.success"), description: tTeams("messages.updateSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tTeams("messages.updateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function handleDeleteTeam() {
    if (!selectedTeam) return;
    setIsSaving(true);
    try {
      await deleteTeam(selectedTeam.id);
      showFeedbackToast({ kind: "success", title: tTeams("alerts.success"), description: tTeams("messages.deleteSuccess") });
      updateSelection(null);
      await loadTeams();
    } catch (error) {
      setDeleteNotice({ kind: "error", message: normalizeError(error, tTeams("messages.deleteError")) });
      throw error;
    } finally {
      setIsSaving(false);
    }
  }

  async function addMember(userID: string, role: string, memberPermissions: string[]) {
    if (!selectedTeam) return;
    setIsSaving(true);
    try {
      await addTeamMember(selectedTeam.id, { user_id: userID, role, team_member_permissions: memberPermissions, metadata: {} });
      const memberPayload = await listTeamMembers(selectedTeam.id);
      setTeamMembers(memberPayload.data ?? []);
      showFeedbackToast({ kind: "success", title: tTeams("alerts.success"), description: tTeams("messages.memberAddSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tTeams("messages.memberAddError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function updateMember(member: TeamMember, role: string, memberPermissions: string[]) {
    if (!selectedTeam) return;
    setIsSaving(true);
    try {
      await updateTeamMember(selectedTeam.id, member.user_id, { role, team_member_permissions: memberPermissions });
      const memberPayload = await listTeamMembers(selectedTeam.id);
      setTeamMembers(memberPayload.data ?? []);
      showFeedbackToast({ kind: "success", title: tTeams("alerts.success"), description: tTeams("messages.memberUpdateSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tTeams("messages.memberUpdateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function removeMember(userID: string) {
    if (!selectedTeam) return;
    setIsSaving(true);
    try {
      await removeTeamMember(selectedTeam.id, userID);
      const memberPayload = await listTeamMembers(selectedTeam.id);
      setTeamMembers(memberPayload.data ?? []);
      showFeedbackToast({ kind: "success", title: tTeams("alerts.success"), description: tTeams("messages.memberRemoveSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tTeams("messages.memberRemoveError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function submitCreateOrgUnit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateAlias(unitForm.alias, orgUnits.map((unit) => unit.alias), tOrgUnits);
    if (validation) return setOverlayNotice({ kind: "error", message: validation });
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      const created = await createOrgUnit(buildOrgUnitPayload(unitForm));
      setCreateUnitOpen(false);
      showFeedbackToast({ kind: "success", title: tOrgUnits("alerts.success"), description: tOrgUnits("messages.createSuccess") });
      await loadOrgUnits();
      updateSelection({ kind: "unit", id: created.id });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tOrgUnits("messages.createError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function submitEditOrgUnit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedOrgUnit) return;
    const validation = validateAlias(unitForm.alias, orgUnits.filter((unit) => unit.id !== selectedOrgUnit.id).map((unit) => unit.alias), tOrgUnits);
    if (validation) return setOverlayNotice({ kind: "error", message: validation });
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      const updated = await updateOrgUnit(selectedOrgUnit.id, buildOrgUnitPayload(unitForm, selectedOrgUnit.metadata));
      setSelectedOrgUnit(updated);
      setOrgUnits((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      setEditUnitOpen(false);
      showFeedbackToast({ kind: "success", title: tOrgUnits("alerts.success"), description: tOrgUnits("messages.updateSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tOrgUnits("messages.updateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function handleDeleteOrgUnit() {
    if (!selectedOrgUnit) return;
    setIsSaving(true);
    try {
      await deleteOrgUnit(selectedOrgUnit.id);
      showFeedbackToast({ kind: "success", title: tOrgUnits("alerts.success"), description: tOrgUnits("messages.deleteSuccess") });
      updateSelection(null);
      await loadOrgUnits();
    } catch (error) {
      setDeleteNotice({ kind: "error", message: normalizeError(error, tOrgUnits("messages.deleteError")) });
      throw error;
    } finally {
      setIsSaving(false);
    }
  }

  async function submitAddUnitMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedOrgUnit || !unitMemberUserID.trim()) return;
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      await addOrgUnitMember(selectedOrgUnit.id, unitMemberUserID.trim(), unitMemberRole);
      const memberPayload = await listOrgUnitMembers(selectedOrgUnit.id);
      setOrgUnitMembers(memberPayload.data ?? []);
      setUnitMemberUserID("");
      showFeedbackToast({ kind: "success", title: tOrgUnits("alerts.success"), description: tOrgUnits("messages.memberAddSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tOrgUnits("messages.memberAddError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function updateUnitMemberRole(userID: string, role: string) {
    if (!selectedOrgUnit) return;
    setIsSaving(true);
    try {
      await updateOrgUnitMember(selectedOrgUnit.id, userID, role);
      const memberPayload = await listOrgUnitMembers(selectedOrgUnit.id);
      setOrgUnitMembers(memberPayload.data ?? []);
      showFeedbackToast({ kind: "success", title: tOrgUnits("alerts.success"), description: tOrgUnits("messages.memberUpdateSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tOrgUnits("messages.memberUpdateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function removeUnitMember(userID: string) {
    if (!selectedOrgUnit) return;
    setIsSaving(true);
    try {
      await removeOrgUnitMember(selectedOrgUnit.id, userID);
      const memberPayload = await listOrgUnitMembers(selectedOrgUnit.id);
      setOrgUnitMembers(memberPayload.data ?? []);
      showFeedbackToast({ kind: "success", title: tOrgUnits("alerts.success"), description: tOrgUnits("messages.memberRemoveSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, tOrgUnits("messages.memberRemoveError")) });
    } finally {
      setIsSaving(false);
    }
  }

  const bothDisabled = hasLoadedEntitlements && isTeamsFeatureEnabled === false && isOrgUnitsFeatureEnabled === false;

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <OperationNotice notice={banner} labels={{ success: tTeams("alerts.success"), error: tTeams("alerts.error"), warning: tTeams("alerts.warning"), info: tTeams("alerts.warning") }} />

      {bothDisabled ? (
        <EmptyState
          icon={Building2}
          title={t("entitlements.bothDisabledTitle")}
          description={t("entitlements.bothDisabledDescription")}
          primaryAction={<Button type="button" onClick={() => setUpgradeDialogOpen(true)}>{t("entitlements.upgrade")}</Button>}
        />
      ) : (
        <section className="grid min-w-0 gap-6 2xl:grid-cols-[minmax(0,1.2fr)_minmax(22rem,0.8fr)]">
          <AppCard className="min-w-0 shadow-none" contentClassName="flex min-w-0 flex-col gap-4">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
              <SearchField containerClassName="w-full sm:min-w-64 sm:flex-1" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("search.placeholder")} />
              <div className="flex flex-wrap gap-2 sm:shrink-0 sm:flex-nowrap">
                {isOrgUnitsFeatureEnabled !== false ? (
                  <Button type="button" variant="outline" onClick={handleCreateOrgUnit} disabled={!canCreateOrgUnit || !hasLoadedEntitlements} className="flex-1 sm:flex-none">
                    <Plus data-icon="inline-start" />
                    {t("actions.newOrgUnit")}
                  </Button>
                ) : null}
                <Button type="button" onClick={handleCreateTeam} disabled={!canCreateTeam || !hasLoadedEntitlements} className="flex-1 sm:flex-none">
                  <Plus data-icon="inline-start" />
                  {t("actions.newTeam")}
                </Button>
                <AppSelectField
                  ariaLabel={t("period.label")}
                  value={usagePeriod}
                  onValueChange={(value) => setUsagePeriod(value as UsagePeriod)}
                  triggerClassName="h-9"
                  options={[
                    { value: "current_month", label: t("period.current_month") },
                    { value: "last_30_days", label: t("period.last_30_days") },
                    { value: "last_7_days", label: t("period.last_7_days") },
                  ]}
                />
                <RefreshButton label={t("actions.refresh")} onClick={() => void refresh()} isLoading={isLoading} />
              </div>
            </div>

            <OrgStructureTree
              rows={rows}
              isLoading={isLoading && orgUnits.length === 0 && teams.length === 0}
              isSearching={query.trim() !== ""}
              searchQuery={query.trim()}
              selection={selection}
              onSelectUnit={(unitID) => updateSelection({ kind: "unit", id: unitID })}
              onSelectTeam={(teamID) => updateSelection({ kind: "team", id: teamID })}
              onToggleExpand={toggleExpand}
              costByTeamID={costByTeamID}
              locale={locale}
            />
          </AppCard>

          <div className="flex min-w-0 flex-col gap-4">
            {!selection ? (
              <EmptyState icon={Users} title={t("detail.empty")} className="min-h-56" />
            ) : selection.kind === "team" && selectedTeam ? (
              <TeamDetail
                team={selectedTeam}
                orgUnit={orgUnitForSelectedTeam}
                members={teamMembers}
                isLoading={detailLoading}
                costIRR={selectionCostIRR}
                locale={locale}
                canManage={canManageTeam}
                canDelete={canDeleteTeam}
                onEdit={() => { setTeamForm(formFromTeam(selectedTeam)); setEditTeamOpen(true); }}
                onManageMembers={() => { setTeamMembersOpen(true); void loadUsers(); }}
                onDelete={() => setDeleteTeamOpen(true)}
              />
            ) : selection.kind === "unit" && selectedOrgUnit ? (
              <OrgUnitDetail
                orgUnit={selectedOrgUnit}
                teams={teamsForSelectedUnit}
                members={orgUnitMembers}
                isLoading={detailLoading}
                costIRR={selectionCostIRR}
                locale={locale}
                canManage={canManageOrg}
                canDelete={canDeleteOrgUnit}
                onEdit={() => { setUnitForm(formFromOrgUnit(selectedOrgUnit)); setEditUnitOpen(true); }}
                onManageMembers={() => { setUnitMembersOpen(true); void loadUsers(); }}
                onSelectTeam={(teamID) => updateSelection({ kind: "team", id: teamID })}
                onDelete={() => setDeleteUnitOpen(true)}
              />
            ) : (
              <EmptyState icon={Users} title={t("detail.empty")} className="min-h-56" />
            )}
          </div>
        </section>
      )}

      <FormSheet
        open={createTeamOpen}
        onOpenChange={(open) => { setCreateTeamOpen(open); setOverlayNotice(null); }}
        title={tTeams("create.title")}
        description={tTeams("create.description")}
        pending={isSaving}
        closeLabel={t("actions.cancel")}
        footer={<FormSheetActions secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => setCreateTeamOpen(false)}>{t("actions.cancel")}</Button>} primary={<Button type="submit" form="org-structure-create-team-form" disabled={isSaving}>{isSaving ? t("actions.saving") : tTeams("create.submit")}</Button>} />}
      >
        <OperationNotice notice={overlayNotice} labels={{ success: tTeams("alerts.success"), error: tTeams("alerts.error"), warning: tTeams("alerts.warning"), info: tTeams("alerts.warning") }} />
        <TeamFormFields formID="org-structure-create-team-form" form={teamForm} setForm={setTeamForm} orgUnits={orgUnits} onSubmit={submitCreateTeam} />
      </FormSheet>

      <FormSheet
        open={editTeamOpen}
        onOpenChange={(open) => { setEditTeamOpen(open); setOverlayNotice(null); }}
        title={tTeams("edit.title")}
        description={tTeams("edit.description")}
        pending={isSaving}
        closeLabel={t("actions.cancel")}
        footer={<FormSheetActions secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => setEditTeamOpen(false)}>{t("actions.cancel")}</Button>} primary={<Button type="submit" form="org-structure-edit-team-form" disabled={isSaving}>{isSaving ? t("actions.saving") : tTeams("edit.submit")}</Button>} />}
      >
        <OperationNotice notice={overlayNotice} labels={{ success: tTeams("alerts.success"), error: tTeams("alerts.error"), warning: tTeams("alerts.warning"), info: tTeams("alerts.warning") }} />
        <TeamFormFields formID="org-structure-edit-team-form" form={teamForm} setForm={setTeamForm} orgUnits={orgUnits} onSubmit={submitEditTeam} />
      </FormSheet>

      <TeamMembersSheet
        open={teamMembersOpen}
        onOpenChange={(open) => { setTeamMembersOpen(open); setOverlayNotice(null); }}
        members={teamMembers}
        candidateOptions={memberCandidateOptions}
        memberLabelByUserID={memberLabelByUserID}
        canManage={canManageTeam}
        isSaving={isSaving}
        notice={overlayNotice}
        onAddMember={addMember}
        onUpdateMember={updateMember}
        onRemoveMember={removeMember}
      />

      {selectedTeam ? (
        <TypedConfirmDialog
          open={deleteTeamOpen}
          onOpenChange={(open) => { setDeleteTeamOpen(open); if (!open) setDeleteNotice(null); }}
          title={tTeams("delete.title")}
          description={tTeams("delete.description")}
          expectedValue={selectedTeam.alias}
          inputLabel={tTeams("delete.confirm", { alias: selectedTeam.alias })}
          cancelLabel={t("actions.cancel")}
          confirmLabel={tTeams("delete.submit")}
          pendingLabel={t("actions.saving")}
          onConfirm={handleDeleteTeam}
          feedback={<OperationNotice notice={deleteNotice} labels={{ success: tTeams("alerts.success"), error: tTeams("alerts.error"), warning: tTeams("alerts.warning"), info: tTeams("alerts.warning") }} />}
          disabled={!canDeleteTeam || getActiveKeysOrMembers(selectedTeam, teamMembers)}
          warning={<Alert variant="warning"><AlertTriangle aria-hidden="true" /><AlertTitle>{tTeams("delete.guardTitle")}</AlertTitle><AlertDescription>{tTeams("delete.guardDescription")}</AlertDescription></Alert>}
        />
      ) : null}

      <FormSheet
        open={createUnitOpen}
        onOpenChange={(open) => { setCreateUnitOpen(open); setOverlayNotice(null); }}
        title={tOrgUnits("create.title")}
        description={tOrgUnits("create.description")}
        size="md"
        pending={isSaving}
        closeLabel={t("actions.cancel")}
        footer={<FormSheetActions secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => setCreateUnitOpen(false)}>{t("actions.cancel")}</Button>} primary={<Button type="submit" form="org-structure-create-unit-form" disabled={isSaving}>{isSaving ? t("actions.saving") : tOrgUnits("create.submit")}</Button>} />}
      >
        <OperationNotice notice={overlayNotice} labels={{ success: tOrgUnits("alerts.success"), error: tOrgUnits("alerts.error"), warning: tOrgUnits("alerts.warning"), info: tOrgUnits("alerts.warning") }} />
        <OrgUnitFormFields formID="org-structure-create-unit-form" form={unitForm} setForm={setUnitForm} onSubmit={submitCreateOrgUnit} />
      </FormSheet>

      <FormSheet
        open={editUnitOpen}
        onOpenChange={(open) => { setEditUnitOpen(open); setOverlayNotice(null); }}
        title={tOrgUnits("edit.title")}
        description={tOrgUnits("edit.description")}
        size="md"
        pending={isSaving}
        closeLabel={t("actions.cancel")}
        footer={<FormSheetActions secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => setEditUnitOpen(false)}>{t("actions.cancel")}</Button>} primary={<Button type="submit" form="org-structure-edit-unit-form" disabled={isSaving}>{isSaving ? t("actions.saving") : tOrgUnits("edit.submit")}</Button>} />}
      >
        <OperationNotice notice={overlayNotice} labels={{ success: tOrgUnits("alerts.success"), error: tOrgUnits("alerts.error"), warning: tOrgUnits("alerts.warning"), info: tOrgUnits("alerts.warning") }} />
        <OrgUnitFormFields formID="org-structure-edit-unit-form" form={unitForm} setForm={setUnitForm} onSubmit={submitEditOrgUnit} />
      </FormSheet>

      <FormSheet
        open={unitMembersOpen}
        onOpenChange={(open) => { setUnitMembersOpen(open); setOverlayNotice(null); }}
        title={tOrgUnits("members.manageTitle")}
        description={tOrgUnits("members.manageDescription")}
        size="md"
        pending={isSaving}
        closeLabel={t("actions.cancel")}
        footer={<FormSheetActions secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => setUnitMembersOpen(false)}>{t("actions.cancel")}</Button>} primary={<Button type="submit" form="org-structure-add-unit-member-form" disabled={isSaving || !canManageOrg}>{tOrgUnits("members.add")}</Button>} />}
      >
        <OperationNotice notice={overlayNotice} labels={{ success: tOrgUnits("alerts.success"), error: tOrgUnits("alerts.error"), warning: tOrgUnits("alerts.warning"), info: tOrgUnits("alerts.warning") }} />
        <form id="org-structure-add-unit-member-form" className="flex flex-col gap-4" onSubmit={submitAddUnitMember}>
          <div className="flex flex-col gap-2">
            <Label htmlFor="org-structure-unit-member-user">{tOrgUnits("members.user")}</Label>
            <FilterCombobox
              id="org-structure-unit-member-user"
              options={candidateUsers.slice(0, 100).map((user) => ({ value: user.principal_id, label: formatMemberCandidateLabel(user) }))}
              value={unitMemberUserID}
              onValueChange={setUnitMemberUserID}
              placeholder={tOrgUnits("members.userPlaceholder")}
              emptyText={tOrgUnits("common.notSet")}
              className="h-10"
            />
          </div>
          <AppSelectField
            label={tOrgUnits("members.role")}
            value={unitMemberRole}
            onValueChange={setUnitMemberRole}
            triggerClassName="h-10"
            options={[{ value: "org_admin", label: tOrgUnits("roles.orgAdmin") }, { value: "member", label: tOrgUnits("roles.member") }]}
          />
        </form>
        <div className="mt-6 flex flex-col gap-3">
          {orgUnitMembers.map((member) => (
            <div key={member.id} className="rounded-lg border p-3">
              <div className="flex items-start justify-between gap-3">
                <span className="truncate text-sm font-medium" dir="auto">{member.user_id}</span>
                <Button size="sm" variant="outline" onClick={() => void updateUnitMemberRole(member.user_id, member.role === "org_admin" ? "member" : "org_admin")} disabled={isSaving || !canManageOrg}>
                  {member.role === "org_admin" ? tOrgUnits("members.makeMember") : tOrgUnits("members.makeOrgAdmin")}
                </Button>
              </div>
              <Button size="sm" variant="outline" className="mt-2 text-destructive" onClick={() => void removeUnitMember(member.user_id)} disabled={isSaving || !canManageOrg}>
                {tOrgUnits("members.remove")}
              </Button>
            </div>
          ))}
        </div>
      </FormSheet>

      {selectedOrgUnit ? (
        <TypedConfirmDialog
          open={deleteUnitOpen}
          onOpenChange={(open) => { setDeleteUnitOpen(open); if (!open) setDeleteNotice(null); }}
          title={tOrgUnits("delete.title")}
          description={tOrgUnits("delete.description")}
          expectedValue={selectedOrgUnit.alias}
          inputLabel={tOrgUnits("delete.confirmLabel", { alias: selectedOrgUnit.alias })}
          cancelLabel={t("actions.cancel")}
          confirmLabel={tOrgUnits("delete.submit")}
          pendingLabel={t("actions.saving")}
          onConfirm={handleDeleteOrgUnit}
          feedback={<OperationNotice notice={deleteNotice} labels={{ success: tOrgUnits("alerts.success"), error: tOrgUnits("alerts.error"), warning: tOrgUnits("alerts.warning"), info: tOrgUnits("alerts.warning") }} />}
          disabled={!canDeleteOrgUnit || teamsForSelectedUnit.length > 0}
          warning={teamsForSelectedUnit.length > 0 ? <Alert variant="warning"><AlertTriangle aria-hidden="true" /><AlertTitle>{tOrgUnits("delete.childTeamsTitle")}</AlertTitle><AlertDescription>{tOrgUnits("delete.childTeamsDescription", { name: getOrgUnitName(selectedOrgUnit) })}</AlertDescription></Alert> : undefined}
        />
      ) : null}

      <CommercialPlanUpgradeDialog
        open={upgradeDialogOpen}
        onOpenChange={setUpgradeDialogOpen}
        title={bothDisabled ? t("entitlements.bothDisabledTitle") : tTeams("upgradeDialog.title")}
        description={bothDisabled ? t("entitlements.bothDisabledDescription") : tTeams("upgradeDialog.description")}
        cancelLabel={tTeams("upgradeDialog.cancel")}
        actionLabel={tTeams("upgradeDialog.action")}
      />
    </div>
  );
}

function TeamFormFields({ formID, form, setForm, orgUnits, onSubmit }: { formID: string; form: TeamForm; setForm: (value: TeamForm) => void; orgUnits: OrgUnit[]; onSubmit: (event: FormEvent<HTMLFormElement>) => void }) {
  const tTeams = useTranslations("teams");
  const update = <K extends keyof TeamForm>(key: K, value: TeamForm[K]) => setForm({ ...form, [key]: value });
  return (
    <form id={formID} className="flex flex-col gap-4" onSubmit={onSubmit}>
      <FieldGroup>
        <DisplayNameField id="org-structure-team-display" label={tTeams("form.displayName")} value={form.displayName} onChange={(event) => update("displayName", event.target.value)} />
        <IdentifierField id="org-structure-team-alias" label={tTeams("form.alias")} value={form.alias} onChange={(event) => update("alias", event.target.value)} required />
        <AppSelectField label={tTeams("form.orgUnit")} value={form.orgUnitID || "root"} onValueChange={(value) => update("orgUnitID", value === "root" ? "" : value)} triggerClassName="h-10" options={[{ value: "root", label: tTeams("table.tenantRoot") }, ...orgUnits.map((org) => ({ value: org.id, label: getOrgUnitName(org) }))]} />
        <AppSelectField label={tTeams("form.workspaceType")} value={form.workspaceType} onValueChange={(value) => update("workspaceType", value)} triggerClassName="h-10" options={WORKSPACE_TYPES.map((type) => ({ value: type, label: tTeams(`workspace.${type}`) }))} />
        <Field><FieldLabel htmlFor="org-structure-team-description">{tTeams("form.description")}</FieldLabel><Input id="org-structure-team-description" value={form.description} onChange={(event) => update("description", event.target.value)} /></Field>
      </FieldGroup>
    </form>
  );
}

function OrgUnitFormFields({ formID, form, setForm, onSubmit }: { formID: string; form: OrgUnitForm; setForm: (value: OrgUnitForm) => void; onSubmit: (event: FormEvent<HTMLFormElement>) => void }) {
  const tOrgUnits = useTranslations("orgUnits");
  const update = <K extends keyof OrgUnitForm>(key: K, value: OrgUnitForm[K]) => setForm({ ...form, [key]: value });
  return (
    <form id={formID} className="flex flex-col gap-5" onSubmit={onSubmit}>
      <IdentifierField id="org-structure-unit-alias" label={tOrgUnits("form.alias")} value={form.alias} onChange={(event) => update("alias", event.target.value)} placeholder={tOrgUnits("form.aliasPlaceholder")} required />
      <DisplayNameField id="org-structure-unit-display-name" label={tOrgUnits("form.displayName")} value={form.displayName} onChange={(event) => update("displayName", event.target.value)} placeholder={tOrgUnits("form.displayNamePlaceholder")} />
      <AppTextField id="org-structure-unit-description" label={tOrgUnits("form.description")} value={form.description} onChange={(event) => update("description", event.target.value)} placeholder={tOrgUnits("form.descriptionPlaceholder")} />
    </form>
  );
}

function selectionFromSearchParams(searchParams: Pick<URLSearchParams, "get">): OrgStructureTreeSelection {
  const raw = searchParams.get("sel");
  if (!raw) return null;
  const [kind, ...rest] = raw.split(":");
  const id = rest.join(":");
  if (!id) return null;
  if (kind === "team") return { kind: "team", id };
  if (kind === "unit") return { kind: "unit", id };
  return null;
}

function buildTeamPayload(form: TeamForm, metadata?: Record<string, unknown>) {
  return {
    alias: form.alias.trim(),
    display_name: form.displayName.trim(),
    org_unit_id: form.orgUnitID.trim() || null,
    workspace_type: form.workspaceType,
    metadata: { ...withoutLegacyTeamMetadata(metadata), description: form.description.trim() },
  };
}

function withoutLegacyTeamMetadata(metadata?: Record<string, unknown>) {
  return Object.fromEntries(Object.entries(metadata ?? {}).filter(([key]) => key !== "allowed_models" && key !== "cost_center"));
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

function buildOrgUnitPayload(form: OrgUnitForm, previousMetadata?: Record<string, unknown>) {
  const metadataWithoutCostCenter = { ...(previousMetadata ?? {}) };
  delete metadataWithoutCostCenter.cost_center;
  delete metadataWithoutCostCenter.costCenter;
  return {
    alias: form.alias.trim(),
    display_name: form.displayName.trim(),
    metadata: { ...metadataWithoutCostCenter, display_name: form.displayName.trim() || form.alias.trim(), description: form.description.trim() },
  };
}

function formFromOrgUnit(orgUnit: OrgUnit): OrgUnitForm {
  return { alias: orgUnit.alias, displayName: getOrgUnitName(orgUnit), description: toText(orgUnit.metadata?.description) };
}

function validateAlias(alias: string, existingAliases: string[], t: ReturnType<typeof useTranslations>) {
  const trimmed = alias.trim();
  if (!trimmed) return t("messages.aliasRequired");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{1,62}$/.test(trimmed)) return t("messages.aliasPattern");
  if (existingAliases.some((existing) => existing.toLowerCase() === trimmed.toLowerCase())) return t("messages.aliasUnique");
  return null;
}

function getActiveKeysOrMembers(team: Team, members: TeamMember[]) {
  const activeKeys = Number(team.metadata?.active_keys_count ?? team.metadata?.active_key_count ?? 0);
  const virtualUsers = Number(team.metadata?.virtual_users_count ?? 0);
  return activeKeys > 0 || members.length > 0 || virtualUsers > 0;
}

function getUsagePeriodRange(period: UsagePeriod, now = new Date()) {
  const to = new Date(now);
  if (period === "last_7_days") return { from: new Date(to.getTime() - 7 * 24 * 60 * 60 * 1000), to };
  if (period === "last_30_days") return { from: new Date(to.getTime() - 30 * 24 * 60 * 60 * 1000), to };
  return getCurrentJalaliMonthRange(to);
}
