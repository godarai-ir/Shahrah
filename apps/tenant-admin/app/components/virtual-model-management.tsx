"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { ArrowDown, ArrowUp, CheckCircle2, Loader2, Plus, RefreshCw, Save, Search, Trash2, XCircle } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ModelEntitySelect, type ModelSelection } from "@/app/components/model-entity-select";
import { PrincipalPicker, type PrincipalPickerItem } from "@/app/components/principal-picker";
import { type PrincipalEntitySelectLabels, type PrincipalSearchKind } from "@/app/components/principal-search-input";
import { cn } from "@/lib/utils";

type VirtualModelPolicy = {
  timeout_sec?: number;
  retry: { attempts: number; delay_ms: number; status_codes?: number[] };
  fallback: { status_codes?: number[]; candidate: boolean };
};

type VirtualModelTarget = {
  id?: string;
  order: number;
  provider: string;
  model: string;
  policy: VirtualModelPolicy;
};

type VirtualModel = {
  id: string;
  alias: string;
  routing_strategy: string;
  enabled: boolean;
  exposed_modes?: string[];
  version: number;
  updated_at?: string;
  targets: VirtualModelTarget[];
  access?: {
    managers?: PrincipalPickerItem[];
    users?: PrincipalPickerItem[];
  };
};

type VMResponse = { data?: VirtualModel[] };
type VMDetail = VirtualModel;
type TenantProviderAccount = { id: string; provider: string; name: string };
type TenantProviderAccountModel = { model: string; enabled: boolean; types?: string[]; effective_capabilities?: string[]; supported_modes?: string[] };
type AccountsResponse = { data?: TenantProviderAccount[] };
type AccountDetailsResponse = {
  account?: TenantProviderAccount;
  models?: TenantProviderAccountModel[];
  data?: {
    account?: TenantProviderAccount;
    models?: TenantProviderAccountModel[];
  };
};

type FormState = {
  id?: string;
  alias: string;
  routing_strategy: "priority";
  enabled: boolean;
  exposed_modes: string[];
  targets: VirtualModelTarget[];
  access: {
    managers: PrincipalPickerItem[];
    users: PrincipalPickerItem[];
  };
};
type AccessPrincipalCandidate = { principal_id: string; display_name: string; secondary?: string; type?: PrincipalSearchKind };
type AccessPrincipalResponse = { data?: AccessPrincipalCandidate[] };

type Notice = { kind: "success" | "error"; message: string } | null;
type StatusCodeChipsInputProps = {
  id: string;
  value: number[];
  onChange: (next: number[]) => void;
  placeholder?: string;
  disabled?: boolean;
};

const DEFAULT_TARGET: VirtualModelTarget = {
  order: 1,
  provider: "",
  model: "",
  policy: {
    timeout_sec: 60,
    retry: {
      attempts: 2,
      delay_ms: 100,
      status_codes: [429, 500, 502, 503],
    },
    fallback: {
      status_codes: [429, 500, 502, 503, 504],
      candidate: true,
    },
  },
};

const MODE_ENDPOINTS: Record<string, string> = {
  chat: "/v1/chat/completions, /v1/responses",
  completion: "/v1/completions",
  embedding: "/v1/embeddings",
  moderation: "/v1/moderations",
  image_generation: "/v1/images/generations",
  text_to_speech: "/v1/audio/speech",
  audio_transcription: "/v1/audio/transcriptions",
  audio_translation: "/v1/audio/translations",
  realtime: "/v1/realtime",
  rerank: "/v1/rerank",
  search: "/v1/search",
};

function makeDefaultForm(): FormState {
  return {
    alias: "",
    routing_strategy: "priority",
    enabled: true,
    exposed_modes: [],
    targets: [{ ...DEFAULT_TARGET }],
    access: {
      managers: [],
      users: [{ type: "everyone", display_name: "Everyone" }],
    },
  };
}

