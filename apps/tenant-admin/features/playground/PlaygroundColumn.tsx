"use client";

import dynamic from "next/dynamic";
import { useEffect, useRef, useState } from "react";
import {
  Braces,
  CopyPlus,
  Download,
  ExternalLink,
  FileUp,
  Image as ImageIcon,
  MessageSquareText,
  Play,
  Plus,
  Save,
  Settings2,
  Square,
  Trash2,
} from "lucide-react";

import { CopyAction } from "@/components/app-ui/CopyAction";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { formatJson, JsonEditorField } from "@/components/app-ui/JsonEditor";
import {
  PromptMessageEditor,
  type PromptMessageRole,
} from "@/components/app-ui/PromptMessageEditor";
import {
  RunTelemetryStrip,
  type RunTelemetryItem,
} from "@/components/app-ui/RunTelemetryStrip";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { VariableTextareaField } from "@/components/app-ui/VariableTextarea";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CodeBlock } from "@/components/app-ui/CodeBlock";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Spinner } from "@/components/ui/spinner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import {
  applyColumnJSON,
  createPlaygroundMediaInput,
  createPlaygroundMessage,
  parseGuardrailChecks,
  serializeColumnJSON,
  type PlaygroundColumnDraft,
  type PlaygroundGeneratedImage,
  type PlaygroundMode,
  type PlaygroundRunResult,
} from "./domain";

const MarkdownRenderer = dynamic(
  () =>
    import("@/components/app-ui/MarkdownRenderer").then(
      (module) => module.MarkdownRenderer,
    ),
  {
    ssr: false,
    loading: () => (
      <div
        className="min-h-24 animate-pulse rounded-xl bg-muted/35"
        aria-hidden="true"
      />
    ),
  },
);

type Translation = (
  key: string,
  values?: Record<string, string | number>,
) => string;

