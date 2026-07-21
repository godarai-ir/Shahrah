"use client";

import { AlertTriangle, FlaskConical, FolderOpen, Play, Plus, Save, Square, Variable } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";

import { usePermissions } from "@/app/components/permissions-context";
import { normalizeError, requestJSON, unwrapData } from "@/app/components/provider-account-wizard/utils";
import {
  AppTextField,
  FormDialog,
  OperationNotice,
  ResizableSplitPane,
  SearchPickerDialog,
  UnsavedChangesDialog,
} from "@/components/app-ui";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Spinner } from "@/components/ui/spinner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

import {
  buildPlaygroundManifest,
  clonePlaygroundColumn,
  createPlaygroundColumn,
  getColumnVariables,
  hydratePlaygroundColumn,
  validateAndBuildRunPayload,
  type PlaygroundColumnDraft,
  type PlaygroundRunResult,
  type PlaygroundTelemetry,
} from "./playground-domain";
import { PlaygroundComparisonColumn, PlaygroundEditorPanel, PlaygroundOutputPanel } from "./playground-column";
import { PlaygroundSettingsSheet } from "./playground-settings-sheet";
import { extractPlaygroundText, usePlaygroundStream } from "./use-playground-stream";
import type {
  GatewayModel,
  GatewayModelListResponse,
  GuardrailItem,
  GuardrailListResponse,
  Prompt,
  PromptListResponse,
  PromptVersion,
  PromptVersionListResponse,
} from "./prompt-management-shared";

type Notice = { kind: "success" | "error" | "warning" | "info"; message: string } | null;
type PlaygroundTranslation = (key: string, values?: Record<string, string | number>) => string;

