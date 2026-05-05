"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { ArrowDown, ArrowUp, CheckCircle2, Loader2, Plus, RefreshCw, Save, Search, XCircle } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
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
  version: number;
  updated_at?: string;
  targets: VirtualModelTarget[];
};

type VMResponse = { data?: VirtualModel[] };
type VMDetail = VirtualModel;

type FormState = {
  id?: string;
  alias: string;
  routing_strategy: "priority";
  enabled: boolean;
  targets: VirtualModelTarget[];
};

type Notice = { kind: "success" | "error"; message: string } | null;

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

function makeDefaultForm(): FormState {
  return {
    alias: "",
    routing_strategy: "priority",
    enabled: true,
    targets: [{ ...DEFAULT_TARGET }],
  };
}

export default function VirtualModelManagement() {
  const t = useTranslations("virtualModels");
  const tc = useTranslations("common");

  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [query, setQuery] = useState("");
  const [notice, setNotice] = useState<Notice>(null);

  const [items, setItems] = useState<VirtualModel[]>([]);
  const [selectedID, setSelectedID] = useState<string>("");
  const [form, setForm] = useState<FormState>(makeDefaultForm());
  const [savedHash, setSavedHash] = useState("");

  const isCreate = !form.id;
  const dirty = useMemo(() => JSON.stringify(form) !== savedHash, [form, savedHash]);

  useEffect(() => {
    void loadList();
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

  async function loadList() {
    setIsLoading(true);
    setNotice(null);
    try {
      const res = await fetch("/api/virtual-models", { cache: "no-store" });
      const data = (await res.json().catch(() => ({}))) as VMResponse;
      if (!res.ok) throw new Error(readError(data, t("messages.listLoadError")));
      const rows = data.data ?? [];
      setItems(rows);
      if (!selectedID && rows.length > 0) {
        await loadDetail(rows[0].id);
      }
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.listLoadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function loadDetail(id: string) {
    if (!id) return;
    if (dirty) {
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
        targets: normalizeTargets(data.targets),
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
    if (!form.enabled && !window.confirm(t("messages.disableConfirm"))) {
      return;
    }
    setIsSaving(true);
    setNotice(null);
    const payload = {
      alias: form.alias.trim(),
      routing_strategy: form.routing_strategy,
      enabled: form.enabled,
      targets: form.targets.map((target, idx) => ({
        ...target,
        order: idx + 1,
      })),
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
      await loadList();
      if (nextID) await loadDetail(nextID);
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, isCreate ? t("messages.createError") : t("messages.updateError")) });
    } finally {
      setIsSaving(false);
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
            <div className="space-y-2">
              <div className="h-9 animate-pulse rounded-md bg-muted" />
              <div className="h-9 animate-pulse rounded-md bg-muted" />
              <div className="h-9 animate-pulse rounded-md bg-muted" />
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
              <p className="text-sm font-medium">{t("form.enabledLabel")}</p>
              <p className="text-xs text-muted-foreground">{t("form.enabledHelp")}</p>
            </div>
            <Button
              variant={form.enabled ? "default" : "secondary"}
              type="button"
              onClick={() => setForm((prev) => ({ ...prev, enabled: !prev.enabled }))}
            >
              {form.enabled ? t("status.enabled") : t("status.disabled")}
            </Button>
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
                    <Input value={target.provider} onChange={(event) => updateTarget(index, { provider: event.target.value })} />
                  </div>
                  <div className="space-y-1.5">
                    <Label>{t("targets.model")}</Label>
                    <Input value={target.model} onChange={(event) => updateTarget(index, { model: event.target.value })} />
                  </div>
                </div>

                <div className="grid gap-3 md:grid-cols-3">
                  <div className="space-y-1.5">
                    <Label>{t("targets.retryAttempts")}</Label>
                    <Input
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
                    <Label>{t("targets.retryDelay")}</Label>
                    <Input
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
                    <Label>{t("targets.timeout")}</Label>
                    <Input
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
                    <Label>{t("targets.retryCodes")}</Label>
                    <Input
                      value={(target.policy.retry.status_codes ?? []).join(",")}
                      onChange={(event) =>
                        updateTarget(index, {
                          policy: {
                            ...target.policy,
                            retry: {
                              ...target.policy.retry,
                              status_codes: parseStatusCodes(event.target.value),
                            },
                          },
                        })
                      }
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>{t("targets.fallbackCodes")}</Label>
                    <Input
                      value={(target.policy.fallback.status_codes ?? []).join(",")}
                      onChange={(event) =>
                        updateTarget(index, {
                          policy: {
                            ...target.policy,
                            fallback: {
                              ...target.policy.fallback,
                              status_codes: parseStatusCodes(event.target.value),
                            },
                          },
                        })
                      }
                    />
                  </div>
                </div>
              </div>
            ))}
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
            <Button type="button" onClick={() => void submitForm()} disabled={isSaving}>
              {isSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              {isCreate ? t("actions.create") : tc("save")}
            </Button>
            {!isCreate ? (
              <Button type="button" variant="outline" onClick={() => selectedID && void loadDetail(selectedID)}>
                {t("actions.reset")}
              </Button>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function parseStatusCodes(input: string): number[] {
  return input
    .split(",")
    .map((value) => Number(value.trim()))
    .filter((value) => Number.isFinite(value) && value >= 100 && value <= 599);
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

