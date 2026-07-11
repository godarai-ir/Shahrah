"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Check,
  Copy,
  KeyRound,
  Plus,
  RefreshCcw,
  RotateCw,
  Search,
  ShieldCheck,
  Trash2,
} from "lucide-react";

import { AppCard, AppSelectField, CommercialPlanUpgradeDialog, EmptyState, HelpLink, IdentifierField, PageHeader, StatusBadge as AppStatusBadge } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group";
import { Label } from "@/components/ui/label";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";

type VirtualUser = {
  id: string;
  tenant_id: string;
  alias: string;
  display_name?: string;
  org_unit_id?: string;
  team_id?: string;
  active: boolean;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

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

type VATToken = {
  id: string;
  key?: string;
  key_alias: string;
  key_name?: string;
  key_type?: string;
  allowed_models?: string[];
  models?: string[];
  team_id?: string;
  org_unit_id?: string;
  virtual_user_id?: string;
  created_at?: string;
  expires_at?: string;
  last_used_at?: string;
  revoked_at?: string;
  blocked?: boolean;
};

type ListResponse<T> = {
  data?: T[];
  total?: number;
};

type AllowedModelResponse = {
  data?: Array<{ id: string }>;
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
type EffectiveEntitlements = { features?: Record<string, unknown> } | null;

const PAGE_SIZE = 25;
const FEATURE_VAT_ENABLED = "auth.vat.enabled";
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
const dateFormatter = new Intl.DateTimeFormat("fa-IR", {
  dateStyle: "short",
  timeStyle: "medium",
  timeZone: "Asia/Tehran",
});

export default function VirtualUserTokenManagement({
  initialVirtualUsers,
  initialTotal,
  initialError,
  teams,
  orgUnits,
  docsHref,
}: {
  initialVirtualUsers: VirtualUser[];
  initialTotal: number;
  initialError: string | null;
  teams: Team[];
  orgUnits: OrgUnit[];
  docsHref?: string;
}) {
  const t = useTranslations("virtualUsers");
  const [virtualUsers, setVirtualUsers] = useState(initialVirtualUsers);
  const [total, setTotal] = useState(initialTotal);
  const [query, setQuery] = useState("");
  const [offset, setOffset] = useState(0);
  const [selectedID, setSelectedID] = useState(initialVirtualUsers[0]?.id ?? "");
  const [tokens, setTokens] = useState<VATToken[]>([]);
  const [tokenTotal, setTokenTotal] = useState(0);
  const [allowedModels, setAllowedModels] = useState<string[]>([]);
  const [isLoadingUsers, setIsLoadingUsers] = useState(false);
  const [isLoadingTokens, setIsLoadingTokens] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isCopyingSecret, setIsCopyingSecret] = useState(false);
  const [secretCopied, setSecretCopied] = useState(false);
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
  const [deleteConfirmation, setDeleteConfirmation] = useState("");
  const [banner, setBanner] = useState<Banner | null>(initialError ? { kind: "error", message: initialError } : null);
  const isVatFeatureEnabled = hasLoadedEntitlements
    ? readBooleanFeature(entitlements?.features?.[FEATURE_VAT_ENABLED], false)
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
      const payload = await requestJSON<ListResponse<VirtualUser>>(`/api/virtual-users?${params.toString()}`);
      const next = payload.data ?? [];
      setVirtualUsers(next);
      setTotal(payload.total ?? next.length);
      setOffset(nextOffset);
      if (!preferredSelectedID || !next.some((item) => item.id === preferredSelectedID)) {
        const nextSelectedID = next[0]?.id ?? "";
        setSelectedID(nextSelectedID);
        if (!nextSelectedID) {
          setTokens([]);
          setTokenTotal(0);
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
      const payload = await requestJSON<ListResponse<VATToken>>(
        `/api/virtual-users/${encodeURIComponent(virtualUserID)}/tokens?limit=50&offset=0`,
      );
      setTokens(payload.data ?? []);
      setTokenTotal(payload.total ?? payload.data?.length ?? 0);
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
      const payload = await requestJSON<AllowedModelResponse>(`/api/allowed-models?key_type=vat&team_id=${encodeURIComponent(teamID)}`);
      setAllowedModels((payload.data ?? []).map((item) => item.id.trim()).filter(Boolean));
    } catch {
      setAllowedModels([]);
    }
  }

  async function createVirtualUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validation = validateVirtualUserForm(virtualUserForm, t);
    if (validation) {
      setBanner({ kind: "error", message: validation });
      return;
    }
    setIsSaving(true);
    try {
      const created = await requestJSON<VirtualUser>("/api/virtual-users", {
        method: "POST",
        body: JSON.stringify(buildVirtualUserPayload(virtualUserForm)),
      });
      setCreateUserOpen(false);
      setVirtualUserForm(emptyVirtualUserForm);
      setVirtualUsers((prev) => [created, ...prev.filter((item) => item.id !== created.id)]);
      setSelectedID(created.id);
      setTokens([]);
      setTokenTotal(0);
      await loadAllowedModels(created.team_id ?? "");
      setBanner({ kind: "success", message: t("messages.createUserSuccess") });
      await loadVirtualUsers(0, query.trim(), created.id);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.createUserError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function updateVirtualUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedVirtualUser) return;
    const validation = validateVirtualUserForm(virtualUserForm, t);
    if (validation) {
      setBanner({ kind: "error", message: validation });
      return;
    }
    setIsSaving(true);
    try {
      const updated = await requestJSON<VirtualUser>(`/api/virtual-users/${encodeURIComponent(selectedVirtualUser.id)}`, {
        method: "PATCH",
        body: JSON.stringify(buildVirtualUserPayload(virtualUserForm, true)),
      });
      setVirtualUsers((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      setEditUserOpen(false);
      setBanner({ kind: "success", message: t("messages.updateUserSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.updateUserError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function deleteVirtualUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedVirtualUser) return;
    if (deleteConfirmation.trim() !== selectedVirtualUser.alias) {
      setBanner({ kind: "error", message: t("messages.deleteConfirmAlias") });
      return;
    }
    setIsSaving(true);
    try {
      await requestJSON(`/api/virtual-users/${encodeURIComponent(selectedVirtualUser.id)}`, { method: "DELETE" });
      setDeleteUserOpen(false);
      setDeleteConfirmation("");
      setSelectedID("");
      setTokens([]);
      setTokenTotal(0);
      setAllowedModels([]);
      setBanner({ kind: "success", message: t("messages.deleteUserSuccess") });
      await loadVirtualUsers(0, query.trim());
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.deleteUserError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function createToken(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedVirtualUser) return;
    if (!selectedVirtualUser.team_id) {
      setBanner({ kind: "warning", message: t("messages.teamRequired") });
      return;
    }
    if (!tokenForm.keyAlias.trim()) {
      setBanner({ kind: "error", message: t("messages.tokenNameRequired") });
      return;
    }
    setIsSaving(true);
    try {
      const created = await requestJSON<VATToken>(`/api/virtual-users/${encodeURIComponent(selectedVirtualUser.id)}/tokens`, {
        method: "POST",
        body: JSON.stringify({
          key_alias: tokenForm.keyAlias.trim(),
          expires_in: tokenForm.expiresIn.trim() || undefined,
          allowed_models: selectedModels,
        }),
      });
      setNewSecret(created.key ?? null);
      setSecretCopied(false);
      setSecretDialogOpen(Boolean(created.key));
      setTokenForm(emptyTokenForm);
      setCreateTokenOpen(false);
      setBanner({ kind: "success", message: t("messages.createTokenSuccess") });
      await loadTokens(selectedVirtualUser.id);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.createTokenError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function confirmTokenAction() {
    if (!selectedVirtualUser || !actionIntent) return;
    setIsSaving(true);
    try {
      if (actionIntent.type === "revoke") {
        await requestJSON(
          `/api/virtual-users/${encodeURIComponent(selectedVirtualUser.id)}/tokens/${encodeURIComponent(actionIntent.token.id)}/revoke`,
          {
            method: "POST",
            body: JSON.stringify({ reason: "tenant_admin_ui" }),
          },
        );
        setBanner({ kind: "success", message: t("messages.revokeSuccess") });
      } else {
        const rotated = await requestJSON<VATToken>(
          `/api/virtual-users/${encodeURIComponent(selectedVirtualUser.id)}/tokens/${encodeURIComponent(actionIntent.token.id)}/rotate`,
          {
            method: "POST",
            body: JSON.stringify({}),
          },
        );
        setNewSecret(rotated.key ?? null);
        setSecretCopied(false);
        setSecretDialogOpen(Boolean(rotated.key));
        setBanner({ kind: "success", message: t("messages.rotateSuccess") });
      }
      setConfirmOpen(false);
      setActionIntent(null);
      await loadTokens(selectedVirtualUser.id);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.tokenActionError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function copySecret() {
    if (!newSecret || isCopyingSecret) return;
    setIsCopyingSecret(true);
    try {
      if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(newSecret);
      } else if (typeof document !== "undefined") {
        const textarea = document.createElement("textarea");
        textarea.value = newSecret;
        textarea.style.position = "fixed";
        textarea.style.opacity = "0";
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        document.execCommand("copy");
        document.body.removeChild(textarea);
      }
      setSecretCopied(true);
    } catch {
      setBanner({ kind: "error", message: t("oneTime.copyFailed") });
    } finally {
      setIsCopyingSecret(false);
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

  function openEditVirtualUser() {
    if (!selectedVirtualUser) return;
    setVirtualUserForm({
      alias: selectedVirtualUser.alias,
      displayName: selectedVirtualUser.display_name ?? "",
      teamID: selectedVirtualUser.team_id ?? "",
      orgUnitID: selectedVirtualUser.org_unit_id ?? "",
      active: selectedVirtualUser.active,
    });
    setEditUserOpen(true);
  }

  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + virtualUsers.length, total);

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("title")}
        titleAddon={docsHref ? <HelpLink href={docsHref} /> : null}
        description={t("description")}
        actions={
          <Button type="button" onClick={openCreateVirtualUser} disabled={!hasLoadedEntitlements}>
            <Plus data-icon="inline-start" />
            {t("actions.createUser")}
          </Button>
        }
      />

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : banner.kind}>
          <AlertTitle>{t(`alerts.${banner.kind}`)}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <section className="grid min-w-0 gap-6 2xl:grid-cols-[minmax(22rem,0.9fr)_minmax(0,1.1fr)]">
        <AppCard className="min-w-0" contentClassName="flex min-w-0 flex-col gap-4">
          <div className="flex flex-wrap items-center gap-2">
            <InputGroup className="min-w-64 flex-1">
              <InputGroupAddon>
                <Search aria-hidden="true" />
              </InputGroupAddon>
              <InputGroupInput
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={t("searchPlaceholder")}
                aria-label={t("searchPlaceholder")}
              />
            </InputGroup>
            <Button
              variant="outline"
              onClick={() => void loadVirtualUsers(0, query.trim())}
              disabled={isLoadingUsers}
            >
              <RefreshCcw data-icon="inline-start" className={cn(isLoadingUsers && "animate-spin")} />
              {t("actions.refresh")}
            </Button>
          </div>

          <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("usersTable.identity")}</TableHead>
                <TableHead>{t("usersTable.team")}</TableHead>
                <TableHead>{t("usersTable.status")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoadingUsers && virtualUsers.length === 0 ? (
                Array.from({ length: 4 }).map((_, index) => (
                  <TableRow key={index}>
                    <TableCell colSpan={3}><Skeleton className="h-10 w-full" /></TableCell>
                  </TableRow>
                ))
              ) : virtualUsers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={3} className="py-8">
                    <EmptyState
                      title={t("usersTable.empty")}
                      description={t("usersTable.emptyDescription")}
                      primaryAction={
                        <Button type="button" onClick={openCreateVirtualUser} disabled={!hasLoadedEntitlements}>
                          <Plus data-icon="inline-start" />
                          {t("actions.createUser")}
                        </Button>
                      }
                      className="border-0 bg-transparent"
                    />
                  </TableCell>
                </TableRow>
              ) : (
                virtualUsers.map((item) => (
                  <TableRow
                    key={item.id}
                    className={cn("cursor-pointer", item.id === selectedID && "bg-muted")}
                    onClick={() => setSelectedID(item.id)}
                  >
                    <TableCell>
                      <div className="flex min-w-0 flex-col gap-1">
                        <span className="truncate font-medium">{item.display_name || item.alias}</span>
                        <span className="truncate font-mono text-xs text-muted-foreground">{item.alias}</span>
                      </div>
                    </TableCell>
                    <TableCell>{getTeamName(item.team_id, teams)}</TableCell>
                    <TableCell>
                      <StatusBadge active={item.active} t={t} />
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
          </div>

          <div className="flex items-center justify-between gap-3 text-sm text-muted-foreground">
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
            <>
              <Card>
                <CardHeader className="flex flex-row items-start justify-between gap-4">
                  <div className="flex min-w-0 flex-col gap-1">
                    <CardTitle className="truncate">{selectedVirtualUser.display_name || selectedVirtualUser.alias}</CardTitle>
                    <p className="truncate font-mono text-xs text-muted-foreground">{selectedVirtualUser.id}</p>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button variant="outline" size="sm" onClick={openEditVirtualUser}>
                      {t("actions.editUser")}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setDeleteConfirmation("");
                        setDeleteUserOpen(true);
                      }}
                    >
                      <Trash2 data-icon="inline-start" />
                      {t("actions.deleteUser")}
                    </Button>
                  </div>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-3">
                    <InfoItem label={t("detail.team")} value={getTeamName(selectedVirtualUser.team_id, teams)} />
                    <InfoItem label={t("detail.orgUnit")} value={getOrgUnitName(selectedVirtualUser.org_unit_id, orgUnits)} />
                    <InfoItem label={t("detail.tokens")} value={String(tokenTotal)} />
                  </dl>
                </CardContent>
              </Card>

              <AppCard contentClassName="flex min-w-0 flex-col gap-4">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <KeyRound className="text-muted-foreground" />
                    <h2 className="text-lg font-semibold">{t("tokens.title")}</h2>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button variant="outline" onClick={() => void loadTokens(selectedVirtualUser.id)} disabled={isLoadingTokens}>
                      <RefreshCcw data-icon="inline-start" className={cn(isLoadingTokens && "animate-spin")} />
                      {t("actions.refresh")}
                    </Button>
                    <Sheet open={createTokenOpen} onOpenChange={setCreateTokenOpen}>
                      <Button
                        type="button"
                        onClick={openCreateToken}
                        disabled={!hasLoadedEntitlements || !selectedVirtualUser.active || !selectedVirtualUser.team_id}
                      >
                        <Plus data-icon="inline-start" />
                        {t("actions.createToken")}
                      </Button>
                      <SheetContent side="left" className="w-full max-w-xl overflow-y-auto sm:w-3/4">
                        <SheetHeader>
                          <SheetTitle>{t("createToken.title")}</SheetTitle>
                          <SheetDescription>{t("createToken.description")}</SheetDescription>
                        </SheetHeader>
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
                  </div>
                </div>

                {!selectedVirtualUser.team_id ? (
                  <Alert variant="warning">
                    <ShieldCheck />
                    <AlertTitle>{t("messages.teamRequiredTitle")}</AlertTitle>
                    <AlertDescription>{t("messages.teamRequired")}</AlertDescription>
                  </Alert>
                ) : null}

                <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("tokensTable.name")}</TableHead>
                      <TableHead>{t("tokensTable.masked")}</TableHead>
                      <TableHead>{t("tokensTable.models")}</TableHead>
                      <TableHead>{t("tokensTable.lastUsed")}</TableHead>
                      <TableHead>{t("tokensTable.status")}</TableHead>
                      <TableHead>{t("tokensTable.actions")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tokens.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={6}>{t("tokensTable.empty")}</TableCell>
                      </TableRow>
                    ) : (
                      tokens.map((token) => (
                        <TableRow key={token.id}>
                          <TableCell>{token.key_alias}</TableCell>
                          <TableCell className="font-mono text-xs">{token.key_name ?? "-"}</TableCell>
                          <TableCell>{modelSummary(token.allowed_models ?? token.models ?? [], t)}</TableCell>
                          <TableCell>{formatDate(token.last_used_at)}</TableCell>
                          <TableCell>
                            <TokenStatusBadge token={token} t={t} />
                          </TableCell>
                          <TableCell>
                            <div className="flex flex-wrap gap-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => {
                                  setActionIntent({ type: "rotate", token });
                                  setConfirmOpen(true);
                                }}
                                disabled={Boolean(token.revoked_at)}
                              >
                                <RotateCw data-icon="inline-start" />
                                {t("actions.rotate")}
                              </Button>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => {
                                  setActionIntent({ type: "revoke", token });
                                  setConfirmOpen(true);
                                }}
                                disabled={Boolean(token.revoked_at)}
                              >
                                <Trash2 data-icon="inline-start" />
                                {t("actions.revoke")}
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
            </>
          )}
        </div>
      </section>

      <Sheet open={createUserOpen} onOpenChange={setCreateUserOpen}>
        <SheetContent side="left" className="w-full max-w-xl overflow-y-auto sm:w-3/4">
          <SheetHeader>
            <SheetTitle>{t("createUser.title")}</SheetTitle>
            <SheetDescription>{t("createUser.description")}</SheetDescription>
          </SheetHeader>
          <VirtualUserFormView
            t={t}
            form={virtualUserForm}
            setForm={setVirtualUserForm}
            teams={teams}
            orgUnits={orgUnits}
            onSubmit={createVirtualUser}
            isSaving={isSaving}
            mode="create"
          />
        </SheetContent>
      </Sheet>

      <Sheet open={editUserOpen} onOpenChange={setEditUserOpen}>
        <SheetContent side="left" className="w-full max-w-xl overflow-y-auto sm:w-3/4">
          <SheetHeader>
            <SheetTitle>{t("editUser.title")}</SheetTitle>
            <SheetDescription>{t("editUser.description")}</SheetDescription>
          </SheetHeader>
          <VirtualUserFormView
            t={t}
            form={virtualUserForm}
            setForm={setVirtualUserForm}
            teams={teams}
            orgUnits={orgUnits}
            onSubmit={updateVirtualUser}
            isSaving={isSaving}
            mode="edit"
          />
        </SheetContent>
      </Sheet>

      <Dialog
        open={deleteUserOpen}
        onOpenChange={(open) => {
          if (isSaving) return;
          setDeleteUserOpen(open);
          if (!open) setDeleteConfirmation("");
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteUser.title")}</DialogTitle>
            <DialogDescription>
              {selectedVirtualUser
                ? t("deleteUser.description", { alias: selectedVirtualUser.alias })
                : t("deleteUser.fallbackDescription")}
            </DialogDescription>
          </DialogHeader>
          {activeTokenCount > 0 ? (
            <Alert variant="warning">
              <ShieldCheck />
              <AlertTitle>{t("deleteUser.activeTokenWarningTitle")}</AlertTitle>
              <AlertDescription>{t("messages.deleteUserActiveTokenWarning", { count: activeTokenCount })}</AlertDescription>
            </Alert>
          ) : null}
          <form className="flex flex-col gap-4" onSubmit={deleteVirtualUser}>
            <div className="flex flex-col gap-2">
              <Label htmlFor="delete-virtual-user-alias">
                {selectedVirtualUser
                  ? t("deleteUser.confirmLabel", { alias: selectedVirtualUser.alias })
                  : t("fields.alias")}
              </Label>
              <Input
                id="delete-virtual-user-alias"
                value={deleteConfirmation}
                onChange={(event) => setDeleteConfirmation(event.target.value)}
                placeholder={selectedVirtualUser?.alias ?? ""}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDeleteUserOpen(false)} disabled={isSaving}>
                {t("confirm.cancel")}
              </Button>
              <Button type="submit" variant="destructive" disabled={isSaving}>
                {isSaving ? t("actions.saving") : t("actions.deleteUser")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog
        open={confirmOpen}
        onOpenChange={(open) => {
          if (isSaving) return;
          setConfirmOpen(open);
          if (!open) setActionIntent(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {actionIntent?.type === "rotate" ? t("confirm.rotateTitle") : t("confirm.revokeTitle")}
            </DialogTitle>
            <DialogDescription>
              {actionIntent?.type === "rotate"
                ? t("confirm.rotateDescription", { name: actionIntent.token.key_alias })
                : t("confirm.revokeDescription", { name: actionIntent?.token.key_alias ?? "" })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(false)} disabled={isSaving}>
              {t("confirm.cancel")}
            </Button>
            <Button variant={actionIntent?.type === "revoke" ? "destructive" : "default"} onClick={() => void confirmTokenAction()} disabled={isSaving}>
              {isSaving ? t("actions.saving") : t("confirm.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={secretDialogOpen && Boolean(newSecret)}
        onOpenChange={(open) => {
          setSecretDialogOpen(open);
          if (!open) {
            setSecretCopied(false);
            setNewSecret(null);
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("oneTime.title")}</DialogTitle>
            <DialogDescription>{t("oneTime.description")}</DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3">
            <Input value={newSecret ?? ""} readOnly className="font-mono text-xs" />
            <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
              <span>{t("oneTime.copyHint")}</span>
              {secretCopied ? <span className="text-foreground">{t("oneTime.copied")}</span> : null}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSecretDialogOpen(false)}>
              {t("oneTime.close")}
            </Button>
            <Button onClick={() => void copySecret()} disabled={isCopyingSecret || !newSecret}>
              {secretCopied ? <Check data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
              {isCopyingSecret ? t("oneTime.copying") : secretCopied ? t("oneTime.copied") : t("oneTime.copy")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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
  isSaving,
  mode,
}: {
  t: ReturnType<typeof useTranslations>;
  form: VirtualUserForm;
  setForm: (value: VirtualUserForm) => void;
  teams: Team[];
  orgUnits: OrgUnit[];
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  isSaving: boolean;
  mode: "create" | "edit";
}) {
  const update = <K extends keyof VirtualUserForm>(key: K, value: VirtualUserForm[K]) => setForm({ ...form, [key]: value });
  return (
    <form className="mt-6 flex flex-col gap-4" onSubmit={onSubmit}>
      <IdentifierField id="vu-alias" label={t("fields.alias")} value={form.alias} onChange={(event) => update("alias", event.target.value)} required />
      <div className="flex flex-col gap-2">
        <Label htmlFor="vu-display">{t("fields.displayName")}</Label>
        <Input id="vu-display" value={form.displayName} onChange={(event) => update("displayName", event.target.value)} />
      </div>
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
      <Button type="submit" disabled={isSaving}>
        {isSaving ? t("actions.saving") : mode === "create" ? t("actions.createUser") : t("actions.saveUser")}
      </Button>
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
      <div className="flex flex-col gap-2">
        <Label htmlFor="vat-name">{t("fields.tokenName")}</Label>
        <Input id="vat-name" value={form.keyAlias} onChange={(event) => update("keyAlias", event.target.value)} required />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="vat-expiry">{t("fields.expiresIn")}</Label>
        <Input id="vat-expiry" value={form.expiresIn} onChange={(event) => update("expiresIn", event.target.value)} placeholder="30d" />
      </div>
      <div className="flex flex-col gap-2">
        <Label>{t("fields.allowedModels")}</Label>
        <div className="max-h-64 overflow-auto rounded-md border p-2">
          {modelOptions.length === 0 ? (
            <p className="text-xs text-muted-foreground">{t("createToken.noModels")}</p>
          ) : (
            modelOptions.map((model) => (
              <label key={model.id} htmlFor={`model-${model.id}`} className="flex items-center gap-2 py-1 text-sm">
                <Checkbox
                  id={`model-${model.id}`}
                  checked={model.checked}
                  onCheckedChange={(checked) => {
                    if (checked === true) {
                      update("allowedModels", Array.from(new Set([...selectedModels, model.id])).join(","));
                      return;
                    }
                    update("allowedModels", selectedModels.filter((item) => item !== model.id).join(","));
                  }}
                />
                <span className="font-mono text-xs">{model.id}</span>
              </label>
            ))
          )}
        </div>
        <p className="text-xs text-muted-foreground">{t("createToken.modelHint")}</p>
      </div>
      <Button type="submit" disabled={isSaving}>
        {isSaving ? t("actions.saving") : t("actions.createToken")}
      </Button>
    </form>
  );
}

function StatusBadge({ active, t }: { active: boolean; t: ReturnType<typeof useTranslations> }) {
  return <AppStatusBadge status={active ? "active" : "inactive"}>{active ? t("status.active") : t("status.inactive")}</AppStatusBadge>;
}

function TokenStatusBadge({ token, t }: { token: VATToken; t: ReturnType<typeof useTranslations> }) {
  if (token.revoked_at) return <AppStatusBadge status="error">{t("status.revoked")}</AppStatusBadge>;
  if (token.blocked) return <AppStatusBadge status="error">{t("status.blocked")}</AppStatusBadge>;
  return <AppStatusBadge status="active">{t("status.active")}</AppStatusBadge>;
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-1 truncate text-sm font-semibold">{value}</dd>
    </div>
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

function formatDate(input?: string) {
  if (!input) return "-";
  const date = new Date(input);
  if (Number.isNaN(date.getTime())) return "-";
  return dateFormatter.format(date);
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