export default function PlaygroundPage() {
  const t = useTranslations("playground");
  const locale = useLocale();
  const searchParams = useSearchParams();
  const permissions = usePermissions() ?? [];
  const canRun = permissions.some((item) => item === "playground:run" || item === "policy:manage");
  const canSave = permissions.some((item) => item === "prompt:manage" || item === "policy:manage");
  const canLoad = permissions.some((item) => item === "prompt:list" || item === "policy:list" || item === "policy:manage");
  const canOpenTrace = permissions.includes("settings:list");
  const consumeStream = usePlaygroundStream();

  const [columns, setColumns] = useState<PlaygroundColumnDraft[]>(() => [createPlaygroundColumn(t("option", { count: 1 }), locale === "fa" ? "شما یک دستیار دقیق و مفید هستید." : "You are a precise and helpful assistant.")]);
  const [results, setResults] = useState<Record<string, PlaygroundRunResult>>(() => ({ [columns[0].id]: emptyRunResult() }));
  const [activeColumnID, setActiveColumnID] = useState(columns[0].id);
  const [models, setModels] = useState<GatewayModel[]>([]);
  const [modelsLoading, setModelsLoading] = useState(true);
  const [modelError, setModelError] = useState("");
  const [guardrails, setGuardrails] = useState<GuardrailItem[]>([]);
  const [guardrailsLoading, setGuardrailsLoading] = useState(true);
  const [notice, setNotice] = useState<Notice>(null);
  const [modelPickerOpen, setModelPickerOpen] = useState(false);
  const [modelSearch, setModelSearch] = useState("");
  const [settingsColumnID, setSettingsColumnID] = useState<string | null>(null);
  const [variablesOpen, setVariablesOpen] = useState(false);
  const [saveColumnID, setSaveColumnID] = useState<string | null>(null);
  const [loadOpen, setLoadOpen] = useState(false);
  const [discardOpen, setDiscardOpen] = useState(false);
  const [pendingVersion, setPendingVersion] = useState<PromptVersion | null>(null);
  const [pendingHref, setPendingHref] = useState<string | null>(null);
  const abortControllers = useRef(new Map<string, AbortController>());
  const confirmedNavigation = useRef(false);
  const initialVersionLoaded = useRef(false);
  const initialModelLoaded = useRef(false);

  const activeColumn = columns.find((column) => column.id === activeColumnID) ?? columns[0];
  const settingsColumn = columns.find((column) => column.id === settingsColumnID) ?? null;
  const allVariables = useMemo(() => Array.from(new Set(columns.flatMap(getColumnVariables))).sort((a, b) => a.localeCompare(b)), [columns]);
  const anyRunning = Object.values(results).some((result) => result.status === "running");

  useEffect(() => {
    void Promise.all([loadModels(), loadGuardrails()]);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const versionID = searchParams.get("versionID");
    if (!versionID || initialVersionLoaded.current) return;
    initialVersionLoaded.current = true;
    void requestJSON<{ data?: PromptVersion } | PromptVersion>(`/api/prompt-versions/${encodeURIComponent(versionID)}`)
      .then((payload) => {
        const version = unwrapData<PromptVersion>(payload);
        if (version) replaceActiveWithVersion(version, true);
      })
      .catch((error) => setNotice({ kind: "error", message: normalizeError(error, t("errors.load")) }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  useEffect(() => {
    function beforeUnload(event: BeforeUnloadEvent) {
      if (!confirmedNavigation.current && columns.some((column) => column.dirty)) event.preventDefault();
    }
    function beforeLinkNavigation(event: MouseEvent) {
      if (!columns.some((column) => column.dirty) || event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
      const target = event.target instanceof Element ? event.target.closest<HTMLAnchorElement>("a[href]") : null;
      if (!target || target.download || (target.target && target.target !== "_self")) return;
      const destination = new URL(target.href, window.location.href);
      if (destination.href === window.location.href) return;
      event.preventDefault();
      setPendingVersion(null);
      setPendingHref(destination.href);
      setDiscardOpen(true);
    }
    window.addEventListener("beforeunload", beforeUnload);
    document.addEventListener("click", beforeLinkNavigation, true);
    return () => {
      window.removeEventListener("beforeunload", beforeUnload);
      document.removeEventListener("click", beforeLinkNavigation, true);
    };
  }, [columns]);

  useEffect(() => {
    function onShortcut(event: KeyboardEvent) {
      if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
        event.preventDefault();
        if (!canRun || anyRunning) return;
        if (columns.length > 1) void runAll();
        else void runColumn(columns[0]);
      }
    }
    window.addEventListener("keydown", onShortcut);
    return () => window.removeEventListener("keydown", onShortcut);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [canRun, anyRunning, columns]);

  async function loadModels() {
    setModelsLoading(true);
    setModelError("");
    try {
      const payload = await requestJSON<GatewayModelListResponse>("/api/models");
      const nextModels = payload.data ?? [];
      setModels(nextModels);
      const requestedModel = !initialModelLoaded.current && !searchParams.get("versionID")
        ? nextModels.find((model) => model.id === searchParams.get("model"))
        : undefined;
      initialModelLoaded.current = true;
      setColumns((current) => current.map((column, index) => {
        const selectedModel = index === 0 && requestedModel
          ? requestedModel.id
          : column.selectedModel || (nextModels.length === 1 ? nextModels[0].id : "");
        return { ...column, selectedModel, selectedProvider: nextModels.find((model) => model.id === selectedModel)?.provider ?? column.selectedProvider };
      }));
    } catch (error) {
      setModelError(normalizeError(error, t("errors.loadModels")));
    } finally {
      setModelsLoading(false);
    }
  }

  async function loadGuardrails() {
    setGuardrailsLoading(true);
    try {
      const payload = await requestJSON<GuardrailListResponse>("/api/guardrails?limit=200");
      setGuardrails(payload.data ?? []);
    } catch {
      setGuardrails([]);
    } finally {
      setGuardrailsLoading(false);
    }
  }

  function updateColumn(next: PlaygroundColumnDraft) {
    setColumns((current) => current.map((column) => column.id === next.id ? next : column));
  }

  function addComparison() {
    if (columns.length >= 3 || !activeColumn) return;
    const next = clonePlaygroundColumn(activeColumn, t("option", { count: columns.length + 1 }));
    setColumns((current) => [...current, next]);
    setResults((current) => ({ ...current, [next.id]: emptyRunResult() }));
    setActiveColumnID(next.id);
  }

  function duplicateColumn(column: PlaygroundColumnDraft) {
    if (columns.length >= 3) return;
    const next = clonePlaygroundColumn(column, t("option", { count: columns.length + 1 }));
    setColumns((current) => [...current, next]);
    setResults((current) => ({ ...current, [next.id]: emptyRunResult() }));
    setActiveColumnID(next.id);
  }

  function removeColumn(columnID: string) {
    if (columns.length === 1) return;
    abortControllers.current.get(columnID)?.abort();
    const nextColumns = columns.filter((column) => column.id !== columnID);
    setColumns(nextColumns);
    setResults((current) => Object.fromEntries(Object.entries(current).filter(([key]) => key !== columnID)));
    if (activeColumnID === columnID) setActiveColumnID(nextColumns[0].id);
  }

  async function runAll() {
    const comparisonID = createClientID("comparison");
    const invalid = columns.map((column) => ({ column, validation: validateAndBuildRunPayload(column, comparisonID) })).filter((entry) => !entry.validation.ok);
    if (invalid.length) {
      setNotice({ kind: "error", message: invalid.map((entry) => `${entry.column.label}: ${entry.validation.ok ? "" : t(entry.validation.errorKey, entry.validation.errorValues)}`).join(" • ") });
      return;
    }
    setNotice(null);
    await Promise.allSettled(columns.map((column) => runColumn(column, comparisonID)));
  }

  async function runColumn(column: PlaygroundColumnDraft, comparisonID = createClientID("comparison")) {
    const validation = validateAndBuildRunPayload(column, comparisonID);
    if (!validation.ok) {
      setNotice({ kind: "error", message: t(validation.errorKey, validation.errorValues) });
      return;
    }
    const controller = new AbortController();
    abortControllers.current.set(column.id, controller);
    const previousResult = results[column.id] ?? emptyRunResult();
    const startedAt = performance.now();
    let firstTokenAt: number | undefined;
    let responseText = "";
    let toolCalls: unknown[] = [];
    let finishReason: string | undefined;
    let usage: Record<string, unknown> | undefined;
    let responseID: string | undefined;
    let responseModel: string | undefined;
    let rawResponse: unknown = null;
    setNotice(null);
    setResults((current) => {
      const previous = current[column.id] ?? previousResult;
      return {
        ...current,
        [column.id]: {
          ...emptyRunResult(),
          status: "running",
          responseText: previous.responseText,
          rawResponse: previous.rawResponse,
          toolCalls: previous.toolCalls,
          finishReason: previous.finishReason,
          telemetry: previous.telemetry,
          responseIsPrevious: Boolean(previous.responseText || previous.toolCalls.length),
          startedAt,
        },
      };
    });

    try {
      const response = await fetch("/api/playground/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(validation.payload),
        signal: controller.signal,
      });
      if (!response.ok) {
        const payload = await response.json().catch(() => null) as { error?: { message?: string } } | null;
        throw new Error(payload?.error?.message || `${t("errors.run")} (${response.status})`);
      }

      const baseTelemetry = telemetryFromHeaders(response.headers, column.selectedModel, startedAt);
      const contentType = response.headers.get("content-type") ?? "";
      if (validation.payload.request.stream === true && contentType.includes("text/event-stream")) {
        const streamed = await consumeStream(response, {
          signal: controller.signal,
          disconnectedMessage: t("errors.streamDisconnected"),
          onUpdate: (snapshot) => {
            responseText = snapshot.responseText;
            toolCalls = snapshot.toolCalls;
            finishReason = snapshot.finishReason;
            usage = snapshot.usage;
            responseID = snapshot.responseID;
            responseModel = snapshot.responseModel;
            firstTokenAt = snapshot.firstTokenAt;
            setResults((current) => ({
              ...current,
              [column.id]: {
                ...current[column.id], status: "running", responseText, toolCalls, finishReason,
                rawResponse: { id: responseID, model: responseModel, choices: [{ message: { role: "assistant", content: responseText, tool_calls: toolCalls }, finish_reason: finishReason }], usage },
                responseIsPrevious: false,
                telemetry: { ...baseTelemetry, effectiveModel: responseModel ?? baseTelemetry.effectiveModel, ttftMs: firstTokenAt === undefined ? undefined : Math.round(firstTokenAt - startedAt), ...usageTelemetry(usage) },
              },
            }));
          },
        });
        responseText = streamed.responseText;
        toolCalls = streamed.toolCalls;
        finishReason = streamed.finishReason;
        usage = streamed.usage;
        responseID = streamed.responseID;
        responseModel = streamed.responseModel;
        firstTokenAt = streamed.firstTokenAt;
      } else {
        const payload = await response.json() as Record<string, unknown>;
        const choice = Array.isArray(payload.choices) ? payload.choices[0] as Record<string, unknown> | undefined : undefined;
        const message = choice?.message && typeof choice.message === "object" ? choice.message as Record<string, unknown> : undefined;
        responseText = extractPlaygroundText(message?.content);
        toolCalls = Array.isArray(message?.tool_calls) ? message.tool_calls : [];
        finishReason = typeof choice?.finish_reason === "string" ? choice.finish_reason : undefined;
        usage = payload.usage && typeof payload.usage === "object" ? payload.usage as Record<string, unknown> : undefined;
        responseID = typeof payload.id === "string" ? payload.id : undefined;
        responseModel = typeof payload.model === "string" ? payload.model : undefined;
        rawResponse = payload;
      }

      const durationMs = Math.round(performance.now() - startedAt);
      const normalizedResponse = { id: responseID, model: responseModel, choices: [{ message: { role: "assistant", content: responseText, ...(toolCalls.length ? { tool_calls: toolCalls } : {}) }, finish_reason: finishReason }], usage };
      setResults((current) => ({
        ...current,
        [column.id]: {
          status: "success", responseText, rawResponse: rawResponse ?? normalizedResponse, toolCalls, finishReason, responseIsPrevious: false,
          telemetry: { ...telemetryFromHeaders(response.headers, column.selectedModel, startedAt), effectiveModel: responseModel ?? column.selectedModel, durationMs, ttftMs: firstTokenAt === undefined ? undefined : Math.round(firstTokenAt - startedAt), ...usageTelemetry(usage) },
        },
      }));
    } catch (error) {
      if (controller.signal.aborted) {
        setResults((current) => {
          const active = current[column.id];
          return {
            ...current,
            [column.id]: {
              ...active,
              status: "stopped",
              error: t("stopCost"),
              telemetry: active?.responseIsPrevious
                ? previousResult.telemetry
                : { ...active?.telemetry, durationMs: Math.round(performance.now() - startedAt) },
            },
          };
        });
      } else {
        const hasNewResponse = Boolean(responseText || toolCalls.length);
        setResults((current) => ({
          ...current,
          [column.id]: {
            ...current[column.id],
            status: "error",
            responseText: hasNewResponse ? responseText : previousResult.responseText,
            rawResponse: hasNewResponse ? current[column.id]?.rawResponse ?? rawResponse : previousResult.rawResponse,
            toolCalls: hasNewResponse ? toolCalls : previousResult.toolCalls,
            finishReason: hasNewResponse ? finishReason : previousResult.finishReason,
            responseIsPrevious: !hasNewResponse && Boolean(previousResult.responseText || previousResult.toolCalls.length),
            error: normalizeError(error, t("errors.run")),
            telemetry: hasNewResponse
              ? { ...current[column.id]?.telemetry, requestedModel: column.selectedModel, durationMs: Math.round(performance.now() - startedAt) }
              : previousResult.telemetry,
          },
        }));
      }
    } finally {
      abortControllers.current.delete(column.id);
    }
  }

  function stopColumn(columnID: string) {
    abortControllers.current.get(columnID)?.abort();
  }

  function stopAll() {
    abortControllers.current.forEach((controller) => controller.abort());
  }

  function replaceActiveWithVersion(version: PromptVersion, force = false) {
    if (!activeColumn) return;
    if (activeColumn.dirty && !force) {
      setPendingHref(null);
      setPendingVersion(version);
      setDiscardOpen(true);
      return;
    }
    const next = hydratePlaygroundColumn(version, activeColumn.label);
    next.id = activeColumn.id;
    const model = models.find((item) => item.id === next.selectedModel);
    next.selectedProvider = model?.provider ?? "";
    setColumns((current) => {
      const sharedValues = Object.assign({}, ...current.map((column) => column.variableValues), next.variableValues) as Record<string, string>;
      return current.map((column) => column.id === next.id
        ? { ...next, variableValues: sharedValues }
        : { ...column, variableValues: { ...column.variableValues, ...next.variableValues } });
    });
    setResults((current) => ({ ...current, [next.id]: emptyRunResult() }));
    setLoadOpen(false);
  }

  function columnProps(column: PlaygroundColumnDraft) { return {
    column,
    result: results[column.id] ?? emptyRunResult(),
    locale,
    readOnly: !canRun,
    canSave,
    canLoad,
    canOpenTrace,
    comparison: columns.length > 1,
    onChange: updateColumn,
    onRun: () => void runColumn(column),
    onStop: () => stopColumn(column.id),
    onOpenModel: () => { setActiveColumnID(column.id); setModelPickerOpen(true); },
    onOpenSettings: () => setSettingsColumnID(column.id),
    onOpenSave: () => setSaveColumnID(column.id),
    onOpenLoad: () => { setActiveColumnID(column.id); setLoadOpen(true); },
    onDuplicateColumn: () => duplicateColumn(column),
    onRemoveColumn: () => removeColumn(column.id),
    t,
  }; }

  if (!activeColumn) return null;

  return (
    <div dir={locale === "fa" ? "rtl" : "ltr"} className="flex h-full min-h-0 flex-col bg-muted/20">
      <div className="flex flex-wrap items-center gap-2 border-b border-border bg-background px-3 py-2">
        <div className="flex items-center gap-2 text-sm font-medium"><FlaskConical className="size-4 text-primary" />{t("newDraft")}</div>
        {columns.length > 1 ? (
          <Tabs value={activeColumnID} onValueChange={setActiveColumnID} className="min-w-0 md:hidden">
            <TabsList>{columns.map((column) => <TabsTrigger key={column.id} value={column.id}>{column.label}</TabsTrigger>)}</TabsList>
          </Tabs>
        ) : null}
        <div className="ms-auto flex flex-wrap items-center gap-2">
          <Button type="button" size="sm" variant="outline" onClick={() => setVariablesOpen(true)}><Variable data-icon="inline-start" />{t("actions.variables", { count: allVariables.length })}</Button>
          <Button type="button" size="sm" variant="outline" disabled={columns.length >= 3} onClick={addComparison}><Plus data-icon="inline-start" />{t("actions.compare")}</Button>
          <Button type="button" size="sm" variant="outline" disabled={!canLoad} onClick={() => setLoadOpen(true)}><FolderOpen data-icon="inline-start" />{t("actions.load")}</Button>
        </div>
      </div>

      {!canRun ? <Alert variant="info" className="m-3 mb-0"><AlertDescription>{t("readOnly")}</AlertDescription></Alert> : null}
      {modelError && models.length > 0 ? <Alert variant="destructive" className="m-3 mb-0"><AlertDescription>{modelError}</AlertDescription></Alert> : null}
      <OperationNotice notice={notice} labels={{ success: t("notice.success"), error: t("notice.error"), warning: t("notice.warning"), info: t("notice.info") }} className="m-3 mb-0" />

      {modelsLoading ? (
        <div className="flex flex-1 items-center justify-center gap-2 text-sm text-muted-foreground"><Spinner label={t("model.loading")} />{t("model.loading")}</div>
      ) : modelError ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-4 p-6 text-center"><AlertTriangle className="size-8 text-destructive" /><div><h2 className="font-semibold">{t("errors.loadModels")}</h2><p className="mt-1 text-sm text-muted-foreground">{modelError}</p></div><Button type="button" onClick={() => void loadModels()}>{t("actions.retry")}</Button></div>
      ) : models.length === 0 ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-4 p-6 text-center"><AlertTriangle className="size-8 text-warning" /><div><h2 className="font-semibold">{t("model.unavailableTitle")}</h2><p className="mt-1 text-sm text-muted-foreground">{t("model.unavailableDescription")}</p></div><Button asChild><a href="/provider-accounts/new">{t("model.connect")}</a></Button></div>
      ) : (
        <div className="min-h-0 flex-1 overflow-hidden p-3">
          {columns.length === 1 ? (
            <>
              <div className="h-full md:hidden">
                <Tabs defaultValue="editor" className="h-full gap-0">
                  <TabsList className="mx-3 mb-2"><TabsTrigger value="editor">{t("editor")}</TabsTrigger><TabsTrigger value="output">{t("output")}</TabsTrigger></TabsList>
                  <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-border bg-background">
                    <TabsContent value="editor" className="h-full overflow-hidden"><PlaygroundEditorPanel {...columnProps(activeColumn)} /></TabsContent>
                    <TabsContent value="output" className="h-full overflow-hidden"><PlaygroundOutputPanel {...columnProps(activeColumn)} /></TabsContent>
                  </div>
                </Tabs>
              </div>
              <ResizableSplitPane
                className="hidden h-full overflow-hidden rounded-lg border border-border bg-background md:grid"
                direction={locale === "fa" ? "rtl" : "ltr"}
                startLabel={t("editor")}
                endLabel={t("output")}
                separatorLabel={t("separator")}
                start={<PlaygroundEditorPanel {...columnProps(activeColumn)} />}
                end={<PlaygroundOutputPanel {...columnProps(activeColumn)} />}
              />
            </>
          ) : (
            <>
              <div className="h-full overflow-hidden rounded-lg border border-border bg-background md:hidden">
                <Tabs defaultValue="editor" className="h-full gap-0">
                  <TabsList className="m-2"><TabsTrigger value="editor">{t("editor")}</TabsTrigger><TabsTrigger value="output">{t("output")}</TabsTrigger></TabsList>
                  <TabsContent value="editor" className="min-h-0 flex-1 overflow-hidden"><PlaygroundEditorPanel {...columnProps(activeColumn)} /></TabsContent>
                  <TabsContent value="output" className="min-h-0 flex-1 overflow-hidden"><PlaygroundOutputPanel {...columnProps(activeColumn)} /></TabsContent>
                </Tabs>
              </div>
              <div className="hidden h-full gap-3 overflow-x-auto md:flex">{columns.map((column) => <PlaygroundComparisonColumn key={column.id} {...columnProps(column)} />)}</div>
            </>
          )}
        </div>
      )}

      <div className="sticky bottom-0 z-10 flex items-center gap-3 border-t border-border bg-background/95 px-4 py-3 backdrop-blur">
        <p className="min-w-0 flex-1 text-xs text-muted-foreground">{columns.length > 1 ? t("billable") : t("shortcutHint")}</p>
        {anyRunning ? <Button type="button" variant="destructive" onClick={columns.length > 1 ? stopAll : () => stopColumn(activeColumn.id)}><Square data-icon="inline-start" />{columns.length > 1 ? t("actions.stopAll") : t("actions.stop")}</Button> : <Button type="button" aria-label={columns.length > 1 ? t("actions.runAllShortcut") : t("actions.runShortcut")} title={columns.length > 1 ? t("actions.runAllShortcut") : t("actions.runShortcut")} disabled={!canRun || !columns.every((column) => column.selectedModel)} onClick={() => columns.length > 1 ? void runAll() : void runColumn(activeColumn)}><Play data-icon="inline-start" />{columns.length > 1 ? t("actions.runAll") : t("actions.run")}</Button>}
      </div>

      <ModelPickerDialog open={modelPickerOpen} onOpenChange={setModelPickerOpen} search={modelSearch} onSearchChange={setModelSearch} models={models} selectedModel={activeColumn.selectedModel} onSelect={(model) => { updateColumn({ ...activeColumn, selectedModel: model.id, selectedProvider: model.provider ?? "", dirty: true }); setModelPickerOpen(false); }} t={t} />
      <PlaygroundSettingsSheet open={Boolean(settingsColumnID)} onOpenChange={(open) => !open && setSettingsColumnID(null)} column={settingsColumn} guardrails={guardrails} guardrailsLoading={guardrailsLoading} readOnly={!canRun} side={locale === "fa" ? "left" : "right"} onChange={updateColumn} t={t} />
      <VariablesSheet
        open={variablesOpen}
        onOpenChange={setVariablesOpen}
        variables={allVariables}
        columns={columns}
        readOnly={!canRun}
        side={locale === "fa" ? "left" : "right"}
        onChange={(name, value) => {
          setColumns((current) => current.map((column) => ({
            ...column,
            variableValues: { ...column.variableValues, [name]: value },
            dirty: true,
          })));
        }}
        t={t}
      />
      <SavePromptDialog key={`save-${saveColumnID ?? "closed"}`} open={Boolean(saveColumnID)} onOpenChange={(open) => !open && setSaveColumnID(null)} column={columns.find((column) => column.id === saveColumnID) ?? null} onSaved={(saved) => { const target = columns.find((column) => column.id === saveColumnID); if (target) updateColumn({ ...target, sourcePromptID: saved.prompt_id, sourceVersionID: saved.id, sourceFQN: saved.fqn, sourcePromptName: saved.name, sourcePromptDescription: saved.description ?? "", dirty: false }); setSaveColumnID(null); setNotice({ kind: "success", message: t("save.success") }); }} t={t} />
      <LoadPromptDialog key={`load-${loadOpen ? "open" : "closed"}`} open={loadOpen} onOpenChange={setLoadOpen} onSelect={(version) => replaceActiveWithVersion(version)} t={t} />
      <UnsavedChangesDialog open={discardOpen} onOpenChange={(open) => { setDiscardOpen(open); if (!open) { setPendingVersion(null); setPendingHref(null); } }} title={t("dirty.title")} description={t("dirty.description")} stayLabel={t("actions.keepEditing")} discardLabel={t("actions.discard")} onDiscard={() => {
        if (pendingVersion) {
          replaceActiveWithVersion(pendingVersion, true);
        } else if (pendingHref) {
          confirmedNavigation.current = true;
          window.location.assign(pendingHref);
          window.setTimeout(() => { confirmedNavigation.current = false; }, 1000);
        }
      }} />
    </div>
  );
}

function ModelPickerDialog(props: { open: boolean; onOpenChange: (open: boolean) => void; search: string; onSearchChange: (value: string) => void; models: GatewayModel[]; selectedModel: string; onSelect: (model: GatewayModel) => void; t: PlaygroundTranslation }) {
  const q = props.search.trim().toLowerCase();
  const models = props.models.filter((model) => `${model.id} ${model.provider ?? ""} ${(model.aliases ?? []).join(" ")}`.toLowerCase().includes(q));
  return (
    <SearchPickerDialog open={props.open} onOpenChange={props.onOpenChange} title={props.t("model.select")} searchValue={props.search} onSearchValueChange={props.onSearchChange} searchPlaceholder={props.t("model.search")}>
      <ScrollArea className="max-h-96">
        <div className="flex flex-col gap-2">
          {models.map((model) => <button key={model.id} type="button" aria-pressed={model.id === props.selectedModel} className={cn("flex w-full items-center gap-3 rounded-lg border border-border p-3 text-start hover:bg-accent", model.id === props.selectedModel && "border-primary bg-primary/5")} onClick={() => props.onSelect(model)}><span className="min-w-0 flex-1"><span className="block truncate font-mono text-sm" dir="ltr">{model.id}</span><span className="block truncate text-xs text-muted-foreground">{model.provider || model.owned_by}</span></span>{model.provider === "virtual" ? <Badge variant="secondary">{props.t("model.virtual")}</Badge> : null}</button>)}
          {models.length === 0 ? <p className="p-6 text-center text-sm text-muted-foreground">{props.t("model.empty")}</p> : null}
        </div>
      </ScrollArea>
    </SearchPickerDialog>
  );
}

function VariablesSheet(props: { open: boolean; onOpenChange: (open: boolean) => void; variables: string[]; columns: PlaygroundColumnDraft[]; readOnly: boolean; side: "left" | "right"; onChange: (name: string, value: string) => void; t: PlaygroundTranslation }) {
  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent
        side={props.side}
        closeLabel={props.t("actions.close")}
        className="flex w-full flex-col gap-0 p-0 sm:max-w-lg"
        dir={props.side === "left" ? "rtl" : "ltr"}
      >
        <SheetHeader className="border-b border-border p-4">
          <SheetTitle>{props.t("variables.title")}</SheetTitle>
          <SheetDescription>{props.t("variables.description")}</SheetDescription>
        </SheetHeader>
        <ScrollArea className="min-h-0 flex-1">
          <div className="flex flex-col gap-4 p-4">
            {props.variables.map((name) => (
              <AppTextField
                key={name}
                label={name}
                dir="auto"
                value={props.columns.find((column) => column.variableValues[name] !== undefined)?.variableValues[name] ?? ""}
                readOnly={props.readOnly}
                placeholder={props.t("variables.value", { name })}
                onChange={(event) => props.onChange(name, event.target.value)}
              />
            ))}
            {props.variables.length === 0 ? <p className="text-sm text-muted-foreground">{props.t("variables.empty", { placeholder: "{{variable_name}}" })}</p> : null}
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}

function SavePromptDialog(props: { open: boolean; onOpenChange: (open: boolean) => void; column: PlaygroundColumnDraft | null; onSaved: (version: PromptVersion) => void; t: PlaygroundTranslation }) {
  const [name, setName] = useState(() => props.column?.sourcePromptName ?? "");
  const [description, setDescription] = useState(() => props.column?.sourcePromptDescription ?? "");
  const [commit, setCommit] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  if (!props.column) return null;
  async function save() {
    if (!props.column || !/^[A-Za-z0-9_-]{1,256}$/.test(name.trim())) { setError(props.t("save.nameHelp")); return; }
    const validation = validateAndBuildRunPayload(props.column, "save_validation");
    if (!validation.ok) { setError(props.t(validation.errorKey, validation.errorValues)); return; }
    setSaving(true); setError("");
    try {
      const manifest = buildPlaygroundManifest(props.column, name.trim(), description.trim());
      const path = props.column.sourcePromptID ? `/api/prompts/${encodeURIComponent(props.column.sourcePromptID)}/versions` : "/api/prompts";
      const saved = await requestJSON<PromptVersion>(path, { method: "POST", body: JSON.stringify({ manifest, commit_message: commit.trim() }) });
      props.onSaved(saved);
    } catch (error) { setError(normalizeError(error, props.t("errors.save"))); }
    finally { setSaving(false); }
  }
  return <FormDialog open={props.open} onOpenChange={props.onOpenChange} title={props.column.sourcePromptID ? props.t("save.titleVersion") : props.t("save.titleNew")} description={props.t("save.description")} pending={saving} footer={<><Button variant="outline" onClick={() => props.onOpenChange(false)} disabled={saving}>{props.t("actions.cancel")}</Button><Button onClick={() => void save()} disabled={saving}>{saving ? <Spinner /> : <Save data-icon="inline-start" />}{saving ? props.t("save.saving") : props.column.sourcePromptID ? props.t("actions.saveVersion") : props.t("actions.save")}</Button></>}><AppTextField label={props.t("save.name")} description={props.t("save.nameHelp")} value={name} dir="ltr" disabled={Boolean(props.column.sourcePromptID)} onChange={(event) => setName(event.target.value)} /><AppTextField label={props.t("save.promptDescription")} value={description} onChange={(event) => setDescription(event.target.value)} /><AppTextField label={props.t("save.commit")} value={commit} onChange={(event) => setCommit(event.target.value)} /><p className="text-xs text-muted-foreground">{props.t("save.summary", { messages: props.column.messages.length, variables: getColumnVariables(props.column).length, tools: props.column.tools.length })}</p>{error ? <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert> : null}</FormDialog>;
}

function LoadPromptDialog(props: { open: boolean; onOpenChange: (open: boolean) => void; onSelect: (version: PromptVersion) => void; t: PlaygroundTranslation }) {
  const { open, t } = props;
  const [search, setSearch] = useState("");
  const [prompts, setPrompts] = useState<Prompt[]>([]);
  const [versions, setVersions] = useState<PromptVersion[]>([]);
  const [selectedPromptID, setSelectedPromptID] = useState("");
  const [promptsLoading, setPromptsLoading] = useState(open);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [error, setError] = useState("");
  const versionsController = useRef<AbortController | null>(null);
  useEffect(() => {
    if (!open) return;
    const controller = new AbortController();
    void requestJSON<PromptListResponse>("/api/prompts?limit=100", { signal: controller.signal })
      .then((payload) => setPrompts(payload.data ?? []))
      .catch((requestError) => { if (!isAbortError(requestError)) setError(normalizeError(requestError, t("errors.load"))); })
      .finally(() => { if (!controller.signal.aborted) setPromptsLoading(false); });
    return () => {
      controller.abort();
      versionsController.current?.abort();
    };
  }, [open, t]);

  async function selectPrompt(prompt: Prompt) {
    versionsController.current?.abort();
    const controller = new AbortController();
    versionsController.current = controller;
    setSelectedPromptID(prompt.id);
    setVersions([]);
    setVersionsLoading(true);
    setError("");
    try {
      const payload = await requestJSON<PromptVersionListResponse>(`/api/prompts/${encodeURIComponent(prompt.id)}/versions?limit=100`, { signal: controller.signal });
      if (!controller.signal.aborted) setVersions(payload.data ?? []);
    } catch (requestError) {
      if (!isAbortError(requestError)) setError(normalizeError(requestError, t("errors.load")));
    } finally {
      if (!controller.signal.aborted) setVersionsLoading(false);
    }
  }
  const visible = prompts.filter((prompt) => `${prompt.name} ${prompt.slug}`.toLowerCase().includes(search.toLowerCase()));
  return (
    <SearchPickerDialog open={props.open} onOpenChange={props.onOpenChange} title={props.t("load.title")} description={props.t("load.description")} searchValue={search} onSearchValueChange={setSearch} searchPlaceholder={props.t("load.search")} size="xl">
      {error ? <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert> : null}
      <div className="grid min-h-80 gap-3 md:grid-cols-2">
        <ScrollArea className="max-h-96 rounded-lg border border-border p-2">
          <div className="flex flex-col gap-1">
            {visible.map((prompt) => <button key={prompt.id} type="button" aria-pressed={selectedPromptID === prompt.id} className={cn("w-full rounded-md px-3 py-2 text-start hover:bg-accent", selectedPromptID === prompt.id && "bg-primary/10 text-primary")} onClick={() => void selectPrompt(prompt)}><span className="block font-medium">{prompt.name}</span><span className="block font-mono text-xs text-muted-foreground" dir="ltr">{prompt.slug}</span></button>)}
            {promptsLoading ? <div className="flex items-center justify-center p-4"><Spinner label={props.t("load.loading")} /></div> : null}
            {!promptsLoading && visible.length === 0 ? <p className="p-4 text-center text-sm text-muted-foreground">{props.t("load.empty")}</p> : null}
          </div>
        </ScrollArea>
        <ScrollArea className="max-h-96 rounded-lg border border-border p-2">
          <div className="flex flex-col gap-1">
            <h3 className="px-2 py-1 text-sm font-medium">{props.t("load.versions")}</h3>
            {versions.map((version) => <button key={version.id} type="button" className="flex w-full items-center justify-between gap-2 rounded-md px-3 py-2 text-start hover:bg-accent" onClick={() => props.onSelect(version)}><span>v{version.version}</span><span className="min-w-0 truncate font-mono text-xs text-muted-foreground" dir="ltr">{version.fqn}</span></button>)}
            {versionsLoading ? <div className="flex items-center justify-center p-4"><Spinner label={props.t("load.loading")} /></div> : null}
            {!versionsLoading && selectedPromptID && versions.length === 0 ? <p className="p-4 text-center text-sm text-muted-foreground">{props.t("load.noVersions")}</p> : null}
          </div>
        </ScrollArea>
      </div>
    </SearchPickerDialog>
  );
}

function emptyRunResult(): PlaygroundRunResult { return { status: "idle", responseText: "", rawResponse: null, toolCalls: [] }; }

function isAbortError(error: unknown) { return error instanceof Error && error.name === "AbortError"; }

function createClientID(prefix: string) { return `${prefix}_${globalThis.crypto?.randomUUID?.() ?? Math.random().toString(36).slice(2)}`; }

function usageTelemetry(usage?: Record<string, unknown>): Partial<PlaygroundTelemetry> {
  if (!usage) return {};
  const number = (value: unknown) => typeof value === "number" && Number.isFinite(value) ? value : undefined;
  return { inputTokens: number(usage.prompt_tokens ?? usage.input_tokens), outputTokens: number(usage.completion_tokens ?? usage.output_tokens), totalTokens: number(usage.total_tokens) };
}

function telemetryFromHeaders(headers: Headers, requestedModel: string, startedAt: number): PlaygroundTelemetry {
  const numeric = (name: string) => { const raw = headers.get(name); if (!raw) return undefined; const value = Number(raw); return Number.isFinite(value) ? value : undefined; };
  return {
    requestedModel,
    traceId: headers.get("x-godarai-trace-id") ?? undefined,
    requestId: headers.get("x-request-id") ?? undefined,
    cacheStatus: headers.get("x-godarai-cache-status") ?? undefined,
    cacheSimilarity: numeric("x-godarai-cache-similarity-score"),
    costUSD: numeric("x-godarai-cost-usd"),
    costIRR: numeric("x-godarai-cost-irr"),
    durationMs: Math.round(performance.now() - startedAt),
  };
}