type PlaygroundColumnProps = {
  column: PlaygroundColumnDraft;
  mode: PlaygroundMode;
  result: PlaygroundRunResult;
  locale: string;
  readOnly: boolean;
  canSave: boolean;
  canLoad: boolean;
  canOpenTrace: boolean;
  comparison: boolean;
  variables?: string[];
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
  const [fileError, setFileError] = useState("");
  const roleOptions = (["system", "user", "assistant", "tool"] as const).map(
    (value) => ({ value, label: props.t(`roles.${value}`) }),
  );

  function change(next: PlaygroundColumnDraft) {
    next.dirty = true;
    if (next.editorMode === "visual")
      next.jsonDraft = serializeColumnJSON(next);
    props.onChange(next);
  }

  function updateMessage(
    index: number,
    patch: Partial<PlaygroundColumnDraft["messages"][number]>,
  ) {
    change({
      ...column,
      messages: column.messages.map((message, current) =>
        current === index ? { ...message, ...patch } : message,
      ),
    });
  }

  function moveMessage(index: number, offset: number) {
    const nextIndex = index + offset;
    if (nextIndex < 0 || nextIndex >= column.messages.length) return;
    const messages = [...column.messages];
    [messages[index], messages[nextIndex]] = [
      messages[nextIndex],
      messages[index],
    ];
    change({ ...column, messages });
  }

  function switchMode(mode: string) {
    if (mode !== "visual" && mode !== "json") return;
    if (mode === "json") {
      props.onChange({
        ...column,
        editorMode: "json",
        jsonDraft: formatJson(serializeColumnJSON(column)),
        jsonError: "",
      });
      return;
    }
    const parsed = applyColumnJSON(column);
    if (!parsed.ok) {
      props.onChange({
        ...column,
        jsonError: props.t(parsed.errorKey, parsed.errorValues),
      });
      return;
    }
    props.onChange({ ...parsed.column, editorMode: "visual" });
  }

  async function loadFile(file: File, target: "ocr" | string) {
    if (target === "ocr") {
      const isPDF = file.type === "application/pdf";
      if (!isPDF && file.type !== "image/png" && file.type !== "image/jpeg") {
        setFileError(props.t("validation.ocrFileType"));
        return;
      }
      if (file.size > (isPDF ? 50 : 10) * 1024 * 1024) {
        setFileError(
          props.t(isPDF ? "validation.ocrPDFSize" : "validation.ocrImageSize"),
        );
        return;
      }
    } else if (
      (file.type !== "image/png" && file.type !== "image/jpeg") ||
      file.size > 5 * 1024 * 1024
    ) {
      setFileError(props.t("validation.zaiLocalImageLimit"));
      return;
    }
    setFileError("");
    const value = await readFileAsDataURL(file);
    if (target === "ocr") {
      props.onChange({
        ...column,
        ocrConfig: { ...column.ocrConfig, file: value },
        dirty: true,
      });
      return;
    }
    props.onChange({
      ...column,
      media: column.media.map((item) =>
        item.id === target ? { ...item, value } : item,
      ),
      dirty: true,
    });
  }

  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-3 py-2">
        {props.comparison ? (
          <Input
            value={column.label}
            onChange={(event) =>
              change({ ...column, label: event.target.value })
            }
            className="h-8 w-28 font-medium"
            aria-label={column.label}
            readOnly={props.readOnly}
          />
        ) : (
          <h2 className="text-sm font-semibold">{props.t("targetModel")}</h2>
        )}
        <Button
          type="button"
          size="sm"
          variant="outline"
          className="min-w-0 max-w-full"
          disabled={props.readOnly}
          onClick={props.onOpenModel}
        >
          <span
            className={column.selectedModel ? "truncate font-mono" : "truncate"}
            dir={column.selectedModel ? "ltr" : undefined}
          >
            {column.selectedModel || props.t("model.select")}
          </span>
          {column.selectedProvider === "virtual" ? (
            <Badge variant="secondary">{props.t("model.virtual")}</Badge>
          ) : null}
        </Button>
        <div className="ms-auto flex items-center gap-1">
          {props.mode === "chat" ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={props.t("actions.load")}
              title={props.t("actions.load")}
              disabled={!props.canLoad}
              onClick={props.onOpenLoad}
            >
              <FileUp data-icon="inline-start" />
            </Button>
          ) : null}
          {props.mode !== "layout_parsing" ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={props.t("actions.parameters")}
              title={props.t("actions.parameters")}
              onClick={props.onOpenSettings}
            >
              <Settings2 data-icon="inline-start" />
            </Button>
          ) : null}
          {props.mode === "chat" ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={
                column.sourceVersionID
                  ? props.t("actions.saveVersion")
                  : props.t("actions.save")
              }
              title={props.t("actions.save")}
              disabled={!props.canSave}
              onClick={props.onOpenSave}
            >
              <Save data-icon="inline-start" />
            </Button>
          ) : null}
          {props.comparison && props.onDuplicateColumn ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={props.t("actions.duplicate")}
              title={props.t("actions.duplicate")}
              onClick={props.onDuplicateColumn}
            >
              <CopyPlus data-icon="inline-start" />
            </Button>
          ) : null}
          {props.comparison && props.onRemoveColumn ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={props.t("actions.remove")}
              title={props.t("actions.remove")}
              onClick={props.onRemoveColumn}
            >
              <Trash2 data-icon="inline-start" />
            </Button>
          ) : null}
        </div>
      </div>

      {fileError ? (
        <Alert variant="destructive" className="m-3 mb-0">
          <AlertDescription>{fileError}</AlertDescription>
        </Alert>
      ) : null}

      {props.mode === "image_generation" ? (
        <ScrollArea className="min-h-0 flex-1">
          <div className="flex flex-col gap-4 p-4">
            <Alert variant="info">
              <AlertDescription>
                {props.t("image.temporaryNotice")}
              </AlertDescription>
            </Alert>
            <VariableTextareaField
              label={props.t("image.prompt")}
              description={props.t("image.promptHelp", {
                subject: "{{subject}}",
              })}
              variables={props.variables ?? []}
              value={column.imagePrompt}
              placeholder={props.t("image.promptPlaceholder")}
              readOnly={props.readOnly}
              error={
                !column.imagePrompt.trim()
                  ? props.t("validation.imagePromptRequired")
                  : undefined
              }
              className="min-h-56 resize-y"
              onChange={(event) =>
                props.onChange({
                  ...column,
                  imagePrompt: event.target.value,
                  dirty: true,
                })
              }
            />
          </div>
        </ScrollArea>
      ) : props.mode === "layout_parsing" ? (
        <ScrollArea className="min-h-0 flex-1">
          <div className="flex flex-col gap-4 p-4">
            <Alert variant="info">
              <AlertDescription>{props.t("ocr.help")}</AlertDescription>
            </Alert>
            <AppTextField
              label={props.t("ocr.file")}
              description={props.t("ocr.fileHelp")}
              value={column.ocrConfig.file}
              dir="ltr"
              disabled={props.readOnly}
              onChange={(event) =>
                props.onChange({
                  ...column,
                  ocrConfig: { ...column.ocrConfig, file: event.target.value },
                  dirty: true,
                })
              }
            />
            <label className="flex flex-col gap-2 text-sm font-medium">
              {props.t("ocr.upload")}
              <Input
                type="file"
                accept="application/pdf,image/png,image/jpeg"
                disabled={props.readOnly}
                onChange={(event) => {
                  const file = event.target.files?.[0];
                  if (file) void loadFile(file, "ocr");
                }}
              />
            </label>
            <div className="grid gap-4 sm:grid-cols-2">
              <AppNumberField
                label={props.t("ocr.startPage")}
                value={column.ocrConfig.startPageID}
                min={1}
                step={1}
                disabled={props.readOnly}
                onChange={(event) =>
                  props.onChange({
                    ...column,
                    ocrConfig: {
                      ...column.ocrConfig,
                      startPageID: event.target.value,
                    },
                    dirty: true,
                  })
                }
              />
              <AppNumberField
                label={props.t("ocr.endPage")}
                value={column.ocrConfig.endPageID}
                min={1}
                step={1}
                disabled={props.readOnly}
                onChange={(event) =>
                  props.onChange({
                    ...column,
                    ocrConfig: {
                      ...column.ocrConfig,
                      endPageID: event.target.value,
                    },
                    dirty: true,
                  })
                }
              />
            </div>
            <AppSwitchField
              surface={false}
              label={props.t("ocr.cropImages")}
              checked={column.ocrConfig.returnCropImages}
              disabled={props.readOnly}
              onCheckedChange={(returnCropImages) =>
                props.onChange({
                  ...column,
                  ocrConfig: { ...column.ocrConfig, returnCropImages },
                  dirty: true,
                })
              }
            />
            <AppSwitchField
              surface={false}
              label={props.t("ocr.layoutVisualization")}
              checked={column.ocrConfig.needLayoutVisualization}
              disabled={props.readOnly}
              onCheckedChange={(needLayoutVisualization) =>
                props.onChange({
                  ...column,
                  ocrConfig: { ...column.ocrConfig, needLayoutVisualization },
                  dirty: true,
                })
              }
            />
          </div>
        </ScrollArea>
      ) : (
        <Tabs
          value={column.editorMode}
          onValueChange={switchMode}
          className="min-h-0 flex-1 gap-0"
        >
          <div className="flex items-center justify-between border-b border-border px-3 py-2">
            <TabsList className="bg-muted">
              <TabsTrigger
                value="visual"
                className="h-8 gap-2 px-3 data-[state=active]:border-transparent data-[state=active]:bg-background data-[state=active]:shadow-sm"
              >
                <MessageSquareText />
                {props.t("visual")}
              </TabsTrigger>
              <TabsTrigger
                value="json"
                className="h-8 gap-2 px-3 data-[state=active]:border-transparent data-[state=active]:bg-background data-[state=active]:shadow-sm"
              >
                <Braces />
                {props.t("json")}
              </TabsTrigger>
            </TabsList>
            {column.editorMode === "visual" ? (
              <Button
                type="button"
                size="sm"
                variant="ghost"
                disabled={props.readOnly}
                onClick={() =>
                  change({
                    ...column,
                    messages: [...column.messages, createPlaygroundMessage()],
                  })
                }
              >
                <Plus data-icon="inline-start" />
                {props.t("actions.addMessage")}
              </Button>
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
                  error={
                    !message.content.trim()
                      ? props.t("message.content")
                      : undefined
                  }
                  highlightedVariables={props.variables ?? []}
                  onRoleChange={(role: PromptMessageRole) =>
                    updateMessage(index, { role })
                  }
                  onContentChange={(content) =>
                    updateMessage(index, { content })
                  }
                  onToolCallIdChange={(toolCallId) =>
                    updateMessage(index, { toolCallId })
                  }
                  onDuplicate={() => {
                    const messages = [...column.messages];
                    messages.splice(index + 1, 0, {
                      ...message,
                      id: createPlaygroundMessage().id,
                    });
                    change({ ...column, messages });
                  }}
                  onRemove={() =>
                    change({
                      ...column,
                      messages: column.messages.filter(
                        (_, current) => current !== index,
                      ),
                    })
                  }
                  onMoveUp={() => moveMessage(index, -1)}
                  onMoveDown={() => moveMessage(index, 1)}
                  onDragStart={(event) => {
                    draggedMessageID.current = message.id;
                    event.dataTransfer.effectAllowed = "move";
                    event.dataTransfer.setData("text/plain", message.id);
                  }}
                  onDragEnd={() => {
                    draggedMessageID.current = null;
                  }}
                  onDragOver={(event) => {
                    if (
                      draggedMessageID.current &&
                      draggedMessageID.current !== message.id
                    )
                      event.preventDefault();
                  }}
                  onDrop={(event) => {
                    event.preventDefault();
                    const sourceID =
                      draggedMessageID.current ||
                      event.dataTransfer.getData("text/plain");
                    draggedMessageID.current = null;
                    const sourceIndex = column.messages.findIndex(
                      (entry) => entry.id === sourceID,
                    );
                    if (sourceIndex < 0 || sourceIndex === index) return;
                    const messages = [...column.messages];
                    const [moved] = messages.splice(sourceIndex, 1);
                    messages.splice(index, 0, moved);
                    change({ ...column, messages });
                  }}
                />
              ))}
              <section className="mb-4 flex flex-col gap-3 rounded-lg border border-border p-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <h3 className="text-sm font-medium">
                      {props.t("media.title")}
                    </h3>
                    <p className="text-xs text-muted-foreground">
                      {props.t("media.help")}
                    </p>
                  </div>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={props.readOnly}
                    onClick={() =>
                      change({
                        ...column,
                        media: [...column.media, createPlaygroundMediaInput()],
                      })
                    }
                  >
                    <Plus data-icon="inline-start" />
                    {props.t("media.add")}
                  </Button>
                </div>
                {column.media.map((item) => (
                  <div
                    key={item.id}
                    className="grid gap-2 sm:grid-cols-[10rem_1fr_auto]"
                  >
                    <AppSelectField
                      label={props.t("media.type")}
                      value={item.type}
                      disabled={props.readOnly}
                      onValueChange={(type) =>
                        change({
                          ...column,
                          media: column.media.map((entry) =>
                            entry.id === item.id
                              ? { ...entry, type: type as typeof item.type }
                              : entry,
                          ),
                        })
                      }
                      options={[
                        { value: "image_url", label: props.t("media.image") },
                        { value: "video_url", label: props.t("media.video") },
                        { value: "file_url", label: props.t("media.document") },
                      ]}
                    />
                    <AppTextField
                      label={props.t("media.value")}
                      value={item.value}
                      dir="ltr"
                      disabled={props.readOnly}
                      placeholder={props.t("media.placeholder")}
                      onChange={(event) =>
                        change({
                          ...column,
                          media: column.media.map((entry) =>
                            entry.id === item.id
                              ? { ...entry, value: event.target.value }
                              : entry,
                          ),
                        })
                      }
                    />
                    <div className="flex items-center gap-1">
                      {item.type === "image_url" ? (
                        <Button
                          asChild
                          type="button"
                          size="sm"
                          variant="outline"
                          aria-label={props.t("media.upload")}
                        >
                          <label className="cursor-pointer">
                            <FileUp />
                            <Input
                              className="sr-only"
                              type="file"
                              accept="image/png,image/jpeg"
                              disabled={props.readOnly}
                              onChange={(event) => {
                                const file = event.target.files?.[0];
                                if (file) void loadFile(file, item.id);
                              }}
                            />
                          </label>
                        </Button>
                      ) : null}
                      <Button
                        type="button"
                        size="icon-sm"
                        variant="ghost"
                        aria-label={props.t("actions.remove")}
                        disabled={props.readOnly}
                        onClick={() =>
                          change({
                            ...column,
                            media: column.media.filter(
                              (entry) => entry.id !== item.id,
                            ),
                          })
                        }
                      >
                        <Trash2 />
                      </Button>
                    </div>
                  </div>
                ))}
              </section>
            </ScrollArea>
          </TabsContent>
          <TabsContent value="json" className="min-h-0 overflow-auto p-3">
            <JsonEditorField
              label={props.t("json")}
              hideLabel
              variables={props.variables ?? []}
              value={column.jsonDraft}
              error={column.jsonError || undefined}
              readOnly={props.readOnly}
              onValueChange={(jsonDraft) =>
                props.onChange({
                  ...column,
                  jsonDraft,
                  jsonError: "",
                  dirty: true,
                })
              }
              className="min-h-96"
            />
            <div className="mt-3 flex justify-end">
              <Button
                type="button"
                size="sm"
                disabled={props.readOnly}
                onClick={() => switchMode("visual")}
              >
                {props.t("actions.applyJson")}
              </Button>
            </div>
          </TabsContent>
        </Tabs>
      )}

      <div className="flex items-center gap-2 border-t border-border p-3 md:hidden">
        {props.result.status === "running" ? (
          <Button
            type="button"
            variant="destructive"
            className="flex-1"
            onClick={props.onStop}
          >
            <Square data-icon="inline-start" />
            {props.t("actions.stop")}
          </Button>
        ) : (
          <Button
            type="button"
            className="flex-1"
            aria-label={props.t("actions.runShortcut")}
            title={props.t("actions.runShortcut")}
            disabled={props.readOnly || !column.selectedModel}
            onClick={props.onRun}
          >
            <Play data-icon="inline-start" />
            {props.t("actions.run")}
          </Button>
        )}
      </div>
    </div>
  );
}

