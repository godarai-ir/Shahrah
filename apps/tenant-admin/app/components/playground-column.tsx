"use client";

import { useRef } from "react";
import { Braces, CopyPlus, FileUp, Play, Plus, Save, Settings2, Square, Trash2 } from "lucide-react";

import {
  AppCodeTextareaField,
  CopyAction,
  EmptyState,
  PromptMessageEditor,
  RunTelemetryStrip,
  StatusBadge,
  type PromptMessageRole,
  type RunTelemetryItem,
} from "@/components/app-ui";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CodeBlock } from "@/components/app-ui/CodeBlock";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Spinner } from "@/components/ui/spinner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import {
  applyColumnJSON,
  createPlaygroundMessage,
  serializeColumnJSON,
  type PlaygroundColumnDraft,
  type PlaygroundRunResult,
} from "./playground-domain";

type Translation = (key: string, values?: Record<string, string | number>) => string;

type PlaygroundColumnProps = {
  column: PlaygroundColumnDraft;
  result: PlaygroundRunResult;
  locale: string;
  readOnly: boolean;
  canSave: boolean;
  canLoad: boolean;
  canOpenTrace: boolean;
  comparison: boolean;
  onChange: (column: PlaygroundColumnDraft) => void;
  onRun: () => void;
  onStop: () => void;
  onOpenModel: () => void;
  onOpenSettings: () => void;
  onOpenSave: () => void;
  onOpenLoad: () => void;
  onDuplicateColumn?: () => void;
  onRemoveColumn?: () => void;
  t: Translation;
};

