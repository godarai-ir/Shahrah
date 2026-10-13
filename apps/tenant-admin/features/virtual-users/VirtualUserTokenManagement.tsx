"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  KeyRound,
  Pencil,
  Plus,
  RefreshCcw,
  RotateCw,
  ShieldCheck,
  Trash2,
} from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { CheckboxOptionList } from "@/components/app-ui/CheckboxOptionList";
import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { ExpirationDurationField } from "@/components/app-ui/ExpirationDurationField";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { ModelNameValue } from "@/components/app-ui/ModelNameValue";
import { OneTimeSecretDialog } from "@/components/app-ui/OneTimeSecretDialog";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { SearchField } from "@/components/app-ui/SearchField";
import { StatusBadge as AppStatusBadge } from "@/components/app-ui/StatusBadge";
import { TableEmptyRow, TableLoadingRows } from "@/components/app-ui/TableStateRows";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { VirtualUser } from "@/lib/contracts/virtual-users";
import { cn } from "@/lib/utils";
import {
  createVATToken,
  createVirtualUser as createVirtualUserRequest,
  deleteVirtualUser as deleteVirtualUserRequest,
  getVirtualUserEntitlements,
  listAllowedVATModels,
  listVATTokens,
  listVirtualUsers,
  revokeVATToken,
  rotateVATToken,
  updateVirtualUser as updateVirtualUserRequest,
  type EffectiveEntitlements,
  type VATToken,
} from "./api";

type Team = {
  id: string;
  alias: string;
  display_name?: string;
  org_unit_id?: string | null;
};

type OrgUnit = {
  id: string;
  alias: string;
  display_name?: string;
};

type VirtualUserForm = {
  alias: string;
  displayName: string;
  teamID: string;
  orgUnitID: string;
  active: boolean;
};

type TokenForm = {
  keyAlias: string;
  expiresIn: string;
  allowedModels: string;
};

type Banner = { kind: "success" | "error" | "warning"; message: string };
type TokenAction = { type: "revoke" | "rotate"; token: VATToken };

const PAGE_SIZE = 25;
const FEATURE_VAT_ENABLED = "auth.vat.enabled";
const VIRTUAL_USER_COLUMN_CLASS = {
  identity: "w-1/3",
  team: "w-1/6",
  orgUnit: "w-1/6",
  status: "w-1/6",
  actions: "w-1/6 text-end",
} as const;
const VAT_COLUMN_CLASS = {
  name: "w-1/5",
  masked: "hidden w-1/5 @lg:table-cell",
  models: "hidden w-2/5 @lg:table-cell",
  lastUsed: "hidden w-36 @2xl:table-cell",
  expiresAt: "hidden w-36 @2xl:table-cell",
  status: "w-1/6",
  actions: "w-24 text-end",
} as const;
const emptyVirtualUserForm: VirtualUserForm = {
  alias: "",
  displayName: "",
  teamID: "",
  orgUnitID: "",
  active: true,
};
const emptyTokenForm: TokenForm = {
  keyAlias: "",
  expiresIn: "30d",
  allowedModels: "",
};

