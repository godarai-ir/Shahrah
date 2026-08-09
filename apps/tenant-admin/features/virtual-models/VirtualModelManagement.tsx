"use client";

import { useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { ArrowDown, ArrowUp, CheckCircle2, Copy, Edit3, Plus, Route, Save, Search, Trash2, XCircle } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { ChipInput } from "@/components/app-ui/ChipInput";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { ModelNameValue } from "@/components/app-ui/ModelNameValue";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { UnsavedChangesDialog } from "@/components/app-ui/UnsavedChangesDialog";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Field, FieldContent, FieldDescription, FieldLabel, FieldLegend, FieldSet, FieldTitle } from "@/components/ui/field";
import { Spinner } from "@/components/ui/spinner";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ModelEntitySelect, type ModelSelection } from "@/components/app-ui/ModelEntitySelect";
import { PrincipalPicker } from "@/components/app-ui/PrincipalPicker";
import { type PrincipalSearchKind } from "@/components/app-ui/PrincipalEntitySelect";
import { buildPrincipalSelectorLabels } from "@/components/app-ui/buildPrincipalSelectorLabels";
import { cn } from "@/lib/utils";
import {
  deleteVirtualModel,
  getProviderAccountDetails,
  getVirtualModel,
  listProviderAccounts,
  listTenantUsers,
  listVirtualModels,
  saveVirtualModel,
  searchAccessPrincipals,
  type SaveVirtualModelPayload,
  type VirtualModel,
  type VirtualModelTarget,
} from "@/features/virtual-models/api";
import {
  DEFAULT_TARGET,
  LIST_PAGE_SIZE,
  MODE_ENDPOINTS,
  buildModeSummary,
  canonicalModelFQN,
  createCopyAlias,
  makeDefaultForm,
  mapWithConcurrency,
  modesForProviderModel,
  normalizeAccess,
  normalizeModeList,
  normalizeTargets,
  parseStatusCodes,
  rebalanceTargetsForWeighted,
  resolveUserLabel,
  sameModelFQN,
  validateForm,
  type VirtualModelFormState,
} from "@/features/virtual-models/domain";

type FormState = VirtualModelFormState;

type Notice = { kind: "success" | "error"; message: string } | null;