export function PlaygroundOutputPanel(props: PlaygroundColumnProps) {
  const { result } = props;
  const statusMap = {
    idle: "neutral",
    running: "pending",
    success: "success",
    stopped: "warning",
    error: "error",
  } as const;
  const telemetryItems = buildTelemetryItems(result, props.locale, props.t);
  const guardrailChecks = parseGuardrailChecks(
    result.telemetry?.guardrailStatus,
  );
  const guardrailsIntervened = guardrailChecks.some(
    (check) => check.status !== "passed",
  );

  return (
    <div className="flex h-full min-h-0 flex-col bg-card">
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <h2 className="text-sm font-semibold">{props.t("output")}</h2>
        <StatusBadge status={statusMap[result.status]}>
          {props.t(`status.${result.status}`)}
        </StatusBadge>
        {result.responseIsPrevious ? (
          <StatusBadge status="neutral">
            {props.t("status.previous")}
          </StatusBadge>
        ) : null}
        {result.telemetry?.cacheStatus === "hit" ? (
          <StatusBadge status="neutral">
            {props.t("status.cacheHit")}
          </StatusBadge>
        ) : null}
        {result.telemetry?.cacheStatus === "bypass" ? (
          <StatusBadge status="warning">
            {props.t("status.cacheBypass")}
          </StatusBadge>
        ) : null}
        {guardrailChecks.length ? (
          <StatusBadge status={guardrailsIntervened ? "warning" : "neutral"}>
            {props.t("status.guardrailsChecked", {
              count: guardrailChecks.length,
            })}
          </StatusBadge>
        ) : null}
        <div className="ms-auto flex items-center gap-1">
          {(props.mode === "chat" || props.mode === "layout_parsing") &&
          result.responseText ? (
            <CopyAction
              value={result.responseText}
              copyLabel={props.t("actions.copy")}
              copiedLabel={props.t("actions.copied")}
            />
          ) : null}
          {props.canOpenTrace && result.telemetry?.traceId ? (
            <Button type="button" variant="ghost" size="sm" asChild>
              <a
                href={`/traces/${encodeURIComponent(result.telemetry.traceId)}`}
              >
                {props.t("actions.openTrace")}
              </a>
            </Button>
          ) : null}
        </div>
      </div>

      {result.status === "idle" ? (
        <div className="flex flex-1 items-center justify-center p-6">
          <EmptyState
            variant="inline"
            icon={props.mode === "image_generation" ? ImageIcon : undefined}
            title={
              props.mode === "image_generation"
                ? props.t("image.emptyTitle")
                : props.mode === "layout_parsing"
                  ? props.t("ocr.emptyTitle")
                  : props.t("empty.title")
            }
            description={
              props.mode === "image_generation"
                ? props.t("image.emptyDescription")
                : props.mode === "layout_parsing"
                  ? props.t("ocr.emptyDescription")
                  : props.t("empty.description")
            }
          />
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
                {result.status === "running" &&
                !result.responseText &&
                result.generatedImages.length === 0 ? (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Spinner />
                    {props.t("status.running")}
                  </div>
                ) : null}
                {result.error ? (
                  <Alert variant="destructive">
                    <AlertDescription>{result.error}</AlertDescription>
                  </Alert>
                ) : null}
                {props.mode === "image_generation" &&
                result.generatedImages.length ? (
                  <GeneratedImageGallery
                    images={result.generatedImages}
                    t={props.t}
                  />
                ) : null}
                {props.mode === "chat" && result.responseText ? (
                  <MarkdownRenderer content={result.responseText} />
                ) : null}
                {props.mode === "layout_parsing" && result.responseText ? (
                  <MarkdownRenderer content={result.responseText} />
                ) : null}
                {props.mode === "chat" && result.toolCalls.length ? (
                  <div className="space-y-2">
                    <p className="text-xs font-medium text-muted-foreground">
                      {props.t("settings.tools")}
                    </p>
                    <CodeBlock
                      code={JSON.stringify(result.toolCalls, null, 2)}
                      language="json"
                    />
                  </div>
                ) : null}
              </div>
            </ScrollArea>
          </TabsContent>
          <TabsContent value="raw" className="min-h-0 overflow-auto p-4">
            <CodeBlock
              code={JSON.stringify(result.rawResponse ?? {}, null, 2)}
              language="json"
            />
          </TabsContent>
        </Tabs>
      )}

      {result.telemetry ? (
        <div className="border-t border-border p-3">
          <RunTelemetryStrip
            items={telemetryItems}
            unavailableLabel={props.t("telemetry.unavailable")}
            copyLabel={props.t("actions.copy")}
            copiedLabel={props.t("actions.copied")}
          />
        </div>
      ) : null}
      <div className="sr-only" role="status" aria-live="polite">
        {result.status === "running"
          ? props.t("streamAnnouncement.started")
          : result.status === "success"
            ? props.t("streamAnnouncement.completed")
            : result.status === "stopped"
              ? props.t("streamAnnouncement.stopped")
              : result.status === "error"
                ? props.t("streamAnnouncement.failed")
                : ""}
      </div>
    </div>
  );
}