export default function VirtualUserTokenManagement({
  initialVirtualUsers,
  initialTotal,
  initialError,
  teams,
  orgUnits,
}: {
  initialVirtualUsers: VirtualUser[];
  initialTotal: number;
  initialError: string | null;
  teams: Team[];
  orgUnits: OrgUnit[];
}) {
  const t = useTranslations("virtualUsers");
  const locale = useLocale();
  const [virtualUsers, setVirtualUsers] = useState(initialVirtualUsers);
  const [total, setTotal] = useState(initialTotal);
  const [query, setQuery] = useState("");
  const [offset, setOffset] = useState(0);
  const [selectedID, setSelectedID] = useState(initialVirtualUsers[0]?.id ?? "");
  const [tokens, setTokens] = useState<VATToken[]>([]);
  const [allowedModels, setAllowedModels] = useState<string[]>([]);
  const [isLoadingUsers, setIsLoadingUsers] = useState(false);
  const [isLoadingTokens, setIsLoadingTokens] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [secretDialogOpen, setSecretDialogOpen] = useState(false);
  const [newSecret, setNewSecret] = useState<string | null>(null);
  const [createUserOpen, setCreateUserOpen] = useState(false);
  const [editUserOpen, setEditUserOpen] = useState(false);
  const [deleteUserOpen, setDeleteUserOpen] = useState(false);
  const [createTokenOpen, setCreateTokenOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
  const [actionIntent, setActionIntent] = useState<TokenAction | null>(null);
  const [virtualUserForm, setVirtualUserForm] = useState(emptyVirtualUserForm);
  const [tokenForm, setTokenForm] = useState(emptyTokenForm);
  const [banner, setBanner] = useState<Banner | null>(initialError ? { kind: "error", message: initialError } : null);
  const [overlayNotice, setOverlayNotice] = useState<Banner | null>(null);
  const [deleteNotice, setDeleteNotice] = useState<Banner | null>(null);
  const [actionNotice, setActionNotice] = useState<Banner | null>(null);
  const isVatFeatureEnabled = hasLoadedEntitlements
    ? readBooleanFeature(entitlements?.features?.[FEATURE_VAT_ENABLED], false)
    : null;

  useEffect(() => {
    let isMounted = true;

    async function loadEntitlements() {
      try {
        const response = await getVirtualUserEntitlements();
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

  const selectedVirtualUser = useMemo(
    () => virtualUsers.find((item) => item.id === selectedID) ?? null,
    [selectedID, virtualUsers],
  );
  const selectedModels = useMemo(() => modelsFromInput(tokenForm.allowedModels), [tokenForm.allowedModels]);
  const activeTokenCount = useMemo(
    () => tokens.filter((token) => !token.revoked_at).length,
    [tokens],
  );
  const modelOptions = useMemo(() => {
    const selected = new Set(selectedModels);
    return allowedModels.map((model) => ({ id: model, checked: selected.has(model) }));
  }, [allowedModels, selectedModels]);

  useEffect(() => {
    if (selectedVirtualUser) {
      void loadTokens(selectedVirtualUser.id);
      void loadAllowedModels(selectedVirtualUser.team_id ?? "");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedVirtualUser?.id, selectedVirtualUser?.team_id]);

  async function loadVirtualUsers(nextOffset = offset, nextQuery = query.trim(), preferredSelectedID = selectedID) {
    setIsLoadingUsers(true);
    try {
      const params = new URLSearchParams();
      params.set("limit", String(PAGE_SIZE));
      params.set("offset", String(nextOffset));
      if (nextQuery) params.set("query", nextQuery);
      const payload = await listVirtualUsers(params);
      const next = payload.data ?? [];
      setVirtualUsers(next);
      setTotal(payload.total ?? next.length);
      setOffset(nextOffset);
      if (!preferredSelectedID || !next.some((item) => item.id === preferredSelectedID)) {
        const nextSelectedID = next[0]?.id ?? "";
        setSelectedID(nextSelectedID);
        if (!nextSelectedID) {
          setTokens([]);
          setAllowedModels([]);
        }
      }
      setBanner(null);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.loadUsersError")) });
    } finally {
      setIsLoadingUsers(false);
    }
  }

  async function loadTokens(virtualUserID = selectedID) {
    if (!virtualUserID) return;
    setIsLoadingTokens(true);
    try {
      const payload = await listVATTokens(virtualUserID);
      setTokens(payload.data ?? []);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.loadTokensError")) });
    } finally {
      setIsLoadingTokens(false);
    }
  }

  async function loadAllowedModels(teamID: string) {
    if (!teamID) {
      setAllowedModels([]);
      return;
    }
    try {
      setAllowedModels(await listAllowedVATModels(teamID));
    } catch {
      setAllowedModels([]);
    }
  }

  async function createVirtualUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateVirtualUserForm(virtualUserForm, t);
    if (validation) {
      setOverlayNotice({ kind: "error", message: validation });
      return;
    }
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      const created = await createVirtualUserRequest(buildVirtualUserPayload(virtualUserForm));
      setCreateUserOpen(false);
      setVirtualUserForm(emptyVirtualUserForm);
      setVirtualUsers((prev) => [created, ...prev.filter((item) => item.id !== created.id)]);
      setSelectedID(created.id);
      setTokens([]);
      await loadAllowedModels(created.team_id ?? "");
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.createUserSuccess") });
      await loadVirtualUsers(0, query.trim(), created.id);
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, t("messages.createUserError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function updateVirtualUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedVirtualUser) return;
    const validation = validateVirtualUserForm(virtualUserForm, t);
    if (validation) {
      setOverlayNotice({ kind: "error", message: validation });
      return;
    }
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      const updated = await updateVirtualUserRequest(
        selectedVirtualUser.id,
        buildVirtualUserPayload(virtualUserForm, true),
      );
      setVirtualUsers((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      setEditUserOpen(false);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.updateUserSuccess") });
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, t("messages.updateUserError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function deleteVirtualUser() {
    if (!selectedVirtualUser) return;
    setIsSaving(true);
    try {
      await deleteVirtualUserRequest(selectedVirtualUser.id);
      setSelectedID("");
      setTokens([]);
      setAllowedModels([]);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.deleteUserSuccess") });
      await loadVirtualUsers(0, query.trim());
    } catch (error) {
      setDeleteNotice({ kind: "error", message: normalizeError(error, t("messages.deleteUserError")) });
      throw error;
    } finally {
      setIsSaving(false);
    }
  }

  async function createToken(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedVirtualUser) return;
    if (!selectedVirtualUser.team_id) {
      setOverlayNotice({ kind: "warning", message: t("messages.teamRequired") });
      return;
    }
    if (!tokenForm.keyAlias.trim()) {
      setOverlayNotice({ kind: "error", message: t("messages.tokenNameRequired") });
      return;
    }
    setOverlayNotice(null);
    setIsSaving(true);
    try {
      const created = await createVATToken(selectedVirtualUser.id, {
          key_alias: tokenForm.keyAlias.trim(),
          expires_in: tokenForm.expiresIn.trim() || undefined,
          allowed_models: selectedModels,
      });
      setNewSecret(created.key ?? null);
      setSecretDialogOpen(Boolean(created.key));
      setTokenForm(emptyTokenForm);
      setCreateTokenOpen(false);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.createTokenSuccess") });
      await loadTokens(selectedVirtualUser.id);
    } catch (error) {
      setOverlayNotice({ kind: "error", message: normalizeError(error, t("messages.createTokenError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function confirmTokenAction() {
    if (!selectedVirtualUser || !actionIntent) return;
    setActionNotice(null);
    setIsSaving(true);
    try {
      if (actionIntent.type === "revoke") {
        await revokeVATToken(selectedVirtualUser.id, actionIntent.token.id);
        showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.revokeSuccess") });
      } else {
        const rotated = await rotateVATToken(selectedVirtualUser.id, actionIntent.token.id);
        setNewSecret(rotated.key ?? null);
        setSecretDialogOpen(Boolean(rotated.key));
        showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.rotateSuccess") });
      }
      setConfirmOpen(false);
      setActionIntent(null);
      await loadTokens(selectedVirtualUser.id);
    } catch (error) {
      setActionNotice({ kind: "error", message: normalizeError(error, t("messages.tokenActionError")) });
    } finally {
      setIsSaving(false);
    }
  }

  function openCreateVirtualUser() {
    if (isVatFeatureEnabled === false) {
      setUpgradeDialogOpen(true);
      return;
    }
    setVirtualUserForm(emptyVirtualUserForm);
    setCreateUserOpen(true);
  }

  function openCreateToken() {
    if (isVatFeatureEnabled === false) {
      setUpgradeDialogOpen(true);
      return;
    }
    setCreateTokenOpen(true);
  }

  function openEditVirtualUser(virtualUser: VirtualUser | null = selectedVirtualUser) {
    if (!virtualUser) return;
    setVirtualUserForm({
      alias: virtualUser.alias,
      displayName: virtualUser.display_name ?? "",
      teamID: virtualUser.team_id ?? "",
      orgUnitID: virtualUser.org_unit_id ?? "",
      active: virtualUser.active,
    });
    setEditUserOpen(true);
  }

  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + virtualUsers.length, total);

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("title")}
        description={t("description")}
      />

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : banner.kind}>
          <AlertTitle>{t(`alerts.${banner.kind}`)}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <section className="grid min-w-0 gap-6 2xl:grid-cols-[minmax(22rem,0.9fr)_minmax(0,1.1fr)]">
        <AppCard className="min-w-0" contentClassName="flex min-w-0 flex-col gap-4">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <SearchField
              containerClassName="w-full sm:min-w-64 sm:flex-1"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={t("searchPlaceholder")}
            />
            <div className="flex flex-wrap gap-2 sm:shrink-0 sm:flex-nowrap">
              <Button
                type="button"
                onClick={openCreateVirtualUser}
                disabled={!hasLoadedEntitlements}
                className="flex-1 sm:flex-none"
              >
                <Plus data-icon="inline-start" />
                {t("actions.createUser")}
              </Button>
              <Button
                variant="outline"
                size="icon"
                aria-label={t("actions.refresh")}
                title={t("actions.refresh")}
                onClick={() => void loadVirtualUsers(0, query.trim())}
                disabled={isLoadingUsers}
              >
                {isLoadingUsers ? <Spinner /> : <RefreshCcw />}
              </Button>
            </div>
          </div>

          <div className="@container/virtual-users overflow-hidden rounded-lg border border-border">
          <Table className="table-fixed text-xs lg:text-sm">
            <TableHeader>
              <TableRow>
                <TableHead className={VIRTUAL_USER_COLUMN_CLASS.identity}>{t("usersTable.identity")}</TableHead>
                <TableHead className={VIRTUAL_USER_COLUMN_CLASS.team}>{t("usersTable.team")}</TableHead>
                <TableHead className={VIRTUAL_USER_COLUMN_CLASS.orgUnit}>{t("usersTable.orgUnit")}</TableHead>
                <TableHead className={VIRTUAL_USER_COLUMN_CLASS.status}>{t("usersTable.status")}</TableHead>
                <TableHead className={VIRTUAL_USER_COLUMN_CLASS.actions}>{t("usersTable.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoadingUsers && virtualUsers.length === 0 ? (
                <TableLoadingRows colSpan={5} label={t("usersTable.loading")} />
              ) : virtualUsers.length === 0 ? (
                <TableEmptyRow colSpan={5} cellClassName="py-8" title={t("usersTable.empty")} description={t("usersTable.emptyDescription")} primaryAction={<Button type="button" onClick={openCreateVirtualUser} disabled={!hasLoadedEntitlements}><Plus data-icon="inline-start" />{t("actions.createUser")}</Button>} />
              ) : (
                virtualUsers.map((item) => (
                  <TableRow
                    key={item.id}
                    className={cn("cursor-pointer", item.id === selectedID && "bg-muted")}
                    onClick={() => setSelectedID(item.id)}
                  >
                    <TableCell className={VIRTUAL_USER_COLUMN_CLASS.identity}>
                      <div className="flex min-w-0 flex-col gap-1">
                        <span className="truncate font-medium">{item.display_name || item.alias}</span>
                        <div
                          className="flex min-w-0 items-center justify-end gap-1"
                          dir="ltr"
                          onClick={(event) => event.stopPropagation()}
                        >
                          <code className="truncate font-mono text-xs text-muted-foreground" title={item.alias}>
                            {item.alias}
                          </code>
                          <CopyAction
                            value={item.alias}
                            copyLabel={t("actions.copyIdentifier")}
                            copiedLabel={t("messages.identifierCopied")}
                            className="shrink-0"
                          />
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className={VIRTUAL_USER_COLUMN_CLASS.team}>
                      <span className="block truncate">{getTeamName(item.team_id, teams)}</span>
                    </TableCell>
                    <TableCell className={VIRTUAL_USER_COLUMN_CLASS.orgUnit}>
                      <span className="block truncate">{getOrgUnitName(item.org_unit_id, orgUnits)}</span>
                    </TableCell>
                    <TableCell className={VIRTUAL_USER_COLUMN_CLASS.status}>
                      <AppStatusBadge status={item.active ? "active" : "inactive"}>{item.active ? t("status.active") : t("status.inactive")}</AppStatusBadge>
                    </TableCell>
                    <TableCell className={VIRTUAL_USER_COLUMN_CLASS.actions}>
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={t("actions.editUser")}
                          title={t("actions.editUser")}
                          onClick={(event) => {
                            event.stopPropagation();
                            setSelectedID(item.id);
                            openEditVirtualUser(item);
                          }}
                        >
                          <Pencil />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={t("actions.deleteUser")}
                          title={t("actions.deleteUser")}
                          onClick={(event) => {
                            event.stopPropagation();
                            setSelectedID(item.id);
                            setDeleteUserOpen(true);
                          }}
                        >
                          <Trash2 />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
          </div>

          <div className="flex flex-col gap-3 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
            <span>{t("pagination.range", { from, to, total })}</span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => void loadVirtualUsers(Math.max(0, offset - PAGE_SIZE), query.trim())}
                disabled={offset <= 0 || isLoadingUsers}
              >
                {t("pagination.prev")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => void loadVirtualUsers(offset + PAGE_SIZE, query.trim())}
                disabled={offset + virtualUsers.length >= total || isLoadingUsers}
              >
                {t("pagination.next")}
              </Button>
            </div>
          </div>
        </AppCard>

        <div className="flex flex-col gap-4">
          {!selectedVirtualUser ? (
            <EmptyState title={t("detail.emptyTitle")} description={t("detail.emptyDescription")} />
          ) : (
            <AppCard contentClassName="flex min-w-0 flex-col gap-4">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex min-w-0 items-start gap-2">
                    <KeyRound className="text-muted-foreground" />
                    <div className="flex min-w-0 flex-col gap-1">
                      <h2 className="text-lg font-semibold">{t("tokens.title")}</h2>
                      <p className="truncate text-sm text-muted-foreground">
                        {selectedVirtualUser.display_name || selectedVirtualUser.alias}
                      </p>
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Sheet open={createTokenOpen} onOpenChange={(open) => { setCreateTokenOpen(open); if (!open) setOverlayNotice(null); }}>
                      <Button
                        type="button"
                        onClick={openCreateToken}
                        disabled={!hasLoadedEntitlements || !selectedVirtualUser.active || !selectedVirtualUser.team_id}
                      >
                        <Plus data-icon="inline-start" />
                        {t("actions.createToken")}
                      </Button>
                      <SheetContent closeLabel={t("oneTime.close")} className="w-full max-w-xl overflow-y-auto sm:w-3/4">
                        <SheetHeader>
                          <SheetTitle>{t("createToken.title")}</SheetTitle>
                          <SheetDescription>{t("createToken.description")}</SheetDescription>
                        </SheetHeader>
                        <OperationNotice notice={overlayNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} className="mt-6" />
                        <TokenFormView
                          t={t}
                          form={tokenForm}
                          setForm={setTokenForm}
                          modelOptions={modelOptions}
                          onSubmit={createToken}
                          isSaving={isSaving}
                        />
                      </SheetContent>
                    </Sheet>
                    <Button variant="outline" size="icon" aria-label={t("actions.refresh")} title={t("actions.refresh")} onClick={() => void loadTokens(selectedVirtualUser.id)} disabled={isLoadingTokens}>
                      {isLoadingTokens ? <Spinner /> : <RefreshCcw />}
                    </Button>
                  </div>
                </div>

                {!selectedVirtualUser.team_id ? (
                  <Alert variant="warning">
                    <ShieldCheck />
                    <AlertTitle>{t("messages.teamRequiredTitle")}</AlertTitle>
                    <AlertDescription>{t("messages.teamRequired")}</AlertDescription>
                  </Alert>
                ) : null}

                <div className="@container/vat-tokens overflow-hidden rounded-lg border border-border">
                <Table className="table-fixed text-xs lg:text-sm">
                  <TableHeader>
                    <TableRow>
                      <TableHead className={VAT_COLUMN_CLASS.name}>{t("tokensTable.name")}</TableHead>
                      <TableHead className={VAT_COLUMN_CLASS.masked}>{t("tokensTable.masked")}</TableHead>
                      <TableHead className={VAT_COLUMN_CLASS.models}>{t("tokensTable.models")}</TableHead>
                      <TableHead className={VAT_COLUMN_CLASS.lastUsed}>{t("tokensTable.lastUsed")}</TableHead>
                      <TableHead className={VAT_COLUMN_CLASS.expiresAt}>{t("fields.expiresIn")}</TableHead>
                      <TableHead className={VAT_COLUMN_CLASS.status}>{t("tokensTable.status")}</TableHead>
                      <TableHead className={VAT_COLUMN_CLASS.actions}>{t("tokensTable.actions")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tokens.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={7} className="py-6">
                          <EmptyState variant="table" density="compact" title={t("tokensTable.empty")} />
                        </TableCell>
                      </TableRow>
                    ) : (
                      tokens.map((token) => (
                        <TableRow key={token.id}>
                          <TableCell className={VAT_COLUMN_CLASS.name}><span className="block truncate" title={token.key_alias}>{token.key_alias}</span></TableCell>
                          <TableCell className={`${VAT_COLUMN_CLASS.masked} font-mono text-xs`}><span className="block truncate" title={token.key_name ?? "-"}>{token.key_name ?? "-"}</span></TableCell>
                          <TableCell className={VAT_COLUMN_CLASS.models}>
                            {(token.allowed_models ?? token.models ?? []).length > 0 ? (
                              <ModelNameValue
                                value={modelSummary(token.allowed_models ?? token.models ?? [], t)}
                                copyValue={(token.allowed_models ?? token.models ?? []).join(",")}
                              />
                            ) : t("tokens.inheritsAll")}
                          </TableCell>
                          <TableCell className={VAT_COLUMN_CLASS.lastUsed}><DateTimeValue value={token.last_used_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                          <TableCell className={VAT_COLUMN_CLASS.expiresAt}><DateTimeValue value={token.expires_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                          <TableCell className={VAT_COLUMN_CLASS.status}>
                            <AppStatusBadge status={token.revoked_at || token.blocked ? "error" : "active"}>{token.revoked_at ? t("status.revoked") : token.blocked ? t("status.blocked") : t("status.active")}</AppStatusBadge>
                          </TableCell>
                          <TableCell className={VAT_COLUMN_CLASS.actions}>
                            <div className="flex flex-nowrap justify-end gap-2">
                              <Button
                                variant="outline"
                                size="icon-sm"
                                aria-label={t("actions.rotate")}
                                title={t("actions.rotate")}
                                onClick={() => {
                                  setActionIntent({ type: "rotate", token });
                                  setConfirmOpen(true);
                                }}
                                disabled={Boolean(token.revoked_at)}
                              >
                                <RotateCw />
                              </Button>
                              <Button
                                variant="outline"
                                size="icon-sm"
                                aria-label={t("actions.revoke")}
                                title={t("actions.revoke")}
                                onClick={() => {
                                  setActionIntent({ type: "revoke", token });
                                  setConfirmOpen(true);
                                }}
                                disabled={Boolean(token.revoked_at)}
                              >
                                <Trash2 />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
                </div>
            </AppCard>
          )}
        </div>
      </section>

      <FormSheet
        open={createUserOpen}
        onOpenChange={(open) => { setCreateUserOpen(open); if (!open) setOverlayNotice(null); }}
        title={t("createUser.title")}
        description={t("createUser.description")}
        pending={isSaving}
        closeLabel={t("oneTime.close")}
        footer={(
          <FormSheetActions
            secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => { setCreateUserOpen(false); setOverlayNotice(null); }}>{t("actions.cancel")}</Button>}
            primary={<Button type="submit" form="virtual-user-create-form" disabled={isSaving}>{isSaving ? t("actions.saving") : t("actions.createUser")}</Button>}
          />
        )}
      >
          <OperationNotice notice={overlayNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} />
          <VirtualUserFormView
            t={t}
            form={virtualUserForm}
            setForm={setVirtualUserForm}
            teams={teams}
            orgUnits={orgUnits}
            onSubmit={createVirtualUser}
            mode="create"
          />
      </FormSheet>

      <FormSheet
        open={editUserOpen}
        onOpenChange={(open) => { setEditUserOpen(open); if (!open) setOverlayNotice(null); }}
        title={t("editUser.title")}
        description={t("editUser.description")}
        pending={isSaving}
        closeLabel={t("oneTime.close")}
        footer={(
          <FormSheetActions
            secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => { setEditUserOpen(false); setOverlayNotice(null); }}>{t("actions.cancel")}</Button>}
            primary={<Button type="submit" form="virtual-user-edit-form" disabled={isSaving}>{isSaving ? t("actions.saving") : t("actions.saveUser")}</Button>}
          />
        )}
      >
          <OperationNotice notice={overlayNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} />
          <VirtualUserFormView
            t={t}
            form={virtualUserForm}
            setForm={setVirtualUserForm}
            teams={teams}
            orgUnits={orgUnits}
            onSubmit={updateVirtualUser}
            mode="edit"
          />
      </FormSheet>

      {selectedVirtualUser ? (
        <TypedConfirmDialog
          open={deleteUserOpen}
          onOpenChange={(open) => { setDeleteUserOpen(open); if (!open) setDeleteNotice(null); }}
          title={t("deleteUser.title")}
          description={t("deleteUser.description", { alias: selectedVirtualUser.alias })}
          expectedValue={selectedVirtualUser.alias}
          inputLabel={t("deleteUser.confirmLabel", { alias: selectedVirtualUser.alias })}
          cancelLabel={t("confirm.cancel")}
          confirmLabel={t("actions.deleteUser")}
          pendingLabel={t("actions.saving")}
          onConfirm={deleteVirtualUser}
          feedback={<OperationNotice notice={deleteNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} />}
          warning={activeTokenCount > 0 ? (
            <Alert variant="warning">
              <ShieldCheck />
              <AlertTitle>{t("deleteUser.activeTokenWarningTitle")}</AlertTitle>
              <AlertDescription>{t("messages.deleteUserActiveTokenWarning", { count: activeTokenCount })}</AlertDescription>
            </Alert>
          ) : undefined}
        />
      ) : null}

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={(open) => {
          if (isSaving) return;
          setConfirmOpen(open);
          if (!open) {
            setActionIntent(null);
            setActionNotice(null);
          }
        }}
        title={actionIntent?.type === "rotate" ? t("confirm.rotateTitle") : t("confirm.revokeTitle")}
        description={actionIntent?.type === "rotate"
          ? t("confirm.rotateDescription", { name: actionIntent.token.key_alias })
          : t("confirm.revokeDescription", { name: actionIntent?.token.key_alias ?? "" })}
        cancelLabel={t("confirm.cancel")}
        confirmLabel={t("confirm.confirm")}
        pendingLabel={t("actions.saving")}
        pending={isSaving}
        destructive={actionIntent?.type === "revoke"}
        closeOnConfirm={false}
        onConfirm={confirmTokenAction}
        feedback={<OperationNotice notice={actionNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} />}
      />

      <OneTimeSecretDialog
        open={secretDialogOpen}
        onOpenChange={setSecretDialogOpen}
        value={newSecret}
        title={t("oneTime.title")}
        description={t("oneTime.description")}
        warning={t("oneTime.copyHint")}
        closeLabel={t("oneTime.close")}
        copyLabels={{
          valueAriaLabel: t("oneTime.title"),
          copy: t("oneTime.copy"),
          copying: t("oneTime.copying"),
          copied: t("oneTime.copied"),
        }}
        onClear={() => setNewSecret(null)}
        onCopyError={() => showFeedbackToast({ kind: "error", title: t("alerts.error"), description: t("oneTime.copyFailed") })}
      />

      <CommercialPlanUpgradeDialog open={upgradeDialogOpen} onOpenChange={setUpgradeDialogOpen} title={t("upgradeDialog.title")} description={t("upgradeDialog.description")} cancelLabel={t("upgradeDialog.cancel")} actionLabel={t("upgradeDialog.action")} />
    </div>
  );
}

function VirtualUserFormView({
  t,
  form,
  setForm,
  teams,
  orgUnits,
  onSubmit,
  mode,
}: {
  t: ReturnType<typeof useTranslations>;
  form: VirtualUserForm;
  setForm: (value: VirtualUserForm) => void;
  teams: Team[];
  orgUnits: OrgUnit[];
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  mode: "create" | "edit";
}) {
  const update = <K extends keyof VirtualUserForm>(key: K, value: VirtualUserForm[K]) => setForm({ ...form, [key]: value });
  return (
    <form id={`virtual-user-${mode}-form`} className="flex flex-col gap-4" onSubmit={onSubmit}>
      <IdentifierField id="vu-alias" label={t("fields.alias")} value={form.alias} onChange={(event) => update("alias", event.target.value)} required />
      <DisplayNameField id="vu-display" label={t("fields.displayName")} value={form.displayName} onChange={(event) => update("displayName", event.target.value)} />
      <AppSelectField
        label={t("fields.team")}
        ariaLabel={t("fields.team")}
        value={form.teamID || "none"}
        onValueChange={(value) => update("teamID", value === "none" ? "" : value)}
        options={[
          { value: "none", label: t("common.notSet") },
          ...teams.map((team) => ({ value: team.id, label: team.display_name || team.alias })),
        ]}
      />
      <AppSelectField
        label={t("fields.orgUnit")}
        ariaLabel={t("fields.orgUnit")}
        value={form.orgUnitID || "none"}
        onValueChange={(value) => update("orgUnitID", value === "none" ? "" : value)}
        options={[
          { value: "none", label: t("common.notSet") },
          ...orgUnits.map((org) => ({ value: org.id, label: org.display_name || org.alias })),
        ]}
      />
      {mode === "edit" ? (
        <label className="flex items-center gap-2 text-sm" htmlFor="vu-active">
          <Checkbox id="vu-active" checked={form.active} onCheckedChange={(checked) => update("active", checked === true)} />
          <span>{t("fields.active")}</span>
        </label>
      ) : null}
    </form>
  );
}

function TokenFormView({
  t,
  form,
  setForm,
  modelOptions,
  onSubmit,
  isSaving,
}: {
  t: ReturnType<typeof useTranslations>;
  form: TokenForm;
  setForm: (value: TokenForm) => void;
  modelOptions: Array<{ id: string; checked: boolean }>;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  isSaving: boolean;
}) {
  const update = <K extends keyof TokenForm>(key: K, value: TokenForm[K]) => setForm({ ...form, [key]: value });
  const selectedModels = modelsFromInput(form.allowedModels);
  return (
    <form className="mt-6 flex flex-col gap-4" onSubmit={onSubmit}>
      <AppTextField id="vat-name" label={t("fields.tokenName")} value={form.keyAlias} onChange={(event) => update("keyAlias", event.target.value)} required />
      <ExpirationDurationField
        id="vat-expiry"
        label={t("fields.expiresIn")}
        value={form.expiresIn}
        onChange={(event) => update("expiresIn", event.target.value)}
        placeholder="30d"
      />
      <CheckboxOptionList label={t("fields.allowedModels")} description={t("createToken.modelHint")} options={modelOptions.map((model) => ({ value: model.id, label: model.id, textDir: "ltr" }))} value={selectedModels} onValueChange={(nextModels) => update("allowedModels", nextModels.join(","))} emptyText={t("createToken.noModels")} />
      <Button type="submit" disabled={isSaving}>
        {isSaving ? t("actions.saving") : t("actions.createToken")}
      </Button>
    </form>
  );
}

function buildVirtualUserPayload(form: VirtualUserForm, includeActive = false) {
  return {
    alias: form.alias.trim(),
    display_name: form.displayName.trim() || undefined,
    team_id: form.teamID.trim() || undefined,
    org_unit_id: form.orgUnitID.trim() || undefined,
    ...(includeActive ? { active: form.active } : {}),
    metadata: {},
  };
}

function validateVirtualUserForm(form: VirtualUserForm, t: ReturnType<typeof useTranslations>) {
  const alias = form.alias.trim();
  if (!alias) return t("messages.aliasRequired");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{1,62}$/.test(alias)) return t("messages.aliasPattern");
  return null;
}

function getTeamName(teamID: string | undefined, teams: Team[]) {
  if (!teamID) return "-";
  const team = teams.find((item) => item.id === teamID);
  return team ? team.display_name || team.alias : teamID;
}

function getOrgUnitName(orgUnitID: string | undefined, orgUnits: OrgUnit[]) {
  if (!orgUnitID) return "-";
  const org = orgUnits.find((item) => item.id === orgUnitID);
  return org ? org.display_name || org.alias : orgUnitID;
}

function modelSummary(models: string[], t: ReturnType<typeof useTranslations>) {
  if (models.length === 0) return t("tokens.inheritsAll");
  if (models.length === 1) return models[0];
  return t("tokens.modelCount", { count: models.length, first: models[0] });
}

function modelsFromInput(value: string) {
  return value.split(",").map((item) => item.trim()).filter(Boolean);
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
