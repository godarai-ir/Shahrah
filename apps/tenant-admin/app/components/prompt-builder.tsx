"use client";

import { AlertCircle, Braces, DatabaseZap, Loader2, Play, Plus, Save, Settings2, ShieldCheck, Sparkles } from "lucide-react";
import { useMemo, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";

import { CacheConfigDialog } from "./cache-config-dialog";
import { CodeBlock } from "./code-block";
import { GuardrailDialog } from "./guardrail-dialog";
import { ModelPickerDialog } from "./model-picker-dialog";
import { PromptVersionsPanel } from "./prompt-versions-panel";
import { ConfigSummaryPills, Field } from "./prompt-ui";
import { StructuredOutputDialog } from "./structured-output-dialog";
import {
  formatNumber,
  type BuilderState,
  type Notice,
  type PromptCatalogState,
  type PromptTranslationFn,
  type PromptRunResponse,
  type PromptVersion,
} from "./prompt-management-shared";

type PromptBuilderProps = {
  builder: BuilderState;
  canSave: boolean;
  detectedVariables: string[];
  isRunning: boolean;
  isSaving: boolean;
  locale: string;
  mode: "create" | "edit";
  notice: Notice;
  onAddComposerMessage: () => void;
  onAddMetadataRow: () => void;
  onBuilderChange: (updater: (current: BuilderState) => BuilderState) => void;
  onCopySelectedFqn: () => Promise<void>;
  onCopySelectedSnippet: () => Promise<void>;
  onRemoveMessage: (index: number) => void;
  onRemoveMetadataRow: (rowID: string) => void;
  onRunPrompt: () => Promise<void>;
  onSavePrompt: () => Promise<void>;
  onSelectVersion: (version: PromptVersion) => void;
  onUpdateMessage: (index: number, patch: Partial<{ role: "user" | "assistant"; content: string }>) => void;
  onUpdateMetadataRow: (rowID: string, patch: Partial<{ key: string; value: string }>) => void;
  promptCatalog: PromptCatalogState;
  runPreview: string;
  runResponse: PromptRunResponse | null;
  selectedPromptName?: string;
  selectedVersion: PromptVersion | null;
  selectedVersionID: string;
  t: PromptTranslationFn;
  versionsLoading: boolean;
  versions: PromptVersion[];
};

export function PromptBuilder(props: PromptBuilderProps) {
  const [modelPickerOpen, setModelPickerOpen] = useState(false);
  const [cacheDialogOpen, setCacheDialogOpen] = useState(false);
  const [guardrailDialogMode, setGuardrailDialogMode] = useState<"input" | "output" | null>(null);
  const [structuredDialogOpen, setStructuredDialogOpen] = useState(false);

  const providerOptions = useMemo(() => {
    const grouped = new Set(props.promptCatalog.models.map((item) => (item.provider || "other").trim().toLowerCase()).filter(Boolean));
    return [...grouped].sort((left, right) => left.localeCompare(right)).map((key) => {
      const providerInfo = props.promptCatalog.providerCatalog.find((item) => item.provider.trim().toLowerCase() === key);
      return {
        id: key,
        label: providerInfo?.display_name || key,
        description: providerInfo?.description || "",
      };
    });
  }, [props.promptCatalog.models, props.promptCatalog.providerCatalog]);

  const selectedModelMetadata = useMemo(() => {
    const entry = props.promptCatalog.models.find((item) => item.id === props.builder.selectedModel);
    const catalogEntry =
      props.promptCatalog.providerModels.find((item) => item.model === props.builder.selectedModel)
      ?? props.promptCatalog.providerModels.find((item) => (entry?.aliases ?? []).includes(item.model));
    return {
      model: entry,
      catalog: catalogEntry ?? null,
    };
  }, [props.builder.selectedModel, props.promptCatalog.models, props.promptCatalog.providerModels]);

  const inputGuardrailItems = useMemo(
    () => props.promptCatalog.guardrails.filter((item) => props.builder.inputGuardrails.includes(item.selector || `guardrails/${item.slug}`)),
    [props.builder.inputGuardrails, props.promptCatalog.guardrails],
  );
  const outputGuardrailItems = useMemo(
    () => props.promptCatalog.guardrails.filter((item) => props.builder.outputGuardrails.includes(item.selector || `guardrails/${item.slug}`)),
    [props.builder.outputGuardrails, props.promptCatalog.guardrails],
  );
  const runPreviewLanguage = useMemo<"json" | "markdown">(() => {
    const trimmed = props.runPreview.trim();
    if (!trimmed) return "markdown";
    if ((trimmed.startsWith("{") || trimmed.startsWith("["))) {
      try {
        JSON.parse(trimmed);
        return "json";
      } catch {
        return "markdown";
      }
    }
    return "markdown";
  }, [props.runPreview]);

  return (
    <div className="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
      <div className="space-y-5">
        {props.notice ? (
          <Alert variant={props.notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{props.notice.kind === "error" ? props.t("notice.errorTitle") : props.t("notice.successTitle")}</AlertTitle>
            <AlertDescription>{props.notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Settings2 className="size-4" />
              {props.t("config.title")}
            </CardTitle>
            <CardDescription>{props.t("config.description")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <section className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <div>
                  <div className="text-sm font-medium">{props.t("config.model.title")}</div>
                  <div className="text-xs text-muted-foreground">{props.builder.selectedModel || props.t("config.model.placeholder")}</div>
                </div>
                <Button type="button" variant="outline" onClick={() => setModelPickerOpen(true)}>
                  {props.promptCatalog.isModelsLoading ? <Loader2 className="animate-spin" /> : <Sparkles />}
                  {props.t("actions.selectModel")}
                </Button>
              </div>
              {selectedModelMetadata.catalog ? (
                <div className="rounded-lg border bg-muted/20 p-3 text-sm">
                  <div className="font-medium">{selectedModelMetadata.catalog.display_name}</div>
                  <div className="mt-1 flex flex-wrap gap-2">
                    {selectedModelMetadata.catalog.context_window ? (
                      <Badge variant="outline">
                        {props.t("config.model.contextBadge", { count: formatNumber(selectedModelMetadata.catalog.context_window, props.locale) })}
                      </Badge>
                    ) : null}
                    {(selectedModelMetadata.catalog.capabilities ?? []).slice(0, 4).map((capability) => (
                      <Badge key={capability} variant="secondary">
                        {capability}
                      </Badge>
                    ))}
                  </div>
                </div>
              ) : null}
            </section>

            <Separator />

            <section className="space-y-3">
              <div className="text-sm font-medium">{props.t("config.modelSettings.title")}</div>
              <div className="grid gap-3 sm:grid-cols-2">
                <Field label={props.t("config.modelSettings.stream")}>
                  <div className="flex h-10 items-center justify-between rounded-md border px-3">
                    <span className="text-sm text-muted-foreground">{props.builder.stream ? props.t("common.enabled") : props.t("common.disabled")}</span>
                    <Switch checked={props.builder.stream} onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, stream: checked }))} />
                  </div>
                </Field>
                <Field label={props.t("config.modelSettings.temperature")}>
                  <Input value={props.builder.temperature} onChange={(event) => props.onBuilderChange((current) => ({ ...current, temperature: event.target.value }))} />
                </Field>
                <Field label={props.t("config.modelSettings.topP")}>
                  <Input value={props.builder.topP} onChange={(event) => props.onBuilderChange((current) => ({ ...current, topP: event.target.value }))} />
                </Field>
                <Field label={props.t("config.modelSettings.maxTokens")}>
                  <Input value={props.builder.maxTokens} onChange={(event) => props.onBuilderChange((current) => ({ ...current, maxTokens: event.target.value }))} />
                </Field>
              </div>
              {props.builder.stream && props.builder.outputGuardrails.length > 0 ? (
                <Alert variant="destructive">
                  <AlertCircle className="size-4" />
                  <AlertTitle>{props.t("validation.streamingConflictTitle")}</AlertTitle>
                  <AlertDescription>{props.t("validation.outputGuardrailsRequireStreamingOff")}</AlertDescription>
                </Alert>
              ) : null}
            </section>

            <Separator />

            <section className="space-y-3">
              <div className="text-sm font-medium">{props.t("config.variables.title", { count: props.detectedVariables.length })}</div>
              <div className="text-xs text-muted-foreground">
                {props.t("config.variables.detectedHelpPrefix")} <code>{"{{variable_name}}"}</code> {props.t("config.variables.detectedHelpSuffix")}
              </div>
              {props.detectedVariables.length === 0 ? (
                <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">{props.t("config.variables.empty")}</div>
              ) : (
                <div className="space-y-2">
                  {props.detectedVariables.map((variable) => (
                    <Field key={variable} label={variable}>
                      <Input
                        value={props.builder.variableValues[variable] ?? ""}
                        onChange={(event) =>
                          props.onBuilderChange((current) => ({
                            ...current,
                            variableValues: { ...current.variableValues, [variable]: event.target.value },
                          }))
                        }
                        placeholder={props.t("config.variables.valuePlaceholder", { name: variable })}
                      />
                    </Field>
                  ))}
                </div>
              )}
            </section>

            <Separator />

            <section className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <div>
                  <div className="text-sm font-medium">{props.t("config.cache.title")}</div>
                  <div className="text-xs text-muted-foreground">{props.t("config.cache.description")}</div>
                </div>
                <Button type="button" variant="outline" onClick={() => setCacheDialogOpen(true)}>
                  <DatabaseZap />
                  {props.t("actions.configure")}
                </Button>
              </div>
              <ConfigSummaryPills
                emptyLabel={props.t("config.cache.empty")}
                pills={
                  props.builder.cacheEnabled
                    ? [
                        props.t("config.cache.summary.type", { value: props.builder.cacheType === "smart" ? props.t("common.smart") : props.t("common.simple") }),
                        props.builder.cacheTTL ? props.t("config.cache.summary.ttl", { value: props.builder.cacheTTL }) : "",
                        props.builder.cacheNamespace ? props.t("config.cache.summary.namespace", { value: props.builder.cacheNamespace }) : "",
                        props.builder.cacheType === "smart" && props.builder.cacheSimilarityThreshold
                          ? props.t("config.cache.summary.similarity", { value: props.builder.cacheSimilarityThreshold })
                          : "",
                      ]
                    : [props.t("common.disabled")]
                }
              />
            </section>

            <Separator />

            <section className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <div>
                  <div className="text-sm font-medium">{props.t("config.inputGuardrails.title", { count: props.builder.inputGuardrails.length })}</div>
                  <div className="text-xs text-muted-foreground">{props.t("config.inputGuardrails.description")}</div>
                </div>
                <Button type="button" variant="outline" onClick={() => setGuardrailDialogMode("input")}>
                  <ShieldCheck />
                  {props.t("actions.select")}
                </Button>
              </div>
              <ConfigSummaryPills pills={inputGuardrailItems.map((item) => item.name)} emptyLabel={props.t("config.inputGuardrails.empty")} />
            </section>

            <Separator />

            <section className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <div>
                  <div className="text-sm font-medium">{props.t("config.outputGuardrails.title", { count: props.builder.outputGuardrails.length })}</div>
                  <div className="text-xs text-muted-foreground">{props.t("config.outputGuardrails.description")}</div>
                </div>
                <Button type="button" variant="outline" onClick={() => setGuardrailDialogMode("output")}>
                  <ShieldCheck />
                  {props.t("actions.select")}
                </Button>
              </div>
              <ConfigSummaryPills pills={outputGuardrailItems.map((item) => item.name)} emptyLabel={props.t("config.outputGuardrails.empty")} />
            </section>

            <Separator />

            <section className="space-y-3">
              <div className="flex items-center justify-between gap-2">
                <div>
                  <div className="text-sm font-medium">{props.t("config.structuredOutput.title")}</div>
                  <div className="text-xs text-muted-foreground">{props.t("config.structuredOutput.description")}</div>
                </div>
                <Button type="button" variant="outline" onClick={() => setStructuredDialogOpen(true)}>
                  <Braces />
                  {props.t("actions.configure")}
                </Button>
              </div>
              <ConfigSummaryPills
                emptyLabel={props.t("config.structuredOutput.empty")}
                pills={
                  props.builder.structuredEnabled
                    ? [
                        props.builder.structuredFormat === "json_schema" ? props.t("common.jsonSchema") : props.t("common.jsonObject"),
                        props.builder.structuredStrict ? props.t("common.strict") : props.t("common.nonStrict"),
                      ]
                    : [props.t("common.disabled")]
                }
              />
            </section>

            <Separator />

            <section className="space-y-3">
              <div className="text-sm font-medium">{props.t("config.metadata.title")}</div>
              <div className="space-y-2">
                {props.builder.metadataRows.length === 0 ? (
                  <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">{props.t("config.metadata.empty")}</div>
                ) : null}
                {props.builder.metadataRows.map((row) => (
                  <div key={row.id} className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
                    <Input value={row.key} placeholder={props.t("config.metadata.keyPlaceholder")} onChange={(event) => props.onUpdateMetadataRow(row.id, { key: event.target.value })} />
                    <Input value={row.value} placeholder={props.t("config.metadata.valuePlaceholder")} onChange={(event) => props.onUpdateMetadataRow(row.id, { value: event.target.value })} />
                    <Button type="button" variant="outline" onClick={() => props.onRemoveMetadataRow(row.id)}>
                      {props.t("actions.remove")}
                    </Button>
                  </div>
                ))}
                <Button type="button" variant="outline" onClick={props.onAddMetadataRow}>
                  <Plus />
                  {props.t("actions.addMetadata")}
                </Button>
              </div>
            </section>

            <Separator />

            <section className="space-y-2">
              <div className="text-sm font-medium">{props.t("config.logging.title")}</div>
              <div className="rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground">{props.t("config.logging.description")}</div>
            </section>
          </CardContent>
        </Card>

        <PromptVersionsPanel
          isLoading={props.versionsLoading}
          onCopyFqn={() => void props.onCopySelectedFqn()}
          onCopySnippet={() => void props.onCopySelectedSnippet()}
          onSelectVersion={props.onSelectVersion}
          selectedVersion={props.selectedVersion}
          selectedVersionID={props.selectedVersionID}
          t={props.t}
          versions={props.versions}
        />
      </div>

      <div className="space-y-5">
        <Card className="shadow-none">
          <CardHeader className="border-b">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <div className="text-xs text-muted-foreground">
                  {props.mode === "edit" && props.selectedPromptName
                    ? props.t("builder.breadcrumbExisting", { name: props.selectedPromptName })
                    : props.t("builder.breadcrumbNew")}
                </div>
                <CardTitle>{props.mode === "edit" ? props.t("builder.titleExisting") : props.t("builder.titleNew")}</CardTitle>
                <CardDescription>{props.t("builder.description")}</CardDescription>
              </div>
              <Button type="button" onClick={() => void props.onSavePrompt()} disabled={!props.canSave || props.isSaving}>
                {props.isSaving ? <Loader2 className="animate-spin" /> : <Save />}
                {props.mode === "edit" ? props.t("actions.saveNewVersion") : props.t("actions.saveNewPrompt")}
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-5 p-4">
            <div className="grid gap-3 md:grid-cols-2">
              <Field label={props.t("builder.fields.name")}>
                <Input value={props.builder.name} onChange={(event) => props.onBuilderChange((current) => ({ ...current, name: event.target.value }))} placeholder={props.t("builder.fields.namePlaceholder")} />
              </Field>
              <Field label={props.t("builder.fields.commitMessage")}>
                <Input
                  value={props.builder.commitMessage}
                  onChange={(event) => props.onBuilderChange((current) => ({ ...current, commitMessage: event.target.value }))}
                  placeholder={props.t("builder.fields.commitMessagePlaceholder")}
                />
              </Field>
            </div>
            <Field label={props.t("builder.fields.description")}>
              <Input value={props.builder.description} onChange={(event) => props.onBuilderChange((current) => ({ ...current, description: event.target.value }))} />
            </Field>

            <Card className="shadow-none">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">{props.t("builder.systemPrompt.title")}</CardTitle>
                <CardDescription>
                  {props.t("builder.systemPrompt.descriptionPrefix")} <code>{"{{application}}"}</code> {props.t("builder.systemPrompt.descriptionSuffix")}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Textarea
                  value={props.builder.systemPrompt}
                  onChange={(event) => props.onBuilderChange((current) => ({ ...current, systemPrompt: event.target.value }))}
                  className="min-h-28"
                  dir="auto"
                />
              </CardContent>
            </Card>

            <Card className="shadow-none">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">{props.t("builder.conversation.title")}</CardTitle>
                <CardDescription>{props.t("builder.conversation.description")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {props.builder.messages.length === 0 ? (
                  <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">{props.t("builder.conversation.empty")}</div>
                ) : (
                  props.builder.messages.map((message, index) => (
                    <div key={`message-${index}`} className="rounded-lg border p-3">
                      <div className="mb-2 flex items-center justify-between gap-3">
                        <select
                          value={message.role}
                          onChange={(event) => props.onUpdateMessage(index, { role: event.target.value as "user" | "assistant" })}
                          className="h-9 rounded-md border bg-background px-2 text-sm"
                        >
                          <option value="user">{props.t("roles.user")}</option>
                          <option value="assistant">{props.t("roles.assistant")}</option>
                        </select>
                        <Button type="button" variant="outline" size="sm" onClick={() => props.onRemoveMessage(index)}>
                          {props.t("actions.remove")}
                        </Button>
                      </div>
                      <Textarea
                        value={message.content}
                        onChange={(event) => props.onUpdateMessage(index, { content: event.target.value })}
                        className="min-h-24"
                        dir="auto"
                      />
                    </div>
                  ))
                )}

                <div className="rounded-lg border bg-muted/20 p-3">
                  <div className="mb-2 flex items-center gap-2">
                    <select
                      value={props.builder.composerRole}
                      onChange={(event) => props.onBuilderChange((current) => ({ ...current, composerRole: event.target.value as "user" | "assistant" }))}
                      className="h-9 rounded-md border bg-background px-2 text-sm"
                    >
                      <option value="user">{props.t("roles.user")}</option>
                      <option value="assistant">{props.t("roles.assistant")}</option>
                    </select>
                    <Button type="button" variant="outline" onClick={props.onAddComposerMessage} disabled={!props.builder.composerText.trim()}>
                      <Plus />
                      {props.t("actions.addMessage")}
                    </Button>
                  </div>
                  <Textarea
                    value={props.builder.composerText}
                    onChange={(event) => props.onBuilderChange((current) => ({ ...current, composerText: event.target.value }))}
                    placeholder={props.t("builder.conversation.composerPlaceholder")}
                    className="min-h-24"
                    dir="auto"
                  />
                </div>
              </CardContent>
            </Card>

            <Card className="shadow-none">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">{props.t("playground.title")}</CardTitle>
                <CardDescription>{props.t("playground.description")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="rounded-lg border bg-muted/20 p-3">
                  {props.runPreview ? (
                    <CodeBlock code={props.runPreview} language={runPreviewLanguage} maxHeightClassName="max-h-[min(24rem,46dvh)]" />
                  ) : (
                    <div className="text-sm text-muted-foreground">{props.t("playground.empty")}</div>
                  )}
                </div>
                {props.runResponse?.usage ? (
                  <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                    <Badge variant="outline">{props.t("playground.usage.prompt", { count: props.runResponse.usage.prompt_tokens ?? 0 })}</Badge>
                    <Badge variant="outline">{props.t("playground.usage.completion", { count: props.runResponse.usage.completion_tokens ?? 0 })}</Badge>
                    <Badge variant="outline">{props.t("playground.usage.total", { count: props.runResponse.usage.total_tokens ?? 0 })}</Badge>
                  </div>
                ) : null}
                <div className="flex flex-wrap gap-2">
                  <Button type="button" onClick={() => void props.onRunPrompt()} disabled={props.isRunning || !props.builder.selectedModel}>
                    {props.isRunning ? <Loader2 className="animate-spin" /> : <Play />}
                    {props.t("actions.run")}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </CardContent>
        </Card>
      </div>

      <ModelPickerDialog
        builder={props.builder}
        locale={props.locale}
        models={props.promptCatalog.models}
        onBuilderChange={props.onBuilderChange}
        open={modelPickerOpen}
        providerCatalog={providerOptions}
        providerModels={props.promptCatalog.providerModels}
        setOpen={setModelPickerOpen}
        t={props.t}
      />

      <CacheConfigDialog builder={props.builder} onBuilderChange={props.onBuilderChange} open={cacheDialogOpen} setOpen={setCacheDialogOpen} t={props.t} />
      <GuardrailDialog
        builder={props.builder}
        guardrails={props.promptCatalog.guardrails}
        isLoading={props.promptCatalog.isGuardrailsLoading}
        mode={guardrailDialogMode}
        onBuilderChange={props.onBuilderChange}
        setMode={setGuardrailDialogMode}
        t={props.t}
      />
      <StructuredOutputDialog builder={props.builder} onBuilderChange={props.onBuilderChange} open={structuredDialogOpen} setOpen={setStructuredDialogOpen} t={props.t} />
    </div>
  );
}