export function PlaygroundEditorPanel(props: PlaygroundColumnProps) {
  const { column } = props;
  const draggedMessageID = useRef<string | null>(null);
  const roleOptions = (["system", "user", "assistant", "tool"] as const).map((value) => ({ value, label: props.t(`roles.${value}`) }));

  function change(next: PlaygroundColumnDraft) {
    next.dirty = true;
    if (next.editorMode === "visual") next.jsonDraft = serializeColumnJSON(next);
    props.onChange(next);
  }

  function updateMessage(index: number, patch: Partial<PlaygroundColumnDraft["messages"][number]>) {
    change({ ...column, messages: column.messages.map((message, current) => current === index ? { ...message, ...patch } : message) });
  }

  function moveMessage(index: number, offset: number) {
    const nextIndex = index + offset;
    if (nextIndex < 0 || nextIndex >= column.messages.length) return;
    const messages = [...column.messages];
    [messages[index], messages[nextIndex]] = [messages[nextIndex], messages[index]];
    change({ ...column, messages });
  }

  function switchMode(mode: string) {
    if (mode !== "visual" && mode !== "json") return;
    if (mode === "json") {
      props.onChange({ ...column, editorMode: "json", jsonDraft: serializeColumnJSON(column), jsonError: "" });
      return;
    }
    const parsed = applyColumnJSON(column);
    if (!parsed.ok) {
      props.onChange({ ...column, jsonError: props.t(parsed.errorKey, parsed.errorValues) });
      return;
    }
    props.onChange({ ...parsed.column, editorMode: "visual" });
  }

  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-3 py-2">
        {props.comparison ? (
          <Input
            value={column.label}
            onChange={(event) => change({ ...column, label: event.target.value })}
            className="h-8 w-28 font-medium"
            aria-label={column.label}
            readOnly={props.readOnly}
          />
        ) : <h2 className="text-sm font-semibold">{props.t("editor")}</h2>}
        <Button type="button" size="sm" variant="outline" className="min-w-0 max-w-full" disabled={props.readOnly} onClick={props.onOpenModel}>
          <span className="truncate font-mono" dir="ltr">{column.selectedModel || props.t("model.select")}</span>
          {column.selectedProvider === "virtual" ? <Badge variant="secondary">{props.t("model.virtual")}</Badge> : null}
        </Button>
        <div className="ms-auto flex items-center gap-1">
          <Button type="button" variant="ghost" size="icon-sm" aria-label={props.t("actions.load")} title={props.t("actions.load")} disabled={!props.canLoad} onClick={props.onOpenLoad}><FileUp data-icon="inline-start" /></Button>
          <Button type="button" variant="ghost" size="icon-sm" aria-label={props.t("actions.parameters")} title={props.t("actions.parameters")} onClick={props.onOpenSettings}><Settings2 data-icon="inline-start" /></Button>
          <Button type="button" variant="ghost" size="icon-sm" aria-label={column.sourceVersionID ? props.t("actions.saveVersion") : props.t("actions.save")} title={props.t("actions.save")} disabled={!props.canSave} onClick={props.onOpenSave}><Save data-icon="inline-start" /></Button>
          {props.comparison && props.onDuplicateColumn ? <Button type="button" variant="ghost" size="icon-sm" aria-label={props.t("actions.duplicate")} title={props.t("actions.duplicate")} onClick={props.onDuplicateColumn}><CopyPlus data-icon="inline-start" /></Button> : null}
          {props.comparison && props.onRemoveColumn ? <Button type="button" variant="ghost" size="icon-sm" aria-label={props.t("actions.remove")} title={props.t("actions.remove")} onClick={props.onRemoveColumn}><Trash2 data-icon="inline-start" /></Button> : null}
        </div>
      </div>

      <Tabs value={column.editorMode} onValueChange={switchMode} className="min-h-0 flex-1 gap-0">
        <div className="flex items-center justify-between border-b border-border px-3 py-2">
          <TabsList>
            <TabsTrigger value="visual">{props.t("visual")}</TabsTrigger>
            <TabsTrigger value="json"><Braces />{props.t("json")}</TabsTrigger>
          </TabsList>
          {column.editorMode === "visual" ? (
            <Button
              type="button"
              size="sm"
              variant="ghost"
              disabled={props.readOnly}
              onClick={() => change({ ...column, messages: [...column.messages, createPlaygroundMessage()] })}
            ><Plus data-icon="inline-start" />{props.t("actions.addMessage")}</Button>
          ) : null}
        </div>
        <TabsContent value="visual" className="min-h-0 overflow-hidden">
          <ScrollArea className="h-full px-3">
            {column.messages.map((message, index) => (
              <PromptMessageEditor
                key={message.id}
                role={message.role}
                content={message.content}
                roleOptions={roleOptions}
                roleLabel={props.t("message.role")}
                contentLabel={props.t("message.content")}
                toolCallId={message.toolCallId}
                toolCallIdLabel={props.t("message.toolCallId")}
                duplicateLabel={props.t("actions.duplicate")}
                removeLabel={props.t("actions.remove")}
                dragLabel={props.t("actions.drag")}
                moveUpLabel={props.t("actions.moveUp")}
                moveDownLabel={props.t("actions.moveDown")}
                readOnly={props.readOnly}
                disableMoveUp={index === 0}
                disableMoveDown={index === column.messages.length - 1}
                error={!message.content.trim() ? props.t("message.content") : undefined}
                onRoleChange={(role: PromptMessageRole) => updateMessage(index, { role })}
                onContentChange={(content) => updateMessage(index, { content })}
                onToolCallIdChange={(toolCallId) => updateMessage(index, { toolCallId })}
                onDuplicate={() => {
                  const messages = [...column.messages];
                  messages.splice(index + 1, 0, { ...message, id: createPlaygroundMessage().id });
                  change({ ...column, messages });
                }}
                onRemove={() => change({ ...column, messages: column.messages.filter((_, current) => current !== index) })}
                onMoveUp={() => moveMessage(index, -1)}
                onMoveDown={() => moveMessage(index, 1)}
                onDragStart={(event) => {
                  draggedMessageID.current = message.id;
                  event.dataTransfer.effectAllowed = "move";
                  event.dataTransfer.setData("text/plain", message.id);
                }}
                onDragEnd={() => { draggedMessageID.current = null; }}
                onDragOver={(event) => {
                  if (draggedMessageID.current && draggedMessageID.current !== message.id) event.preventDefault();
                }}
                onDrop={(event) => {
                  event.preventDefault();
                  const sourceID = draggedMessageID.current || event.dataTransfer.getData("text/plain");
                  draggedMessageID.current = null;
                  const sourceIndex = column.messages.findIndex((entry) => entry.id === sourceID);
                  if (sourceIndex < 0 || sourceIndex === index) return;
                  const messages = [...column.messages];
                  const [moved] = messages.splice(sourceIndex, 1);
                  messages.splice(index, 0, moved);
                  change({ ...column, messages });
                }}
              />
            ))}
          </ScrollArea>
        </TabsContent>
        <TabsContent value="json" className="min-h-0 overflow-auto p-3">
          <AppCodeTextareaField
            label={props.t("json")}
            value={column.jsonDraft}
            error={column.jsonError || undefined}
            readOnly={props.readOnly}
            onChange={(event) => props.onChange({ ...column, jsonDraft: event.target.value, jsonError: "", dirty: true })}
            className="min-h-96"
          />
          <div className="mt-3 flex justify-end">
            <Button type="button" size="sm" disabled={props.readOnly} onClick={() => switchMode("visual")}>{props.t("actions.applyJson")}</Button>
          </div>
        </TabsContent>
      </Tabs>

      <div className="flex items-center gap-2 border-t border-border p-3 md:hidden">
        {props.result.status === "running" ? (
          <Button type="button" variant="destructive" className="flex-1" onClick={props.onStop}><Square data-icon="inline-start" />{props.t("actions.stop")}</Button>
        ) : (
          <Button type="button" className="flex-1" aria-label={props.t("actions.runShortcut")} title={props.t("actions.runShortcut")} disabled={props.readOnly || !column.selectedModel} onClick={props.onRun}><Play data-icon="inline-start" />{props.t("actions.run")}</Button>
        )}
      </div>
    </div>
  );
}