function GeneratedImageGallery(props: {
  images: PlaygroundGeneratedImage[];
  t: Translation;
}) {
  return (
    <div className="grid gap-4 xl:grid-cols-2">
      {props.images.map((image) => (
        <GeneratedImageCard
          key={generatedImageKey(image)}
          image={image}
          t={props.t}
        />
      ))}
    </div>
  );
}

function GeneratedImageCard(props: {
  image: PlaygroundGeneratedImage;
  t: Translation;
}) {
  const { image, t } = props;
  const { objectURL, error } = useBase64ObjectURL(
    image.b64Json,
    image.mimeType,
  );
  const source = image.url ?? objectURL;
  const extension =
    image.mimeType === "image/jpeg"
      ? "jpg"
      : image.mimeType === "image/webp"
        ? "webp"
        : "png";
  const downloadName = `generated-image-${image.index + 1}.${extension}`;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="flex-row items-center justify-between gap-2 p-4">
        <div className="flex min-w-0 items-center gap-2">
          <CardTitle className="text-sm">
            {t("image.result", { count: image.index + 1 })}
          </CardTitle>
          <Badge variant="secondary" className="font-mono">
            {image.mimeType}
          </Badge>
        </div>
        {image.revisedPrompt ? (
          <CopyAction
            value={image.revisedPrompt}
            copyLabel={t("image.copyRevisedPrompt")}
            copiedLabel={t("actions.copied")}
          />
        ) : null}
      </CardHeader>
      <CardContent className="p-0">
        {source ? (
          // A native image element is required for transient blob URLs and provider URLs with unknown dimensions.
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={source}
            alt={t("image.alt", { count: image.index + 1 })}
            className="aspect-square w-full bg-muted/30 object-contain"
            referrerPolicy="no-referrer"
          />
        ) : (
          <div className="flex aspect-square items-center justify-center p-6">
            <Alert variant="destructive">
              <AlertDescription>
                {error === "too_large"
                  ? t("image.tooLarge")
                  : error === "invalid"
                    ? t("image.invalidBase64")
                    : t("image.decodeError")}
              </AlertDescription>
            </Alert>
          </div>
        )}
      </CardContent>
      {image.revisedPrompt ? (
        <CardContent className="p-4">
          <p className="text-xs font-medium text-muted-foreground">
            {t("image.revisedPrompt")}
          </p>
          <p className="mt-2 text-sm leading-6" dir="auto">
            {image.revisedPrompt}
          </p>
        </CardContent>
      ) : null}
      {source ? (
        <CardFooter className="justify-end p-4 pt-0">
          <Button type="button" size="sm" variant="outline" asChild>
            <a
              href={source}
              download={downloadName}
              target={image.url ? "_blank" : undefined}
              rel={image.url ? "noopener noreferrer" : undefined}
            >
              <Download data-icon="inline-start" />
              {t("image.download")}
            </a>
          </Button>
          <Button type="button" size="sm" variant="ghost" asChild>
            <a href={source} target="_blank" rel="noopener noreferrer">
              <ExternalLink data-icon="inline-start" />
              {t("image.open")}
            </a>
          </Button>
        </CardFooter>
      ) : null}
    </Card>
  );
}