export default function VirtualModelManagement() {
  const t = useTranslations("virtualModels");
  const tc = useTranslations("common");

  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [notice, setNotice] = useState<Notice>(null);

  const [items, setItems] = useState<VirtualModel[]>([]);
  const [providerAllowedModels, setProviderAllowedModels] = useState<ModelSelection[]>([]);
  const [providerModelModes, setProviderModelModes] = useState<Record<string, string[]>>({});
  const [selectedID, setSelectedID] = useState<string>("");
  const [form, setForm] = useState<FormState>(() => makeDefaultForm());
  const [savedHash, setSavedHash] = useState(() => JSON.stringify(makeDefaultForm()));

  const isCreate = !form.id;
  const dirty = useMemo(() => JSON.stringify(form) !== savedHash, [form, savedHash]);
  const modeSummary = useMemo(() => buildModeSummary(form.targets, providerModelModes), [form.targets, providerModelModes]);
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

  const principalSelectorLabels = useMemo<PrincipalEntitySelectLabels>(() => ({
    users: "Users",
    teams: "Teams",
    orgUnits: "Org Units",
    virtualAccounts: "Virtual Accounts",
    externalIdentities: "External Identities",
    everyone: "Everyone",
    searchPlaceholder: "Search principals...",
    triggerPlaceholder: "Select principal",
    loading: "Loading...",
    empty: "No results",
  }), []);

  async function loadList(options?: { skipDirtyConfirmOnAutoload?: boolean }) {
    setIsLoading(true);
    setNotice(null);
    try {
      const res = await fetch("/api/virtual-models", { cache: "no-store" });
      const data = (await res.json().catch(() => ({}))) as VMResponse;
      if (!res.ok) throw new Error(readError(data, t("messages.listLoadError")));
      const rows = data.data ?? [];
      setItems(rows);
      if (!selectedID && rows.length > 0) {
        await loadDetail(rows[0].id, { skipDirtyConfirm: options?.skipDirtyConfirmOnAutoload });
      }
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.listLoadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadDetail(id: string, options?: { skipDirtyConfirm?: boolean }) {
    if (!id) return;
    if (!options?.skipDirtyConfirm && dirty) {
      const confirmed = window.confirm(t("messages.unsavedConfirm"));
      if (!confirmed) return;
    }
    setIsLoading(true);
    setNotice(null);
    try {
      const res = await fetch(`/api/virtual-models/${encodeURIComponent(id)}`, { cache: "no-store" });
      const data = (await res.json().catch(() => ({}))) as VMDetail;
      if (!res.ok) throw new Error(readError(data, t("messages.detailLoadError")));
      const mapped: FormState = {
        id: data.id,
        alias: data.alias,
        routing_strategy: "priority",
        enabled: data.enabled,
        exposed_modes: normalizeModeList(data.exposed_modes),
        targets: normalizeTargets(data.targets),
        access: normalizeAccess(data.access),
      };
      setSelectedID(id);
      setForm(mapped);
      setSavedHash(JSON.stringify(mapped));
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.detailLoadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadProviderAllowedModels() {
    try {
      const accountsRes = await fetch("/api/tenant-provider-accounts", { cache: "no-store" });
      const accountsData = (await accountsRes.json().catch(() => ({}))) as AccountsResponse;
      if (!accountsRes.ok) throw new Error(readError(accountsData, t("messages.listLoadError")));
      const accounts = accountsData.data ?? [];
      const allSelections = await Promise.all(
        accounts.map(async (account) => {
          const detailRes = await fetch(`/api/tenant-provider-accounts/${encodeURIComponent(account.id)}`, { cache: "no-store" });
          const detailData = (await detailRes.json().catch(() => ({}))) as AccountDetailsResponse;
          if (!detailRes.ok) return [] as ModelSelection[];
          const payload = detailData.data ?? detailData;
          const models = payload.models ?? [];
          return models
            .filter((item) => item.enabled)
            .map<ModelSelection>((item) => ({
              id: `${account.id}:${item.model}`,
              name: `${account.provider}:${account.name}:${item.model}`,
              kind: "provider_allowed",
              secondary: `${account.provider}/${account.name}`,
            }));
        }),
      );
      setProviderAllowedModels(allSelections.flat());
      const modeEntries = await Promise.all(
        accounts.map(async (account) => {
          const detailRes = await fetch(`/api/tenant-provider-accounts/${encodeURIComponent(account.id)}`, { cache: "no-store" });
          const detailData = (await detailRes.json().catch(() => ({}))) as AccountDetailsResponse;
          if (!detailRes.ok) return [] as Array<[string, string[]]>;
          const payload = detailData.data ?? detailData;
          return (payload.models ?? [])
            .filter((item) => item.enabled)
            .flatMap((item): Array<[string, string[]]> => {
              const rawFQN = `${account.provider}:${account.name}:${item.model}`;
              const modes = modesForProviderModel(item);
              return [
                [rawFQN, modes],
                [canonicalModelFQN(rawFQN), modes],
              ];
            });
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
      const confirmed = window.confirm(t("messages.unsavedConfirm"));
      if (!confirmed) return;
    }
    const next = makeDefaultForm();
    setSelectedID("");
    setForm(next);
    setSavedHash(JSON.stringify(next));
    setNotice(null);
  }

  function updateTarget(index: number, patch: Partial<VirtualModelTarget>) {
    setForm((prev) => {
      const targets = prev.targets.map((item, i) => (i === index ? { ...item, ...patch } : item));
      return { ...prev, targets };
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
    setForm((prev) => ({
      ...prev,
      targets: [
        ...prev.targets,
        {
          ...DEFAULT_TARGET,
          order: prev.targets.length + 1,
        },
      ],
    }));
  }

  function removeTarget(index: number) {
    setForm((prev) => {
      if (prev.targets.length <= 1) return prev;
      const targets = prev.targets.filter((_, i) => i !== index).map((item, idx) => ({ ...item, order: idx + 1 }));
      return { ...prev, targets };
    });
  }

  async function submitForm() {
    const validationError = validateForm(form, t);
    if (validationError) {
      setNotice({ kind: "error", message: validationError });
      return;
    }
    if (modeSummary.ready && modeSummary.intersection.length === 0) {
      setNotice({ kind: "error", message: t("messages.noCommonModes") });
      return;
    }
    if (!form.enabled && !window.confirm(t("messages.disableConfirm"))) {
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
      const url = form.id ? `/api/virtual-models/${encodeURIComponent(form.id)}` : "/api/virtual-models";
      const method = form.id ? "PATCH" : "POST";
      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = (await res.json().catch(() => ({}))) as Partial<VirtualModel>;
      if (!res.ok) throw new Error(readError(data, isCreate ? t("messages.createError") : t("messages.updateError")));
      const nextID = data.id ?? form.id ?? "";
      setNotice({ kind: "success", message: isCreate ? t("messages.createSuccess") : t("messages.updateSuccess") });
      // The current form has just been persisted; prevent unsaved confirmation
      // while hydrating the latest server state back into the form.
      setSavedHash(JSON.stringify(form));
      await loadList({ skipDirtyConfirmOnAutoload: true });
      if (nextID) await loadDetail(nextID, { skipDirtyConfirm: true });
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, isCreate ? t("messages.createError") : t("messages.updateError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function onSearchPrincipal(kind: PrincipalSearchKind, query: string) {
    if (query.trim().length < 2) return [];
    const response = await fetch(
      `/api/access-principals?kind=${kind}&q=${encodeURIComponent(query.trim())}&limit=20`,
      { cache: "no-store" },
    );
    const payload = (await response.json().catch(() => ({}))) as AccessPrincipalResponse;
    if (!response.ok) return [];
    return payload.data ?? [];
  }

  async function deleteCurrent() {
    if (!form.id) return;

    setIsDeleting(true);
    setNotice(null);
    try {
      const res = await fetch(`/api/virtual-models/${encodeURIComponent(form.id)}`, { method: "DELETE" });
      const data = (await res.json().catch(() => ({}))) as Partial<VirtualModel>;
      if (!res.ok) throw new Error(readError(data, t("messages.deleteError")));

      const next = makeDefaultForm();
      setSelectedID("");
      setForm(next);
      setSavedHash(JSON.stringify(next));
      setNotice({ kind: "success", message: t("messages.deleteSuccess") });
      setDeleteDialogOpen(false);
      await loadList({ skipDirtyConfirmOnAutoload: true });
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.deleteError")) });
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <div className="grid gap-4 lg:grid-cols-[1.05fr_1.6fr]">
      <Card className="overflow-hidden">
        <CardHeader className="space-y-3">
          <div>
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>{t("description")}</CardDescription>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row">
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute top-3 left-3 h-4 w-4 text-muted-foreground" />
              <Input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                className="pl-9"
                placeholder={t("searchPlaceholder")}
                aria-label={t("searchPlaceholder")}
              />
            </div>
            <Button variant="outline" onClick={() => void loadList()} disabled={isLoading}>
              <RefreshCw className={cn("h-4 w-4", isLoading && "animate-spin")} />
              {t("actions.refresh")}
            </Button>
            <Button onClick={startCreate}>
              <Plus className="h-4 w-4" />
              {t("actions.new")}
            </Button>
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
            <div className="rounded-lg border border-dashed p-5 text-sm text-muted-foreground">
              <p className="font-medium text-foreground">{t("empty.title")}</p>
              <p className="mt-1">{t("empty.description")}</p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("list.alias")}</TableHead>
                    <TableHead>{t("list.strategy")}</TableHead>
                    <TableHead>{t("list.targets")}</TableHead>
                    <TableHead>{t("list.status")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.map((item) => (
                    <TableRow
                      key={item.id}
                      className={cn("cursor-pointer", selectedID === item.id && "bg-muted/60")}
                      onClick={() => void loadDetail(item.id)}
                    >
                      <TableCell className="font-medium">{item.alias}</TableCell>
                      <TableCell className="capitalize">{item.routing_strategy}</TableCell>
                      <TableCell>{item.targets?.length ?? 0}</TableCell>
                      <TableCell>
                        <Badge variant={item.enabled ? "default" : "secondary"}>
                          {item.enabled ? t("status.enabled") : t("status.disabled")}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{isCreate ? t("form.createTitle") : t("form.editTitle")}</CardTitle>
          <CardDescription>{isCreate ? t("form.createDescription") : t("form.editDescription")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {notice ? (
            <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
              {notice.kind === "error" ? <XCircle className="h-4 w-4" /> : <CheckCircle2 className="h-4 w-4" />}
              <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
              <AlertDescription>{notice.message}</AlertDescription>
            </Alert>
          ) : null}

          <div className="grid gap-3 md:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor="vm-alias">{t("form.alias")}</Label>
              <Input
                id="vm-alias"
                value={form.alias}
                onChange={(event) => setForm((prev) => ({ ...prev, alias: event.target.value }))}
                placeholder={t("form.aliasPlaceholder")}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="vm-strategy">{t("form.strategy")}</Label>
              <Input id="vm-strategy" value={t("form.priorityValue")} readOnly aria-readonly />
            </div>
          </div>

          <div className="flex items-center justify-between rounded-lg border p-3">
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

          <div className="space-y-3 rounded-lg border p-3">
            <div className="flex items-center justify-between">
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
              <div key={`${target.id ?? "new"}-${index}`} className="space-y-3 rounded-lg border p-3">
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

                <div className="grid gap-3 md:grid-cols-2">
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
                  <div className="space-y-1.5">
                    <Label htmlFor={`vm-target-${index}-retry-attempts`}>{t("targets.retryAttempts")}</Label>
                    <Input
                      id={`vm-target-${index}-retry-attempts`}
                      type="number"
                      min={0}
                      max={5}
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
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor={`vm-target-${index}-retry-delay`}>{t("targets.retryDelay")}</Label>
                    <Input
                      id={`vm-target-${index}-retry-delay`}
                      type="number"
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
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor={`vm-target-${index}-timeout`}>{t("targets.timeout")}</Label>
                    <Input
                      id={`vm-target-${index}-timeout`}
                      type="number"
                      min={1}
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
                </div>

                <div className="grid gap-3 md:grid-cols-2">
                  <div className="space-y-1.5">
                    <Label htmlFor={`vm-target-${index}-retry-codes`}>{t("targets.retryCodes")}</Label>
                    <StatusCodeChipsInput
                      id={`vm-target-${index}-retry-codes`}
                      value={target.policy.retry.status_codes ?? []}
                      onChange={(next) =>
                        updateTarget(index, {
                          policy: {
                            ...target.policy,
                            retry: {
                              ...target.policy.retry,
                              status_codes: next,
                            },
                          },
                        })
                      }
                      placeholder="429, 500, 502, 503"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor={`vm-target-${index}-fallback-codes`}>{t("targets.fallbackCodes")}</Label>
                    <StatusCodeChipsInput
                      id={`vm-target-${index}-fallback-codes`}
                      value={target.policy.fallback.status_codes ?? []}
                      onChange={(next) =>
                        updateTarget(index, {
                          policy: {
                            ...target.policy,
                            fallback: {
                              ...target.policy.fallback,
                              status_codes: next,
                            },
                          },
                        })
                      }
                      placeholder="429, 500, 502, 503, 504"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="flex flex-col gap-3 rounded-lg border p-3">
            <div>
              <p className="text-sm font-medium">{t("modes.title")}</p>
              <p className="text-xs text-muted-foreground">{t("modes.description")}</p>
            </div>
            {modeSummary.ready && modeSummary.intersection.length === 0 ? (
              <Alert variant="destructive">
                <XCircle className="h-4 w-4" />
                <AlertTitle>{t("modes.noCommonTitle")}</AlertTitle>
                <AlertDescription>{t("modes.noCommonDescription")}</AlertDescription>
              </Alert>
            ) : modeSummary.ready ? (
              <div className="flex flex-col gap-3">
                <div className="flex flex-wrap gap-2">
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
                            const current = normalizeModeList(prev.exposed_modes).length === 0 ? modeSummary.intersection : normalizeModeList(prev.exposed_modes);
                            const next = current.includes(mode) ? current.filter((item) => item !== mode) : [...current, mode];
                            return { ...prev, exposed_modes: next.length > 0 ? next : current };
                          });
                        }}
                      >
                        {mode}
                      </Button>
                    );
                  })}
                </div>
                <div className="flex flex-col gap-2">
                  {effectiveExposedModes.map((mode) => (
                    <div key={mode} className="flex flex-wrap items-center gap-2 text-xs">
                      <Badge variant="secondary">{mode}</Badge>
                      <span className="text-muted-foreground">{MODE_ENDPOINTS[mode] ?? t("modes.endpointUnknown")}</span>
                    </div>
                  ))}
                </div>
              </div>
            ) : null}
            {modeSummary.partialModes.length > 0 ? (
              <Alert>
                <AlertTitle>{t("modes.partialTitle")}</AlertTitle>
                <AlertDescription>
                  {t("modes.partialDescription", { modes: modeSummary.partialModes.join(", ") })}
                </AlertDescription>
              </Alert>
            ) : null}
          </div>

          <div className="rounded-lg border bg-muted/30 p-3">
            <p className="text-sm font-medium">{t("preview.title")}</p>
            <p className="text-xs text-muted-foreground">{t("preview.description")}</p>
            <div className="mt-2 flex flex-wrap items-center gap-2 text-xs">
              {form.targets.map((target) => (
                <span key={`${target.order}-${target.provider}-${target.model}`} className="rounded-full border px-2 py-1">
                  {target.order}. {target.provider || t("preview.providerPlaceholder")} / {target.model || t("preview.modelPlaceholder")}
                </span>
              ))}
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Button type="button" onClick={() => void submitForm()} disabled={isSaving || isDeleting}>
              {isSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              {isCreate ? t("actions.create") : tc("save")}
            </Button>
            {!isCreate ? (
              <Button type="button" variant="outline" onClick={() => selectedID && void loadDetail(selectedID)}>
                {t("actions.reset")}
              </Button>
            ) : null}
            {!isCreate ? (
              <Button type="button" variant="destructive" onClick={() => setDeleteDialogOpen(true)} disabled={isSaving || isDeleting}>
                {isDeleting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
                {t("actions.delete")}
              </Button>
            ) : null}
          </div>

          <div className="space-y-3 rounded-lg border p-3">
            <div>
              <p className="text-sm font-medium">Access Control</p>
              <p className="text-xs text-muted-foreground">Define who can manage and use this virtual model.</p>
            </div>
            <div className="grid gap-3 lg:grid-cols-2">
              <PrincipalPicker
                id="vm-managers"
                title="Managers"
                selected={form.access.managers}
                setSelected={(value) => setForm((prev) => ({ ...prev, access: { ...prev.access, managers: value } }))}
                allowEveryone={false}
                addLabel="Add"
                removeLabel="Remove"
                onSearchPrincipal={onSearchPrincipal}
                principalSelectorLabels={principalSelectorLabels}
              />
              <PrincipalPicker
                id="vm-users"
                title="Users"
                selected={form.access.users}
                setSelected={(value) => setForm((prev) => ({ ...prev, access: { ...prev.access, users: value } }))}
                allowEveryone
                addLabel="Add"
                removeLabel="Remove"
                onSearchPrincipal={onSearchPrincipal}
                principalSelectorLabels={principalSelectorLabels}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <Dialog
        open={deleteDialogOpen}
        onOpenChange={(open) => {
          if (isDeleting) return;
          setDeleteDialogOpen(open);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteDialog.title")}</DialogTitle>
            <DialogDescription>
              {t("deleteDialog.description", { alias: form.alias || form.id || "-" })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDeleteDialogOpen(false)} disabled={isDeleting}>
              {t("deleteDialog.cancel")}
            </Button>
            <Button type="button" variant="destructive" onClick={() => void deleteCurrent()} disabled={isDeleting || !form.id}>
              {isDeleting ? t("deleteDialog.deleting") : t("deleteDialog.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function StatusCodeChipsInput(props: StatusCodeChipsInputProps) {
  const [draft, setDraft] = useState("");

  function addCodes(raw: string) {
    const codes = parseStatusCodes(raw);
    if (codes.length === 0) return;
    const seen = new Set(props.value);
    const next = [...props.value];
    for (const code of codes) {
      if (seen.has(code)) continue;
      seen.add(code);
      next.push(code);
    }
    props.onChange(next);
  }

  function commitDraft() {
    if (!draft.trim()) return;
    addCodes(draft);
    setDraft("");
  }

  function removeCode(code: number) {
    props.onChange(props.value.filter((item) => item !== code));
  }

  return (
    <div className="rounded-md border bg-background px-2 py-2">
      <div className="flex flex-wrap items-center gap-2">
        {props.value.map((code) => (
          <Badge key={code} variant="secondary" className="gap-1 pe-1">
            <span>{code}</span>
            <button
              type="button"
              className="rounded-sm p-0.5 text-muted-foreground transition hover:bg-accent hover:text-foreground"
              onClick={() => removeCode(code)}
              aria-label={`Remove status code ${code}`}
              disabled={props.disabled}
            >
              <XCircle className="h-3.5 w-3.5" />
            </button>
          </Badge>
        ))}
        <Input
          id={props.id}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onBlur={commitDraft}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === ",") {
              event.preventDefault();
              commitDraft();
              return;
            }
            if (event.key === "Backspace" && !draft && props.value.length > 0) {
              removeCode(props.value[props.value.length - 1]);
            }
          }}
          placeholder={props.placeholder}
          className="h-8 min-w-[120px] flex-1 border-0 p-0 ps-1 shadow-none focus-visible:ring-0"
          disabled={props.disabled}
        />
      </div>
    </div>
  );
}

function parseStatusCodes(input: string): number[] {
  return input
    .split(/[,\s]+/)
    .map((value) => Number(value.trim()))
    .filter((value) => Number.isInteger(value) && value >= 100 && value <= 599);
}

function readError(payload: unknown, fallback: string) {
  if (!payload || typeof payload !== "object") return fallback;
  const data = payload as { error?: { message?: string }; message?: string };
  return data.error?.message ?? data.message ?? fallback;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

function normalizeTargets(targets: VirtualModelTarget[] | undefined): VirtualModelTarget[] {
  if (!targets || targets.length === 0) return [{ ...DEFAULT_TARGET }];
  return targets
    .map((target, idx) => ({
      ...target,
      order: target.order || idx + 1,
      policy: {
        timeout_sec: target.policy?.timeout_sec ?? 60,
        retry: {
          attempts: target.policy?.retry?.attempts ?? 2,
          delay_ms: target.policy?.retry?.delay_ms ?? 100,
          status_codes: target.policy?.retry?.status_codes ?? [429, 500, 502, 503],
        },
        fallback: {
          candidate: target.policy?.fallback?.candidate ?? true,
          status_codes: target.policy?.fallback?.status_codes ?? [429, 500, 502, 503, 504],
        },
      },
    }))
    .sort((a, b) => a.order - b.order);
}

function normalizeAccess(access: VirtualModel["access"] | undefined): FormState["access"] {
  const managers = (access?.managers ?? []).filter(Boolean);
  const users = (access?.users ?? []).filter(Boolean);
  if (managers.length === 0 && users.length === 0) {
    return {
      managers: [],
      users: [{ type: "everyone", display_name: "Everyone" }],
    };
  }
  return {
    managers,
    users,
  };
}

function validateForm(form: FormState, t: ReturnType<typeof useTranslations<"virtualModels">>) {
  if (!form.alias.trim()) return t("messages.aliasRequired");
  if (!/^[a-zA-Z0-9_-]+$/.test(form.alias.trim())) return t("messages.aliasPattern");
  if (form.targets.length === 0) return t("messages.targetsRequired");
  const seenOrders = new Set<number>();
  for (const target of form.targets) {
    if (!target.provider.trim()) return t("messages.providerRequired");
    if (!target.model.trim()) return t("messages.modelRequired");
    if (seenOrders.has(target.order)) return t("messages.duplicateOrder");
    seenOrders.add(target.order);
    if (target.policy.retry.attempts < 0 || target.policy.retry.attempts > 5) return t("messages.retryAttemptsRange");
    if (target.policy.retry.delay_ms < 0 || target.policy.retry.delay_ms > 10000) return t("messages.retryDelayRange");
  }
  return null;
}

function modesForProviderModel(model: TenantProviderAccountModel): string[] {
  return normalizeModeList(model.supported_modes ?? model.effective_capabilities ?? model.types ?? []);
}

function normalizeModeList(values: string[] | undefined): string[] {
  const seen = new Set<string>();
  const out: string[] = [];
  for (const value of values ?? []) {
    const mode = value.trim();
    if (!mode || seen.has(mode)) continue;
    seen.add(mode);
    out.push(mode);
  }
  return out;
}

function buildModeSummary(targets: VirtualModelTarget[], modeMap: Record<string, string[]>) {
  const selectedTargets = targets.filter((target) => target.model.trim());
  if (selectedTargets.length === 0) {
    return { intersection: [] as string[], partialModes: [] as string[], ready: false };
  }
  const targetModes = selectedTargets.map((target) => normalizeModeList(modeMap[target.model] ?? modeMap[canonicalModelFQN(target.model)]));
  if (targetModes.some((modes) => modes.length === 0)) {
    return { intersection: [] as string[], partialModes: [] as string[], ready: true };
  }
  const intersection = targetModes.reduce((acc, modes) => acc.filter((mode) => modes.includes(mode)));
  const union = normalizeModeList(targetModes.flat());
  return {
    intersection,
    partialModes: union.filter((mode) => !intersection.includes(mode)),
    ready: true,
  };
}

function sameModelFQN(left: string, right: string): boolean {
  return canonicalModelFQN(left) === canonicalModelFQN(right);
}

function canonicalModelFQN(value: string): string {
  const parts = value.trim().split(":");
  if (parts.length !== 3) return value.trim();
  const provider = parts[0].trim();
  const account = parts[1].trim();
  let model = parts[2].trim();
  const prefix = `${provider.toLowerCase()}/`;
  if (model.toLowerCase().startsWith(prefix)) {
    model = model.slice(prefix.length).trim();
  }
  if (!provider || !account || !model) return value.trim();
  return `${provider}:${account}:${model}`;
}