export function PlaygroundOutputPanel(props: PlaygroundColumnProps) {
  const { result } = props;
  const statusMap = { idle: "neutral", running: "pending", success: "success", stopped: "warning", error: "error" } as const;
  const telemetryItems = buildTelemetryItems(result, props.locale, props.t);

  return (
    <div className="flex h-full min-h-0 flex-col bg-card">
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <h2 className="text-sm font-semibold">{props.t("output")}</h2>
        <StatusBadge status={statusMap[result.status]}>{props.t(`status.${result.status}`)}</StatusBadge>
        {result.responseIsPrevious ? <StatusBadge status="neutral">{props.t("status.previous")}</StatusBadge> : null}
        {result.telemetry?.cacheStatus === "hit" ? <StatusBadge status="neutral">{props.t("status.cacheHit")}</StatusBadge> : null}
        <div className="ms-auto flex items-center gap-1">
          {result.responseText ? <CopyAction value={result.responseText} copyLabel={props.t("actions.copy")} copiedLabel={props.t("actions.copied")} /> : null}
          {props.canOpenTrace && result.telemetry?.traceId ? (
            <Button type="button" variant="ghost" size="sm" asChild>
              <a href={`/traces/${encodeURIComponent(result.telemetry.traceId)}`}>{props.t("actions.openTrace")}</a>
            </Button>
          ) : null}
        </div>
      </div>

      {result.status === "idle" ? (
        <div className="flex flex-1 items-center justify-center p-6">
          <EmptyState variant="inline" title={props.t("empty.title")} description={props.t("empty.description")} />
        </div>
      ) : (
        <Tabs defaultValue="response" className="min-h-0 flex-1 gap-0">
          <div className="border-b border-border px-3 py-2">
            <TabsList>
              <TabsTrigger value="response">{props.t("response")}</TabsTrigger>
              <TabsTrigger value="raw">{props.t("rawResponse")}</TabsTrigger>
            </TabsList>
          </div>
          <TabsContent value="response" className="min-h-0 overflow-hidden">
            <ScrollArea className="h-full">
              <div className="flex flex-col gap-4 p-4">
                {result.status === "running" && !result.responseText ? <div className="flex items-center gap-2 text-sm text-muted-foreground"><Spinner />{props.t("status.running")}</div> : null}
                {result.error ? <Alert variant="destructive"><AlertDescription>{result.error}</AlertDescription></Alert> : null}
                {result.responseText ? <div className="whitespace-pre-wrap text-sm leading-7" dir="auto">{result.responseText}</div> : null}
                {result.toolCalls.length ? (
                  <div className="space-y-2">
                    <p className="text-xs font-medium text-muted-foreground">{props.t("settings.tools")}</p>
                    <CodeBlock code={JSON.stringify(result.toolCalls, null, 2)} language="json" />
                  </div>
                ) : null}
              </div>
            </ScrollArea>
          </TabsContent>
          <TabsContent value="raw" className="min-h-0 overflow-auto p-4">
            <CodeBlock code={JSON.stringify(result.rawResponse ?? {}, null, 2)} language="json" />
          </TabsContent>
        </Tabs>
      )}

      {result.telemetry ? (
        <div className="border-t border-border p-3">
          <RunTelemetryStrip items={telemetryItems} unavailableLabel={props.t("telemetry.unavailable")} copyLabel={props.t("actions.copy")} copiedLabel={props.t("actions.copied")} />
        </div>
      ) : null}
      <div className="sr-only" role="status" aria-live="polite">
        {result.status === "running" ? props.t("streamAnnouncement.started") : result.status === "success" ? props.t("streamAnnouncement.completed") : result.status === "stopped" ? props.t("streamAnnouncement.stopped") : result.status === "error" ? props.t("streamAnnouncement.failed") : ""}
      </div>
    </div>
  );
}