export default function VirtualModelManagement({ docsHref }: { docsHref: string }) {
  const t = useTranslations("virtualModels");
  const locale = useLocale();
  const tc = useTranslations("common");
  const principalT = useTranslations("providerAccounts");
  const principalUx = useTranslations("providerAccounts.ux");
  const defaultEveryoneLabel = principalT("principal.everyone");
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [actingID, setActingID] = useState<string>("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [pendingDeleteItem, setPendingDeleteItem] = useState<VirtualModel | null>(null);
  const [query, setQuery] = useState("");
  const [notice, setNotice] = useState<Notice>(null);
  const [loadNotice, setLoadNotice] = useState<Notice>(null);
  const [deleteNotice, setDeleteNotice] = useState<Notice>(null);

  const [items, setItems] = useState<VirtualModel[]>([]);
  const [userLabelByID, setUserLabelByID] = useState<Record<string, string>>({});
  const [providerAllowedModels, setProviderAllowedModels] = useState<ModelSelection[]>([]);
  const [providerModelModes, setProviderModelModes] = useState<Record<string, string[]>>({});
  const [selectedID, setSelectedID] = useState<string>("");
  const [form, setForm] = useState<FormState>(() => makeDefaultForm(defaultEveryoneLabel));
  const [savedHash, setSavedHash] = useState(() => JSON.stringify(makeDefaultForm(defaultEveryoneLabel)));
  const [editorOpen, setEditorOpen] = useState(false);
  const [listPage, setListPage] = useState(0);
  const [pendingDiscardAction, setPendingDiscardAction] = useState<null | (() => void | Promise<void>)>(null);
  const [disableConfirmOpen, setDisableConfirmOpen] = useState(false);

  const isCreate = !form.id;
  const dirty = useMemo(() => JSON.stringify(form) !== savedHash, [form, savedHash]);
  const modeSummary = useMemo(() => buildModeSummary(form.targets, providerModelModes), [form.targets, providerModelModes]);
  const totalWeight = useMemo(
    () => form.targets.reduce((sum, target) => sum + (Number.isFinite(target.weight) ? Number(target.weight) : 0), 0),
    [form.targets],
  );
  const effectiveExposedModes = useMemo(() => {
    const requested = normalizeModeList(form.exposed_modes);
    if (requested.length === 0) return modeSummary.intersection;
    return requested.filter((mode) => modeSummary.intersection.includes(mode));
  }, [form.exposed_modes, modeSummary.intersection]);

  useEffect(() => {
    void loadList();
    void loadProviderAllowedModels();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    void loadUserLabels(items);
  }, [items]);

  useEffect(() => {
    const onBeforeUnload = (event: BeforeUnloadEvent) => {
      if (!dirty) return;
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [dirty]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return items;
    return items.filter((item) => item.alias.toLowerCase().includes(q));
  }, [items, query]);
  const listPageCount = Math.max(1, Math.ceil(filtered.length / LIST_PAGE_SIZE));
  const effectiveListPage = Math.min(listPage, listPageCount - 1);
  const visibleItems = useMemo(
    () => filtered.slice(effectiveListPage * LIST_PAGE_SIZE, (effectiveListPage + 1) * LIST_PAGE_SIZE),
    [effectiveListPage, filtered],
  );

  const principalSelectorLabels = useMemo(
    () => buildPrincipalSelectorLabels(principalT, principalUx),
    [principalT, principalUx],
  );

  async function loadList() {
    setIsLoading(true);
    setLoadNotice(null);
    try {
      setItems(await listVirtualModels());
    } catch (error) {
      setLoadNotice({ kind: "error", message: normalizeError(error, t("messages.listLoadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadDetail(id: string, options?: { skipDirtyConfirm?: boolean }) {
    if (!id) return;
    if (!options?.skipDirtyConfirm && dirty) {
      setPendingDiscardAction(() => () => loadDetail(id, { skipDirtyConfirm: true }));
      return;
    }
    setIsLoading(true);
    setNotice(null);
    try {
      const data = await getVirtualModel(id);
      const mapped: FormState = {
        id: data.id,
        alias: data.alias,
        routing_strategy: data.routing_strategy === "weighted" ? "weighted" : "priority",
        enabled: data.enabled,
        exposed_modes: normalizeModeList(data.exposed_modes),
        targets: normalizeTargets(data.targets),
        access: normalizeAccess(data.access, defaultEveryoneLabel),
      };
      setSelectedID(id);
      setForm(mapped);
      setSavedHash(JSON.stringify(mapped));
      setEditorOpen(true);
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.detailLoadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadProviderAllowedModels() {
    try {
      const accounts = await listProviderAccounts();
      const accountDetails = await mapWithConcurrency(
        accounts,
        4,
        async (account) => {
          try {
            const detail = await getProviderAccountDetails(account.id);
            return { account, models: detail.models ?? [] };
          } catch {
            return null;
          }
        },
      );
      const availableAccountDetails = accountDetails.filter((item) => item !== null);
      const allSelections = availableAccountDetails.map(({ account, models }) =>
        models
            .filter((item) => item.enabled)
            .map<ModelSelection>((item) => ({
              id: `${account.id}:${item.model}`,
              name: `${account.provider}:${account.name}:${item.model}`,
              kind: "provider_allowed",
              secondary: `${account.provider}/${account.name}`,
            })),
      );
      setProviderAllowedModels(allSelections.flat());
      const modeEntries = availableAccountDetails.map(({ account, models }) =>
        models
            .filter((item) => item.enabled)
            .flatMap((item): Array<[string, string[]]> => {
              const rawFQN = `${account.provider}:${account.name}:${item.model}`;
              const modes = modesForProviderModel(item);
              return [
                [rawFQN, modes],
                [canonicalModelFQN(rawFQN), modes],
              ];
            }),
      );
      setProviderModelModes(Object.fromEntries(modeEntries.flat()));
    } catch {
      setProviderAllowedModels([]);
      setProviderModelModes({});
    }
  }

  function startCreate() {
    if (dirty) {
      setPendingDiscardAction(() => resetForCreate);
      return;
    }
    resetForCreate();
  }

  function resetForCreate() {
    const next = makeDefaultForm(defaultEveryoneLabel);
    setSelectedID("");
    setForm(next);
    setSavedHash(JSON.stringify(next));
    setNotice(null);
    setEditorOpen(true);
  }

  function closeEditor() {
    if (dirty) {
      setPendingDiscardAction(() => () => setEditorOpen(false));
      return;
    }
    setEditorOpen(false);
  }

  function handleEditorOpenChange(open: boolean) {
    if (open) {
      setEditorOpen(true);
      return;
    }
    closeEditor();
  }

  function updateTarget(index: number, patch: Partial<VirtualModelTarget>) {
    setForm((prev) => {
      const targets = prev.targets.map((item, i) => (i === index ? { ...item, ...patch } : item));
      return { ...prev, targets };
    });
  }

  function changeRoutingStrategy(strategy: FormState["routing_strategy"]) {
    setForm((prev) => {
      if (prev.routing_strategy === strategy) return prev;
      if (strategy !== "weighted") {
        return { ...prev, routing_strategy: strategy };
      }

      const hasValidWeights = prev.targets.every((target) => (target.weight ?? 0) > 0);
      const nextTargets = hasValidWeights ? prev.targets : rebalanceTargetsForWeighted(prev.targets);

      return {
        ...prev,
        routing_strategy: strategy,
        targets: nextTargets,
      };
    });
  }

  function moveTarget(index: number, direction: -1 | 1) {
    setForm((prev) => {
      const nextIndex = index + direction;
      if (nextIndex < 0 || nextIndex >= prev.targets.length) return prev;
      const targets = [...prev.targets];
      const current = targets[index];
      targets[index] = targets[nextIndex];
      targets[nextIndex] = current;
      const normalized = targets.map((item, idx) => ({ ...item, order: idx + 1 }));
      return { ...prev, targets: normalized };
    });
  }

  function addTarget() {
    setForm((prev) => {
      const nextTargets = [
        ...prev.targets,
        {
          ...DEFAULT_TARGET,
          order: prev.targets.length + 1,
          weight: prev.routing_strategy === "weighted" ? 1 : 100,
        },
      ];

      return {
        ...prev,
        targets: nextTargets,
      };
    });
  }

  function removeTarget(index: number) {
    setForm((prev) => {
      if (prev.targets.length <= 1) return prev;
      const targets = prev.targets.filter((_, i) => i !== index).map((item, idx) => ({ ...item, order: idx + 1 }));
      return { ...prev, targets };
    });
  }

  async function submitForm(skipDisableConfirm = false) {
    const validationError = validateForm(form, t);
    if (validationError) {
      setNotice({ kind: "error", message: validationError });
      return;
    }
    if (modeSummary.ready && modeSummary.intersection.length === 0) {
      setNotice({ kind: "error", message: t("messages.noCommonModes") });
      return;
    }
    if (!form.enabled && !skipDisableConfirm) {
      setDisableConfirmOpen(true);
      return;
    }
    setIsSaving(true);
    setNotice(null);
    const payload = {
      alias: form.alias.trim(),
      routing_strategy: form.routing_strategy,
      enabled: form.enabled,
      exposed_modes: effectiveExposedModes,
      targets: form.targets.map((target, idx) => ({
        ...target,
        order: idx + 1,
      })),
      access: {
        managers: form.access.managers,
        users: form.access.users,
      },
    };
    try {
      const data = await saveVirtualModel(form.id, payload);
      const nextID = data.id ?? form.id ?? "";
      setNotice({ kind: "success", message: isCreate ? t("messages.createSuccess") : t("messages.updateSuccess") });
      // The current form has just been persisted; prevent unsaved confirmation
      // while hydrating the latest server state back into the form.
      setSavedHash(JSON.stringify(form));
      await loadList();
      if (nextID) await loadDetail(nextID, { skipDirtyConfirm: true });
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, isCreate ? t("messages.createError") : t("messages.updateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function onSearchPrincipal(kind: PrincipalSearchKind, query: string) {
    if (query.trim().length < 2) return [];
    try {
      return await searchAccessPrincipals(kind, query.trim());
    } catch {
      return [];
    }
  }

  async function loadUserLabels(rows: VirtualModel[]) {
    const userIDs = [...new Set(rows.map((item) => (item.created_by ?? "").trim()).filter(Boolean))];
    if (userIDs.length === 0) {
      setUserLabelByID({});
      return;
    }
    try {
      const users = await listTenantUsers();
      setUserLabelByID(
        Object.fromEntries(
          userIDs.map((userID) => {
            const match = users.find((item) => item.id === userID);
            return [userID, resolveUserLabel(match, userID)];
          }),
        ),
      );
    } catch {
      setUserLabelByID(Object.fromEntries(userIDs.map((userID) => [userID, userID])));
    }
  }

  async function deleteCurrent() {
    if (!pendingDeleteItem?.id) return;

    setIsDeleting(true);
    setNotice(null);
    try {
      await deleteVirtualModel(pendingDeleteItem.id);

      const next = makeDefaultForm(defaultEveryoneLabel);
      if (form.id === pendingDeleteItem.id) {
        setSelectedID("");
        setForm(next);
        setSavedHash(JSON.stringify(next));
        setEditorOpen(false);
      }
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.deleteSuccess") });
      setDeleteDialogOpen(false);
      setPendingDeleteItem(null);
      await loadList();
    } catch (error) {
      setDeleteNotice({ kind: "error", message: normalizeError(error, t("messages.deleteError")) });
      throw error;
    } finally {
      setIsDeleting(false);
    }
  }

  async function duplicateItem(item: VirtualModel) {
    setActingID(item.id);
    setNotice(null);
    try {
      const detail = await getVirtualModel(item.id);

      const existingAliases = new Set(items.map((entry) => entry.alias.toLowerCase()));
      const nextAlias = createCopyAlias(item.alias, existingAliases);
      const payload: SaveVirtualModelPayload = {
        alias: nextAlias,
        routing_strategy: detail.routing_strategy === "weighted" ? "weighted" : "priority",
        enabled: detail.enabled,
        exposed_modes: normalizeModeList(detail.exposed_modes),
        targets: normalizeTargets(detail.targets).map((target, index) => ({
          ...target,
          id: undefined,
          order: index + 1,
        })),
        access: normalizeAccess(detail.access, defaultEveryoneLabel),
      };

      await saveVirtualModel(undefined, payload);

      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.copySuccess", { alias: nextAlias }) });
      await loadList();
    } catch (error) {
      showFeedbackToast({ kind: "error", title: t("alerts.error"), description: normalizeError(error, t("messages.copyError")) });
    } finally {
      setActingID("");
    }
  }

  return (
    <>
      <OperationNotice notice={loadNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} className="mb-6" />
      <Card className="overflow-hidden">
        <CardHeader className="space-y-3">
          <div className="flex flex-col gap-2 sm:flex-row">
            <div className="flex min-w-0 flex-1 items-center gap-2">
              <div className="relative min-w-0 flex-1">
                <Search className="pointer-events-none absolute start-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  value={query}
                  onChange={(event) => {
                    setQuery(event.target.value);
                    setListPage(0);
                  }}
                  className="ps-9"
                  placeholder={t("searchPlaceholder")}
                  aria-label={t("searchPlaceholder")}
                />
              </div>
              <RefreshButton label={t("actions.refresh")} onClick={() => void loadList()} isLoading={isLoading} />
            </div>
            <div className="flex items-center gap-2">
              <Button className="flex-1 sm:flex-none" onClick={startCreate}>
                <Plus data-icon="inline-start" />
                {t("actions.new")}
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {isLoading && items.length === 0 ? (
            <div className="flex flex-col gap-2">
              <Skeleton className="h-9 w-full rounded-md" />
              <Skeleton className="h-9 w-full rounded-md" />
              <Skeleton className="h-9 w-full rounded-md" />
            </div>
          ) : filtered.length === 0 ? (
            <EmptyState
              icon={Route}
              title={query ? t("empty.noResultsTitle") : t("empty.title")}
              description={query ? t("empty.noResultsDescription") : t("empty.description")}
              primaryAction={
                query ? (
                  <Button type="button" variant="outline" onClick={() => {
                    setQuery("");
                    setListPage(0);
                  }}>
                    {t("actions.clearSearch")}
                  </Button>
                ) : (
                  <Button type="button" onClick={startCreate}>
                    <Plus data-icon="inline-start" />
                    {t("actions.new")}
                  </Button>
                )
              }
              secondaryAction={
                <Button variant="outline" asChild>
                  <a href={docsHref} target="_blank" rel="noreferrer">
                    {t("empty.guide")}
                  </a>
                </Button>
              }
            />
          ) : (
            <div className="@container overflow-x-auto rounded-lg border">
              <Table className="table-fixed">
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("list.alias")}</TableHead>
                    <TableHead className="hidden w-24 @sm:table-cell">{t("list.strategy")}</TableHead>
                    <TableHead className="hidden w-20 @md:table-cell">{t("list.targets")}</TableHead>
                    <TableHead className="hidden w-40 @4xl:table-cell">{t("list.updatedAt")}</TableHead>
                    <TableHead className="hidden w-48 @4xl:table-cell">{t("list.createdBy")}</TableHead>
                    <TableHead className="w-24">{t("list.status")}</TableHead>
                    <TableHead className="w-16 whitespace-nowrap text-center">{t("list.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visibleItems.map((item) => (
                    <TableRow
                      key={item.id}
                      className={cn(selectedID === item.id && "bg-muted/60")}
                    >
                      <TableCell>
                        <ModelNameValue value={item.alias} onCopyError={() => showFeedbackToast({ kind: "error", title: t("alerts.error"), description: t("messages.copyAliasError") })} />
                      </TableCell>
                      <TableCell className="hidden @sm:table-cell">{formatStrategyLabel(item.routing_strategy, t)}</TableCell>
                      <TableCell className="hidden @md:table-cell">{item.targets?.length ?? 0}</TableCell>
                      <TableCell className="hidden text-sm text-muted-foreground @4xl:table-cell"><DateTimeValue value={item.updated_at} locale={locale} format={{ year: "numeric", month: "numeric", day: "numeric", hour: "2-digit", minute: "2-digit" }} /></TableCell>
                      <TableCell className="hidden truncate text-sm text-muted-foreground @4xl:table-cell">{userLabelByID[item.created_by ?? ""] || item.created_by || "-"}</TableCell>
                      <TableCell>
                        <StatusBadge status={item.enabled ? "active" : "inactive"}>
                          {item.enabled ? t("status.enabled") : t("status.disabled")}
                        </StatusBadge>
                      </TableCell>
                      <TableCell onClick={(event) => event.stopPropagation()}>
                        <RowActionsMenu
                          label={t("actions.openMenu", { alias: item.alias })}
                          dir={locale.startsWith("fa") ? "rtl" : "ltr"}
                          disabled={isLoading}
                          actions={[
                            {
                              key: "copy",
                              label: t("actions.copy"),
                              pendingLabel: t("actions.copying"),
                              pending: actingID === item.id,
                              icon: <Copy />,
                              onSelect: () => void duplicateItem(item),
                            },
                            {
                              key: "edit",
                              label: t("actions.edit"),
                              disabled: actingID === item.id,
                              icon: <Edit3 />,
                              onSelect: () => void loadDetail(item.id),
                            },
                            {
                              key: "delete",
                              label: t("actions.delete"),
                              disabled: actingID === item.id,
                              destructive: true,
                              separatorBefore: true,
                              icon: <Trash2 />,
                              onSelect: () => {
                                setPendingDeleteItem(item);
                                setDeleteDialogOpen(true);
                              },
                            },
                          ]}
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {listPageCount > 1 ? (
                <nav className="flex items-center justify-between gap-3 border-t p-3" aria-label={t("pagination.label")}>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    disabled={effectiveListPage === 0}
                    onClick={() => setListPage(Math.max(0, effectiveListPage - 1))}
                  >
                    {t("pagination.previous")}
                  </Button>
                  <span className="text-xs tabular-nums text-muted-foreground">
                    {t("pagination.summary", {
                      start: effectiveListPage * LIST_PAGE_SIZE + 1,
                      end: Math.min((effectiveListPage + 1) * LIST_PAGE_SIZE, filtered.length),
                      total: filtered.length,
                    })}
                  </span>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    disabled={effectiveListPage >= listPageCount - 1}
                    onClick={() => setListPage(Math.min(listPageCount - 1, effectiveListPage + 1))}
                  >
                    {t("pagination.next")}
                  </Button>
                </nav>
              ) : null}
            </div>
          )}
        </CardContent>
      </Card>

      <FormSheet
        open={editorOpen}
        onOpenChange={handleEditorOpenChange}
        title={isCreate ? t("form.createTitle") : t("form.editTitle")}
        description={isCreate ? t("form.createDescription") : t("form.editDescription")}
        size="lg"
        pending={isSaving || isDeleting}
        bodyClassName="flex flex-col gap-4"
        footer={(
          <div className="flex w-full flex-wrap items-center justify-between gap-2">
            <div className="flex flex-wrap items-center gap-2">
              <Button type="button" onClick={() => void submitForm()} disabled={isSaving || isDeleting}>
                {isSaving ? <Spinner /> : <Save />}
                {isCreate ? t("actions.create") : tc("save")}
              </Button>
              <Button type="button" variant="outline" onClick={closeEditor} disabled={isSaving || isDeleting}>
                {tc("cancel")}
              </Button>
              {!isCreate ? (
                <Button type="button" variant="outline" onClick={() => selectedID && void loadDetail(selectedID, { skipDirtyConfirm: true })}>
                  {t("actions.reset")}
                </Button>
              ) : null}
            </div>
            {!isCreate ? (
              <Button type="button" variant="destructive" onClick={() => setDeleteDialogOpen(true)} disabled={isSaving || isDeleting}>
                {isDeleting ? <Spinner /> : <Trash2 />}
                {t("actions.delete")}
              </Button>
            ) : null}
          </div>
        )}
      >
              {notice ? (
                <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
                  {notice.kind === "error" ? <XCircle className="h-4 w-4" /> : <CheckCircle2 className="h-4 w-4" />}
                  <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
                  <AlertDescription>{notice.message}</AlertDescription>
                </Alert>
              ) : null}

              <div className="grid gap-3 lg:grid-cols-2">
                <IdentifierField id="vm-alias" label={t("form.alias")} value={form.alias} onChange={(event) => setForm((prev) => ({ ...prev, alias: event.target.value }))} placeholder={t("form.aliasPlaceholder")} />
                <FieldSet className="gap-2">
                  <FieldLegend variant="label">{t("form.strategy")}</FieldLegend>
                  <RadioGroup
                    value={form.routing_strategy}
                    onValueChange={(value) => changeRoutingStrategy(value as FormState["routing_strategy"])}
                    aria-label={t("form.strategy")}
                    dir={locale.startsWith("fa") ? "rtl" : "ltr"}
                    className="grid gap-2 sm:grid-cols-2"
                  >
                    <FieldLabel htmlFor="vm-strategy-priority" className="cursor-pointer">
                      <Field orientation="horizontal">
                        <RadioGroupItem id="vm-strategy-priority" value="priority" />
                        <FieldContent>
                          <FieldTitle>{t("form.priorityValue")}</FieldTitle>
                          <FieldDescription>{t("form.priorityDescription")}</FieldDescription>
                        </FieldContent>
                      </Field>
                    </FieldLabel>
                    <FieldLabel htmlFor="vm-strategy-weighted" className="cursor-pointer">
                      <Field orientation="horizontal">
                        <RadioGroupItem id="vm-strategy-weighted" value="weighted" />
                        <FieldContent>
                          <FieldTitle>{t("form.weightedValue")}</FieldTitle>
                          <FieldDescription>{t("form.weightedDescription")}</FieldDescription>
                        </FieldContent>
                      </Field>
                    </FieldLabel>
                  </RadioGroup>
                </FieldSet>
              </div>

              <div className="flex items-center justify-between rounded-2xl bg-muted/35 px-4 py-3">
                <div>
                  <Label htmlFor="vm-enabled" className="text-sm font-medium">
                    {t("form.enabledLabel")}
                  </Label>
                  <p className="text-xs text-muted-foreground">{t("form.enabledHelp")}</p>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">{form.enabled ? t("status.enabled") : t("status.disabled")}</span>
                  <Switch
                    id="vm-enabled"
                    checked={form.enabled}
                    onCheckedChange={(checked) => setForm((prev) => ({ ...prev, enabled: checked }))}
                    aria-label={t("form.enabledLabel")}
                  />
                </div>
              </div>

              <div className="flex flex-col gap-4 rounded-2xl bg-muted/20 px-4 py-4">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium">{t("targets.title")}</p>
                    <p className="text-xs text-muted-foreground">{t("targets.description")}</p>
                  </div>
                  <Button type="button" variant="outline" size="sm" onClick={addTarget}>
                    <Plus className="h-4 w-4" />
                    {t("targets.add")}
                  </Button>
                </div>

                {form.targets.map((target, index) => (
                  <div
                    key={`${target.id ?? "new"}-${index}`}
                    className="flex flex-col gap-3 rounded-2xl bg-background/90 px-4 py-4 shadow-sm ring-1 ring-border/40"
                  >
                    <div className="flex items-center justify-between">
                      <Badge variant="secondary">{t("targets.itemLabel", { order: target.order })}</Badge>
                      <div className="flex items-center gap-1">
                        <Button type="button" variant="ghost" size="icon" onClick={() => moveTarget(index, -1)} disabled={index === 0}>
                          <ArrowUp className="h-4 w-4" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => moveTarget(index, 1)}
                          disabled={index === form.targets.length - 1}
                        >
                          <ArrowDown className="h-4 w-4" />
                        </Button>
                        <Button type="button" variant="ghost" size="icon" onClick={() => removeTarget(index)} disabled={form.targets.length <= 1}>
                          <XCircle className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>

                    {form.routing_strategy === "weighted" ? (
                      <AppNumberField
                        id={`vm-target-${index}-weight`}
                        label={t("targets.weight")}
                        min={1}
                        max={100}
                        step="0.01"
                        value={target.weight ?? 0}
                        onChange={(event) =>
                          updateTarget(index, {
                            weight: Number(event.target.value),
                          })
                        }
                      />
                    ) : null}

                    <div className="grid gap-3 lg:grid-cols-2">
                      <div className="space-y-1.5">
                        <Label>{t("targets.provider")}</Label>
                        <ModelEntitySelect
                          id={`vm-target-${index}-provider-model`}
                          selected={
                            providerAllowedModels.find(
                              (item) => sameModelFQN(item.name, target.model) && item.secondary?.startsWith(`${target.provider}/`),
                            ) ?? null
                          }
                          onSelectedChange={(value) => {
                            if (!value) return;
                            const provider = value.secondary?.split("/")[0] ?? "";
                            updateTarget(index, { provider, model: value.name });
                          }}
                          providerAllowedModels={providerAllowedModels}
                          virtualModels={[]}
                          showVirtualModels={false}
                          labels={{
                            triggerPlaceholder: t("targets.model"),
                            searchPlaceholder: t("searchPlaceholder"),
                            providerAllowed: t("targets.provider"),
                            virtualModels: t("title"),
                            loading: "",
                            empty: t("empty.description"),
                          }}
                        />
                      </div>
                    </div>

                    <div className="grid gap-3 md:grid-cols-3">
                      <AppNumberField
                        id={`vm-target-${index}-retry-attempts`}
                        label={t("targets.retryAttempts")}
                        inputMode="numeric"
                        min={0}
                        max={10}
                        value={target.policy.retry.attempts}
                        onChange={(event) =>
                          updateTarget(index, {
                            policy: {
                              ...target.policy,
                              retry: {
                                ...target.policy.retry,
                                attempts: Number(event.target.value),
                              },
                            },
                          })
                        }
                      />
                      <AppNumberField
                        id={`vm-target-${index}-retry-delay`}
                        label={t("targets.retryDelay")}
                        inputMode="numeric"
                        min={0}
                        max={10000}
                        value={target.policy.retry.delay_ms}
                        onChange={(event) =>
                          updateTarget(index, {
                            policy: {
                              ...target.policy,
                              retry: {
                                ...target.policy.retry,
                                delay_ms: Number(event.target.value),
                              },
                            },
                          })
                        }
                      />
                      <AppNumberField
                        id={`vm-target-${index}-timeout`}
                        label={t("targets.timeout")}
                        inputMode="numeric"
                        min={1}
                        max={600}
                        value={target.policy.timeout_sec ?? 60}
                        onChange={(event) =>
                          updateTarget(index, {
                            policy: {
                              ...target.policy,
                              timeout_sec: Number(event.target.value),
                            },
                          })
                        }
                      />
                    </div>

                    <div className="grid gap-3 md:grid-cols-2">
                      <div className="space-y-1.5">
                        <Label htmlFor={`vm-target-${index}-retry-codes`}>{t("targets.retryCodes")}</Label>
                        <ChipInput
                          id={`vm-target-${index}-retry-codes`}
                          value={(target.policy.retry.status_codes ?? []).map(String)}
                          onValueChange={(next) =>
                            updateTarget(index, {
                              policy: {
                                ...target.policy,
                                retry: {
                                  ...target.policy.retry,
                                  status_codes: next.map(Number),
                                },
                              },
                            })
                          }
                          parseValues={(raw) => parseStatusCodes(raw).map(String)}
                          getRemoveLabel={(value) => `Remove status code ${value}`}
                          placeholder="408, 425, 429, 500, 502, 503, 504"
                        />
                      </div>
                      <div className="space-y-1.5">
                        <Label htmlFor={`vm-target-${index}-fallback-codes`}>{t("targets.fallbackCodes")}</Label>
                        <ChipInput
                          id={`vm-target-${index}-fallback-codes`}
                          value={(target.policy.fallback.status_codes ?? []).map(String)}
                          onValueChange={(next) =>
                            updateTarget(index, {
                              policy: {
                                ...target.policy,
                                fallback: {
                                  ...target.policy.fallback,
                                  status_codes: next.map(Number),
                                },
                              },
                            })
                          }
                          parseValues={(raw) => parseStatusCodes(raw).map(String)}
                          getRemoveLabel={(value) => `Remove status code ${value}`}
                          placeholder="429, 500, 502, 503, 504"
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              <div className="flex flex-col gap-3 rounded-2xl bg-muted/20 px-4 py-4">
                <div>
                  <p className="text-sm font-medium">{t("access.title")}</p>
                  <p className="text-xs text-muted-foreground">{t("access.description")}</p>
                </div>
                <div className="grid gap-3 lg:grid-cols-2">
                  <PrincipalPicker
                    id="vm-managers"
                    title={t("access.managers")}
                    selected={form.access.managers}
                    setSelected={(value) => setForm((prev) => ({ ...prev, access: { ...prev.access, managers: value } }))}
                    allowEveryone={false}
                    addLabel={t("access.add")}
                    removeLabel={t("access.remove")}
                    onSearchPrincipal={onSearchPrincipal}
                    principalSelectorLabels={principalSelectorLabels}
                  />
                  <PrincipalPicker
                    id="vm-users"
                    title={t("access.users")}
                    selected={form.access.users}
                    setSelected={(value) => setForm((prev) => ({ ...prev, access: { ...prev.access, users: value } }))}
                    allowEveryone
                    addLabel={t("access.add")}
                    removeLabel={t("access.remove")}
                    onSearchPrincipal={onSearchPrincipal}
                    principalSelectorLabels={principalSelectorLabels}
                  />
                </div>
              </div>

              <div className="rounded-2xl bg-muted/30 px-4 py-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium">{t("preview.title")}</p>
                    <p className="text-xs text-muted-foreground">
                      {form.routing_strategy === "weighted" ? t("preview.weightedDescription") : t("preview.description")}
                    </p>
                  </div>
                  {form.routing_strategy === "weighted" ? (
                    <Badge variant="secondary" className="rounded-full px-2.5 py-1 text-xs">
                      {t("targets.weightTotal", { total: totalWeight })}
                    </Badge>
                  ) : null}
                </div>
                <div className="mt-3 flex flex-col gap-2">
                  {form.targets.map((target) => (
                    <div
                      key={`${target.order}-${target.provider}-${target.model}`}
                      className="flex items-center justify-between gap-3 rounded-xl bg-background/90 px-3 py-2 ring-1 ring-border/30"
                    >
                      <div className="flex min-w-0 items-center gap-2">
                        <Badge variant="outline" className="shrink-0 rounded-full px-2">
                          {target.order}
                        </Badge>
                        <span className="truncate text-sm">
                          {target.provider || t("preview.providerPlaceholder")} / {target.model || t("preview.modelPlaceholder")}
                        </span>
                      </div>
                      {form.routing_strategy === "weighted" ? (
                        <Badge variant="secondary" className="shrink-0 rounded-full px-2">
                          {target.weight ?? 0}%
                        </Badge>
                      ) : null}
                    </div>
                  ))}
                </div>
              </div>

              <Alert className="rounded-2xl border-info/30 bg-info/8">
                <AlertTitle>{t("modes.title")}</AlertTitle>
                <AlertDescription className="flex flex-col gap-3">
                  <span>{t("modes.description")}</span>
                  {modeSummary.ready && modeSummary.intersection.length === 0 ? (
                    <span className="font-medium text-destructive">{t("modes.noCommonDescription")}</span>
                  ) : null}
                  {modeSummary.ready && modeSummary.intersection.length > 0 ? (
                    <span className="flex flex-wrap gap-2">
                      {modeSummary.intersection.map((mode) => {
                        const selected = effectiveExposedModes.includes(mode);
                        return (
                          <Button
                            key={mode}
                            type="button"
                            variant={selected ? "default" : "outline"}
                            size="sm"
                            onClick={() => {
                              setForm((prev) => {
                                const current =
                                  normalizeModeList(prev.exposed_modes).length === 0
                                    ? modeSummary.intersection
                                    : normalizeModeList(prev.exposed_modes);
                                const next = current.includes(mode) ? current.filter((item) => item !== mode) : [...current, mode];
                                return { ...prev, exposed_modes: next.length > 0 ? next : current };
                              });
                            }}
                          >
                            {mode}
                          </Button>
                        );
                      })}
                    </span>
                  ) : null}
                  {effectiveExposedModes.length > 0 ? (
                    <span className="flex flex-col gap-2">
                      {effectiveExposedModes.map((mode) => (
                        <span key={mode} className="flex flex-wrap items-center gap-2 text-xs">
                          <Badge variant="secondary">{mode}</Badge>
                          <span className="text-muted-foreground">{MODE_ENDPOINTS[mode] ?? t("modes.endpointUnknown")}</span>
                        </span>
                      ))}
                    </span>
                  ) : null}
                  {modeSummary.partialModes.length > 0 ? (
                    <span>{t("modes.partialDescription", { modes: modeSummary.partialModes.join(", ") })}</span>
                  ) : null}
                </AlertDescription>
              </Alert>
      </FormSheet>

      <UnsavedChangesDialog
        open={pendingDiscardAction !== null}
        onOpenChange={(open) => { if (!open) setPendingDiscardAction(null); }}
        title={t("unsavedDialog.title")}
        description={t("messages.unsavedConfirm")}
        stayLabel={t("unsavedDialog.stay")}
        discardLabel={t("unsavedDialog.discard")}
        onDiscard={async () => {
          const action = pendingDiscardAction;
          setPendingDiscardAction(null);
          await action?.();
        }}
      />
      <ConfirmDialog open={disableConfirmOpen} onOpenChange={setDisableConfirmOpen} title={t("actions.save")} description={t("messages.disableConfirm")} cancelLabel={t("actions.cancel")} confirmLabel={t("actions.save")} pendingLabel={t("actions.saving")} onConfirm={() => submitForm(true)} destructive={false} />

      <TypedConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={(open) => {
          if (isDeleting) return;
          setDeleteDialogOpen(open);
          if (!open) {
            setPendingDeleteItem(null);
            setDeleteNotice(null);
          }
        }}
        title={t("deleteDialog.title")}
        description={t("deleteDialog.description", { alias: pendingDeleteItem?.alias || "-" })}
        expectedValue={pendingDeleteItem?.alias ?? ""}
        inputLabel={t("deleteDialog.inputLabel")}
        warning={pendingDeleteItem ? <p className="text-xs text-muted-foreground">{t("deleteDialog.inputHint", { alias: pendingDeleteItem.alias })}</p> : undefined}
        cancelLabel={t("deleteDialog.cancel")}
        confirmLabel={t("deleteDialog.confirm")}
        pendingLabel={t("deleteDialog.deleting")}
        onConfirm={deleteCurrent}
        feedback={<OperationNotice notice={deleteNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />}
        disabled={!pendingDeleteItem?.id}
      />
    </>
  );
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

function formatStrategyLabel(
  value: string,
  t: ReturnType<typeof useTranslations<"virtualModels">>,
) {
  return value === "weighted" ? t("form.weightedValue") : t("form.priorityValue");
}

