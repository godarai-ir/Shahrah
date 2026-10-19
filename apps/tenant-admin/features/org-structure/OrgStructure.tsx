"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { AlertTriangle, Plus, Users } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { Input } from "@/components/ui/input";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { SearchField } from "@/components/app-ui/SearchField";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { usePermissions } from "@/components/permissions-context";
import { canAnyPermission } from "@/lib/authz";
import { getCurrentJalaliMonthRange } from "@/lib/date/jalali-month";
import {
  addTeamMember,
  createTeam,
  deleteTeam,
  getTeam,
  getTeamEntitlements,
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

import { OrgStructureTree, type OrgStructureTreeSelection } from "./OrgStructureTree";
import { TeamDetail } from "./TeamDetail";
import { formatMemberCandidateLabel, TeamMembersSheet } from "./TeamMembersSheet";
import { getTeamName, getWorkspaceType, normalizeError, readBooleanFeature, toText, WORKSPACE_TYPES, type UsagePeriod } from "./helpers";

const FEATURE_TEAMS_ENABLED = "teams.enabled";

type Banner = { kind: "success" | "error" | "warning"; message: string } | null;
type TeamForm = { alias: string; displayName: string; workspaceType: string; description: string };

const emptyTeamForm: TeamForm = { alias: "", displayName: "", workspaceType: "production", description: "" };

export default function OrgStructure({
  initialTeams,
  initialError,
  currentUserRole,
}: {
  initialTeams: Team[];
  initialError: string | null;
  currentUserRole: string;
}) {
  const t = useTranslations("orgStructure");
  const tTeams = useTranslations("teams");
  const locale = useLocale();
  const permissions = usePermissions() ?? [];
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [teams, setTeams] = useState(initialTeams);
  const [isLoading, setIsLoading] = useState(false);
  const [query, setQuery] = useState("");
  const [selection, setSelection] = useState<OrgStructureTreeSelection>(() => selectionFromSearchParams(searchParams));

  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
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

  const isTeamsFeatureEnabled = hasLoadedEntitlements ? readBooleanFeature(entitlements?.[FEATURE_TEAMS_ENABLED], false) : null;

  const canManageTeam = ["tenant_admin", "proxy_admin", "org_admin", "team_admin"].includes(currentUserRole) || canAnyPermission(permissions, ["team:manage"]);
  const canCreateTeam = canManageTeam || canAnyPermission(permissions, ["team:create"]);
  const canDeleteTeam = canAnyPermission(permissions, ["team:delete"]);

  // Load entitlements once.
  useEffect(() => {
    let isMounted = true;
    void getTeamEntitlements()
      .then((response) => { if (isMounted) setEntitlements(response?.features ?? null); })
      .catch(() => { if (isMounted) setEntitlements(null); })
      .finally(() => { if (isMounted) setHasLoadedEntitlements(true); });
    return () => { isMounted = false; };
  }, []);

  function updateSelection(next: OrgStructureTreeSelection) {
    setSelection(next);
    setSelectionCostIRR(null);
    if (!next) {
      setSelectedTeam(null);
      setTeamMembers([]);
    }
    const params = new URLSearchParams(searchParams.toString());
    if (next) params.set("sel", `${next.kind}:${next.id}`);
    else params.delete("sel");
    const nextQuery = params.toString();
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
  }

  const rows = useMemo<Team[]>(() => {
    const needle = query.trim().toLowerCase();
    return [...teams]
      .filter((team) => needle === "" || [team.alias, team.display_name ?? "", team.id].some((value) => value.toLowerCase().includes(needle)))
      .sort((a, b) => getTeamName(a).localeCompare(getTeamName(b)));
  }, [teams, query]);

  async function loadTeams() {
    const search = new URLSearchParams({ limit: "200", offset: "0" });
    const payload = await listTeams(search);
    setTeams(payload.data ?? []);
  }

  async function refresh() {
    setIsLoading(true);
    try {
      await loadTeams();
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

  // Load the selected team's detail (fresh record + members).
  useEffect(() => {
    if (!selection) return;
    let isMounted = true;

    async function run() {
      setDetailLoading(true);
      try {
        const [team, memberPayload] = await Promise.all([getTeam(selection!.id), listTeamMembers(selection!.id)]);
        if (!isMounted) return;
        setSelectedTeam(team);
        setTeamMembers(memberPayload.data ?? []);
        setTeams((prev) => prev.map((item) => (item.id === team.id ? team : item)));
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

  // Usage for the selected team (single number, existing /api/logs/summary proxy).
  useEffect(() => {
    if (!selection) return;
    const controller = new AbortController();
    const range = getUsagePeriodRange(usagePeriod);
    const params = new URLSearchParams({ from: range.from.toISOString(), to: range.to.toISOString(), team_id: selection.id });

    async function run() {
      try {
        const payload = await getTeamUsage(params, controller.signal);
        setSelectionCostIRR(payload.data?.cost_irr ?? 0);
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") return;
        setSelectionCostIRR(null);
      }
    }

    void run();
    return () => controller.abort();
  }, [selection, usagePeriod]);

  // Best-effort per-team cost breakdown for the table's "cost this month" column.
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

  function handleCreateTeam() {
    if (!canCreateTeam) return;
    if (isTeamsFeatureEnabled === false) { setUpgradeDialogOpen(true); return; }
    setTeamForm(emptyTeamForm);
    setCreateTeamOpen(true);
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

  const teamsDisabled = hasLoadedEntitlements && isTeamsFeatureEnabled === false;
  const noticeLabels = { success: tTeams("alerts.success"), error: tTeams("alerts.error"), warning: tTeams("alerts.warning"), info: tTeams("alerts.warning") };

  return (
    <div className="flex flex-col gap-6">
      <OperationNotice notice={banner} labels={noticeLabels} />

      {teamsDisabled ? (
        <EmptyState
          icon={Users}
          title={t("entitlements.disabledTitle")}
          description={t("entitlements.disabledDescription")}
          primaryAction={<Button type="button" onClick={() => setUpgradeDialogOpen(true)}>{t("entitlements.upgrade")}</Button>}
        />
      ) : (
        <section className="grid min-w-0 gap-6 2xl:grid-cols-[minmax(0,1.2fr)_minmax(22rem,0.8fr)]">
          <AppCard className="min-w-0 shadow-none" contentClassName="flex min-w-0 flex-col gap-4">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
              <SearchField containerClassName="w-full sm:min-w-64 sm:flex-1" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("search.placeholder")} />
              <div className="flex flex-wrap gap-2 sm:shrink-0 sm:flex-nowrap">
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
              isLoading={isLoading && teams.length === 0}
              isSearching={query.trim() !== ""}
              searchQuery={query.trim()}
              selection={selection}
              onSelectTeam={(teamID) => updateSelection({ kind: "team", id: teamID })}
              costByTeamID={costByTeamID}
              locale={locale}
            />
          </AppCard>

          <div className="flex min-w-0 flex-col gap-4">
            {selection && selectedTeam ? (
              <TeamDetail
                team={selectedTeam}
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
        <OperationNotice notice={overlayNotice} labels={noticeLabels} />
        <TeamFormFields formID="org-structure-create-team-form" form={teamForm} setForm={setTeamForm} onSubmit={submitCreateTeam} />
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
        <OperationNotice notice={overlayNotice} labels={noticeLabels} />
        <TeamFormFields formID="org-structure-edit-team-form" form={teamForm} setForm={setTeamForm} onSubmit={submitEditTeam} />
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
          feedback={<OperationNotice notice={deleteNotice} labels={noticeLabels} />}
          disabled={!canDeleteTeam || getActiveKeysOrMembers(selectedTeam, teamMembers)}
          warning={<Alert variant="warning"><AlertTriangle aria-hidden="true" /><AlertTitle>{tTeams("delete.guardTitle")}</AlertTitle><AlertDescription>{tTeams("delete.guardDescription")}</AlertDescription></Alert>}
        />
      ) : null}

      <CommercialPlanUpgradeDialog
        open={upgradeDialogOpen}
        onOpenChange={setUpgradeDialogOpen}
        title={teamsDisabled ? t("entitlements.disabledTitle") : tTeams("upgradeDialog.title")}
        description={teamsDisabled ? t("entitlements.disabledDescription") : tTeams("upgradeDialog.description")}
        cancelLabel={tTeams("upgradeDialog.cancel")}
        actionLabel={tTeams("upgradeDialog.action")}
      />
    </div>
  );
}

function TeamFormFields({ formID, form, setForm, onSubmit }: { formID: string; form: TeamForm; setForm: (value: TeamForm) => void; onSubmit: (event: FormEvent<HTMLFormElement>) => void }) {
  const tTeams = useTranslations("teams");
  const update = <K extends keyof TeamForm>(key: K, value: TeamForm[K]) => setForm({ ...form, [key]: value });
  return (
    <form id={formID} className="flex flex-col gap-4" onSubmit={onSubmit}>
      <FieldGroup>
        <DisplayNameField id="org-structure-team-display" label={tTeams("form.displayName")} value={form.displayName} onChange={(event) => update("displayName", event.target.value)} />
        <IdentifierField id="org-structure-team-alias" label={tTeams("form.alias")} value={form.alias} onChange={(event) => update("alias", event.target.value)} required />
        <AppSelectField label={tTeams("form.workspaceType")} value={form.workspaceType} onValueChange={(value) => update("workspaceType", value)} triggerClassName="h-10" options={WORKSPACE_TYPES.map((type) => ({ value: type, label: tTeams(`workspace.${type}`) }))} />
        <Field><FieldLabel htmlFor="org-structure-team-description">{tTeams("form.description")}</FieldLabel><Input id="org-structure-team-description" value={form.description} onChange={(event) => update("description", event.target.value)} /></Field>
      </FieldGroup>
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
  return null;
}

function buildTeamPayload(form: TeamForm, metadata?: Record<string, unknown>) {
  return {
    alias: form.alias.trim(),
    display_name: form.displayName.trim(),
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
    workspaceType: getWorkspaceType(team),
    description: toText(team.metadata?.description),
  };
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