function useBase64ObjectURL(
  base64: string | undefined,
  mimeType: PlaygroundGeneratedImage["mimeType"],
) {
  const [state, setState] = useState<{
    objectURL?: string;
    error?: "too_large" | "invalid";
  }>({});
  useEffect(() => {
    let objectURL = "";
    const frame = window.requestAnimationFrame(() => {
      const next = createBase64ObjectURL(base64, mimeType);
      objectURL = next.objectURL ?? "";
      setState(next);
    });
    return () => {
      window.cancelAnimationFrame(frame);
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [base64, mimeType]);
  return state;
}

function createBase64ObjectURL(
  base64: string | undefined,
  mimeType: PlaygroundGeneratedImage["mimeType"],
) {
  if (!base64) return {};
  const maxDecodedBytes = 10 * 1024 * 1024;
  if (base64.length > Math.ceil((maxDecodedBytes * 4) / 3) + 4)
    return { error: "too_large" as const };
  try {
    const binary = window.atob(base64);
    if (binary.length > maxDecodedBytes) return { error: "too_large" as const };
    const bytes = new Uint8Array(binary.length);
    for (let index = 0; index < binary.length; index += 1)
      bytes[index] = binary.charCodeAt(index);
    return {
      objectURL: URL.createObjectURL(new Blob([bytes], { type: mimeType })),
    };
  } catch {
    return { error: "invalid" as const };
  }
}

function generatedImageKey(image: PlaygroundGeneratedImage) {
  const base64Identity = image.b64Json
    ? `${image.b64Json.length}:${image.b64Json.slice(0, 16)}:${image.b64Json.slice(-16)}`
    : "";
  return `${image.index}:${image.mimeType}:${image.url ?? base64Identity}`;
}

function readFileAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () =>
      typeof reader.result === "string"
        ? resolve(reader.result)
        : reject(new Error("File could not be encoded."));
    reader.onerror = () =>
      reject(reader.error ?? new Error("File could not be read."));
    reader.readAsDataURL(file);
  });
}