export function PlaygroundComparisonColumn(props: PlaygroundColumnProps) {
  return (
    <div className="flex min-h-0 min-w-80 flex-1 basis-1/2 flex-col overflow-hidden rounded-lg border border-border bg-background">
      <div className="min-h-80 flex-1 overflow-hidden"><PlaygroundEditorPanel {...props} /></div>
      <div className="min-h-64 flex-1 overflow-hidden border-t border-border"><PlaygroundOutputPanel {...props} /></div>
    </div>
  );
}

function buildTelemetryItems(result: PlaygroundRunResult, locale: string, t: Translation): RunTelemetryItem[] {
  const telemetry = result.telemetry ?? {};
  const number = new Intl.NumberFormat(locale);
  const money = new Intl.NumberFormat(locale, { maximumFractionDigits: 6 });
  const duration = (value?: number) => value === undefined ? t("telemetry.unavailable") : `${number.format(value)} ms`;
  return [
    { key: "input", label: t("telemetry.inputTokens"), value: telemetry.inputTokens === undefined ? null : number.format(telemetry.inputTokens) },
    { key: "output", label: t("telemetry.outputTokens"), value: telemetry.outputTokens === undefined ? null : number.format(telemetry.outputTokens) },
    { key: "total", label: t("telemetry.totalTokens"), value: telemetry.totalTokens === undefined ? null : number.format(telemetry.totalTokens) },
    { key: "duration", label: t("telemetry.duration"), value: duration(telemetry.durationMs) },
    { key: "ttft", label: t("telemetry.ttft"), value: duration(telemetry.ttftMs) },
    { key: "usd", label: t("telemetry.costUSD"), value: telemetry.costUSD === undefined ? null : `$${money.format(telemetry.costUSD)}` },
    { key: "irr", label: t("telemetry.costIRR"), value: telemetry.costIRR === undefined ? null : number.format(telemetry.costIRR) },
    { key: "model", label: t("telemetry.model"), value: telemetry.effectiveModel ?? telemetry.requestedModel ?? null, technical: true },
    { key: "cache", label: t("telemetry.cache"), value: telemetry.cacheStatus ?? null },
    { key: "cache-similarity", label: t("telemetry.cacheSimilarity"), value: telemetry.cacheSimilarity === undefined ? null : number.format(telemetry.cacheSimilarity) },
    { key: "finish", label: t("telemetry.finishReason"), value: result.finishReason ?? null, technical: true },
    { key: "trace", label: t("telemetry.trace"), value: telemetry.traceId ?? telemetry.requestId ?? null, rawValue: telemetry.traceId ?? telemetry.requestId, technical: true },
  ];
}