export function PlaygroundComparisonColumn(props: PlaygroundColumnProps) {
  return (
    <div className="flex min-h-0 min-w-80 flex-1 basis-1/2 flex-col overflow-hidden rounded-lg border border-border bg-background">
      <div className="min-h-80 flex-1 overflow-hidden">
        <PlaygroundEditorPanel {...props} />
      </div>
      <div className="min-h-64 flex-1 overflow-hidden border-t border-border">
        <PlaygroundOutputPanel {...props} />
      </div>
    </div>
  );
}

function buildTelemetryItems(
  result: PlaygroundRunResult,
  locale: string,
  t: Translation,
): RunTelemetryItem[] {
  const telemetry = result.telemetry ?? {};
  const number = new Intl.NumberFormat(locale);
  const money = new Intl.NumberFormat(locale, { maximumFractionDigits: 6 });
  const creditAmount = new Intl.NumberFormat(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  const duration = (value?: number) =>
    value === undefined
      ? t("telemetry.unavailable")
      : `${number.format(value)} ms`;
  return [
    {
      key: "input",
      label: t("telemetry.inputTokens"),
      value:
        telemetry.inputTokens === undefined
          ? null
          : number.format(telemetry.inputTokens),
    },
    {
      key: "output",
      label: t("telemetry.outputTokens"),
      value:
        telemetry.outputTokens === undefined
          ? null
          : number.format(telemetry.outputTokens),
    },
    {
      key: "total",
      label: t("telemetry.totalTokens"),
      value:
        telemetry.totalTokens === undefined
          ? null
          : number.format(telemetry.totalTokens),
    },
    {
      key: "duration",
      label: t("telemetry.duration"),
      value: duration(telemetry.durationMs),
    },
    {
      key: "ttft",
      label: t("telemetry.ttft"),
      value: duration(telemetry.ttftMs),
    },
    {
      key: "usd",
      label: t("telemetry.costUSD"),
      value:
        telemetry.costUSD === undefined
          ? null
          : `$${money.format(telemetry.costUSD)}`,
    },
    {
      key: "credits",
      label: t("telemetry.credits"),
      value:
        telemetry.credits === undefined
          ? null
          : creditAmount.format(Number(telemetry.credits)),
    },
    {
      key: "model",
      label: t("telemetry.model"),
      value: telemetry.effectiveModel ?? telemetry.requestedModel ?? null,
      technical: true,
    },
    {
      key: "cache",
      label: t("telemetry.cache"),
      value: telemetry.cacheStatus ?? null,
    },
    {
      key: "cache-similarity",
      label: t("telemetry.cacheSimilarity"),
      value:
        telemetry.cacheSimilarity === undefined
          ? null
          : number.format(telemetry.cacheSimilarity),
    },
    {
      key: "guardrails",
      label: t("telemetry.guardrails"),
      value:
        parseGuardrailChecks(telemetry.guardrailStatus)
          .map((check) =>
            t("telemetry.guardrailCheck", {
              hook: check.hook,
              slug: check.slug,
              status: check.status,
            }),
          )
          .join(" · ") || null,
      technical: true,
    },
    {
      key: "finish",
      label: t("telemetry.finishReason"),
      value: result.finishReason ?? null,
      technical: true,
    },
    ...(telemetry.imageCount === undefined
      ? []
      : ([
          {
            key: "image-count",
            label: t("telemetry.imageCount"),
            value: number.format(telemetry.imageCount),
          },
          {
            key: "image-size",
            label: t("telemetry.imageSize"),
            value: telemetry.requestedImageSize ?? null,
            technical: true,
          },
          {
            key: "image-quality",
            label: t("telemetry.imageQuality"),
            value: telemetry.requestedImageQuality ?? null,
            technical: true,
          },
          {
            key: "image-mime",
            label: t("telemetry.imageMimeType"),
            value: telemetry.imageMimeType ?? null,
            technical: true,
          },
        ] satisfies RunTelemetryItem[])),
    {
      key: "trace",
      label: t("telemetry.trace"),
      value: telemetry.traceId ?? telemetry.requestId ?? null,
      rawValue: telemetry.traceId ?? telemetry.requestId,
      technical: true,
    },
  ];
}
