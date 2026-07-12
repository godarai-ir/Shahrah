"use client";

import { AlertCircle, Bot, Braces, DatabaseZap, Play, Plus, Save, Settings2, ShieldCheck, Sparkles, UserRound } from "lucide-react";
import { useMemo, useState } from "react";

import { AppSwitchField, AppTextField, HelpLink, IdentifierField, PageHeader } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";

import { CacheConfigDialog } from "./cache-config-dialog";
import { CodeBlock } from "./code-block";
import { GuardrailDialog } from "./guardrail-dialog";
import { ModelPickerDialog } from "./model-picker-dialog";
import { PromptVersionsPanel } from "./prompt-versions-panel";
import { ConfigSummaryPills } from "./prompt-ui";
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
  docsHref?: string;
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
  const pageDir = props.locale === "fa" ? "rtl" : "ltr";

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
  const selectedProviderLabel = useMemo(() => {
    const key = (selectedModelMetadata.model?.provider || "").trim().toLowerCase();
    return providerOptions.find((item) => item.id === key)?.label ?? "";
  }, [providerOptions, selectedModelMetadata.model?.provider]);

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
  const showVersionsPanel = props.mode === "edit" || props.versionsLoading || props.versions.length > 0;
  const roleOptions = [
    { value: "user" as const, label: props.t("roles.user"), description: props.t("builder.roleDescriptions.user"), icon: UserRound },
    { value: "assistant" as const, label: props.t("roles.assistant"), description: props.t("builder.roleDescriptions.assistant"), icon: Bot },
  ];

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_22rem]">
      <div className="flex min-w-0 flex-col gap-6">
        <PageHeader
          title={props.mode === "edit" ? props.t("builder.titleExisting") : props.t("builder.titleNew")}
          titleClassName="text-xl md:text-2xl"
          description={(
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
              <span>{props.t("builder.description")}</span>
              {props.docsHref ? <HelpLink href={props.docsHref} /> : null}
            </div>
          )}
          actions={(
            <Button type="button" onClick={() => void props.onSavePrompt()} disabled={!props.canSave || props.isSaving}>
              {props.isSaving ? <Spinner /> : <Save />}
              {props.mode === "edit" ? props.t("actions.saveNewVersion") : props.t("actions.saveNewPrompt")}
            </Button>
          )}
        />

        {props.notice ? (
          <Alert variant={props.notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{props.notice.kind === "error" ? props.t("notice.errorTitle") : props.t("notice.successTitle")}</AlertTitle>
            <AlertDescription>{props.notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>{props.t("builder.sections.basicsTitle")}</CardTitle>
            <CardDescription>{props.t("builder.sections.basicsDescription")}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <div className="grid gap-4 md:grid-cols-2">
              <IdentifierField label={props.t("builder.fields.name")} value={props.builder.name} onChange={(event) => props.onBuilderChange((current) => ({ ...current, name: event.target.value }))} placeholder={props.t("builder.fields.namePlaceholder")} description={props.t("builder.fields.nameHelp")} />
              <AppTextField label={props.t("builder.fields.commitMessage")} value={props.builder.commitMessage} onChange={(event) => props.onBuilderChange((current) => ({ ...current, commitMessage: event.target.value }))} placeholder={props.t("builder.fields.commitMessagePlaceholder")} />
            </div>
            <AppTextField label={props.t("builder.fields.description")} value={props.builder.description} onChange={(event) => props.onBuilderChange((current) => ({ ...current, description: event.target.value }))} />
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>{props.t("builder.systemPrompt.title")}</CardTitle>
            <CardDescription>{props.t("builder.systemPrompt.description", { placeholder: "{{application}}" })}</CardDescription>
          </CardHeader>
          <CardContent>
            <Textarea
              value={props.builder.systemPrompt}
              onChange={(event) => props.onBuilderChange((current) => ({ ...current, systemPrompt: event.target.value }))}
              className="min-h-32"
              dir="auto"
            />
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>{props.t("builder.conversation.title")}</CardTitle>
            <CardDescription>{props.t("builder.conversation.description")}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {props.builder.messages.length === 0 ? (
              <div className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">{props.t("builder.conversation.empty")}</div>
            ) : (
              props.builder.messages.map((message, index) => (
                <div key={`message-${index}`} className="flex flex-col gap-3 rounded-xl border p-4">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <RadioGroup
                      dir={pageDir}
                      value={message.role}
                      onValueChange={(value) => props.onUpdateMessage(index, { role: value as "user" | "assistant" })}
                      className="grid w-full gap-2 sm:grid-cols-2"
                    >
                      {roleOptions.map((option) => {
                        const itemId = `message-role-${index}-${option.value}`;
                        const Icon = option.icon;
                        return (
                          <Label
                            key={option.value}
                            htmlFor={itemId}
                            className="flex cursor-pointer items-start gap-3 rounded-xl border border-border p-3 transition-colors hover:bg-accent/40 has-[[data-state=checked]]:border-primary has-[[data-state=checked]]:bg-primary/5"
                          >
                            <RadioGroupItem id={itemId} value={option.value} className="sr-only" />
                            <div className="flex size-9 shrink-0 items-center justify-center rounded-lg border bg-muted/40 text-muted-foreground has-[[data-state=checked]]:border-primary/20 has-[[data-state=checked]]:bg-primary/10 has-[[data-state=checked]]:text-primary">
                              <Icon className="size-4" />
                            </div>
                            <div className="flex min-w-0 flex-col gap-1">
                              <span className="text-sm font-medium">{option.label}</span>
                              <span className="text-xs text-muted-foreground">{option.description}</span>
                            </div>
                          </Label>
                        );
                      })}
                    </RadioGroup>
                    <Button type="button" variant="outline" size="sm" onClick={() => props.onRemoveMessage(index)}>
                      {props.t("actions.remove")}
                    </Button>
                  </div>
                  <Textarea
                    value={message.content}
                    onChange={(event) => props.onUpdateMessage(index, { content: event.target.value })}
                    className="min-h-28"
                    dir="auto"
                  />
                </div>
              ))
            )}

            <div className="flex flex-col gap-4 rounded-xl border border-dashed bg-muted/20 p-4">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <RadioGroup
                  dir={pageDir}
                  value={props.builder.composerRole}
                  onValueChange={(value) => props.onBuilderChange((current) => ({ ...current, composerRole: value as "user" | "assistant" }))}
                  className="grid w-full gap-2 sm:grid-cols-2"
                >
                  {roleOptions.map((option) => {
                    const itemId = `composer-role-${option.value}`;
                    const Icon = option.icon;
                    return (
                      <Label
                        key={option.value}
                        htmlFor={itemId}
                        className="flex cursor-pointer items-start gap-3 rounded-xl border border-border bg-background p-3 transition-colors hover:bg-accent/40 has-[[data-state=checked]]:border-primary has-[[data-state=checked]]:bg-primary/5"
                      >
                        <RadioGroupItem id={itemId} value={option.value} className="sr-only" />
                        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg border bg-muted/40 text-muted-foreground has-[[data-state=checked]]:border-primary/20 has-[[data-state=checked]]:bg-primary/10 has-[[data-state=checked]]:text-primary">
                          <Icon className="size-4" />
                        </div>
                        <div className="flex min-w-0 flex-col gap-1">
                          <span className="text-sm font-medium">{option.label}</span>
                          <span className="text-xs text-muted-foreground">{option.description}</span>
                        </div>
                      </Label>
                    );
                  })}
                </RadioGroup>
                <Button type="button" variant="outline" onClick={props.onAddComposerMessage} disabled={!props.builder.composerText.trim()}>
                  <Plus />
                  {props.t("actions.addMessage")}
                </Button>
              </div>
              <Textarea
                value={props.builder.composerText}
                onChange={(event) => props.onBuilderChange((current) => ({ ...current, composerText: event.target.value }))}
                placeholder={props.t("builder.conversation.composerPlaceholder")}
                className="min-h-28"
                dir="rtl"
              />
            </div>
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader className="gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0">
              <CardTitle>{props.t("playground.title")}</CardTitle>
              <CardDescription>{props.t("playground.description")}</CardDescription>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <Button type="button" onClick={() => void props.onRunPrompt()} disabled={props.isRunning || !props.builder.selectedModel}>
                {props.isRunning ? <Spinner /> : <Play />}
                {props.t("actions.run")}
              </Button>
            </div>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <div className="rounded-xl border bg-muted/20 p-3">
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
          </CardContent>
        </Card>
      </div>

      <div className="flex min-w-0 flex-col gap-6">
        <Card className="shadow-none">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Settings2 className="size-4" />
              {props.t("config.title")}
            </CardTitle>
            <CardDescription>{props.t("config.description")}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-6">
            <section className="flex flex-col gap-3">
              <div className="flex flex-col gap-3 rounded-xl border bg-muted/20 p-4">
                <div className="flex flex-col gap-1">
                  <div className="text-sm font-medium">{props.t("config.model.title")}</div>
                  <div className="text-sm text-muted-foreground">{props.builder.selectedModel || props.t("config.model.placeholder")}</div>
                </div>
                <Button type="button" variant="outline" className="w-full sm:w-fit" onClick={() => setModelPickerOpen(true)}>
                  {props.promptCatalog.isModelsLoading ? <Spinner /> : <Sparkles />}
                  {props.t("actions.selectModel")}
                </Button>
              </div>
              {selectedModelMetadata.catalog ? (
                <div className="flex flex-col gap-3 rounded-xl border p-4 text-sm">
                  <div className="font-medium">{selectedModelMetadata.catalog.display_name}</div>
                  <div className="flex flex-wrap gap-2">
                    {selectedProviderLabel ? <Badge variant="secondary">{selectedProviderLabel}</Badge> : null}
                    {selectedModelMetadata.catalog.context_window ? (
                      <Badge variant="outline">
                        {props.t("config.model.contextBadge", { count: formatNumber(selectedModelMetadata.catalog.context_window, props.locale) })}
                      </Badge>
                    ) : null}
                  </div>
                </div>
              ) : null}
            </section>

            <Separator />

            <section className="flex flex-col gap-3">
              <div className="text-sm font-medium">{props.t("config.modelSettings.title")}</div>
              <div className="grid gap-3 sm:grid-cols-2">
                <AppSwitchField label={props.t("config.modelSettings.stream")} description={props.builder.stream ? props.t("common.enabled") : props.t("common.disabled")} checked={props.builder.stream} onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, stream: checked }))} />
                <AppTextField label={props.t("config.modelSettings.temperature")} value={props.builder.temperature} onChange={(event) => props.onBuilderChange((current) => ({ ...current, temperature: event.target.value }))} />
                <AppTextField label={props.t("config.modelSettings.topP")} value={props.builder.topP} onChange={(event) => props.onBuilderChange((current) => ({ ...current, topP: event.target.value }))} />
                <AppTextField label={props.t("config.modelSettings.maxTokens")} value={props.builder.maxTokens} onChange={(event) => props.onBuilderChange((current) => ({ ...current, maxTokens: event.target.value }))} />
              </div>
              {props.builder.stream && props.builder.outputGuardrails.length > 0 ? (
                <Alert variant="destructive">
                  <AlertCircle className="size-4" />
                  <AlertTitle>{props.t("validation.streamingConflictTitle")}</AlertTitle>
                  <AlertDescription>{props.t("validation.outputGuardrailsRequireStreamingOff")}</AlertDescription>
                </Alert>
              ) : null}
            </section>
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>{props.t("config.variables.title", { count: props.detectedVariables.length })}</CardTitle>
            <CardDescription>{props.t("config.variables.detectedHelp", { placeholder: "{{variable_name}}" })}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            {props.detectedVariables.length === 0 ? (
              <div className="rounded-lg border border-dashed p-3 text-xs text-muted-foreground">{props.t("config.variables.empty")}</div>
            ) : (
              props.detectedVariables.map((variable) => (
                <AppTextField
                    key={variable}
                    label={variable}
                    value={props.builder.variableValues[variable] ?? ""}
                    onChange={(event) =>
                      props.onBuilderChange((current) => ({
                        ...current,
                        variableValues: { ...current.variableValues, [variable]: event.target.value },
                      }))
                    }
                    placeholder={props.t("config.variables.valuePlaceholder", { name: variable })}
                  />
              ))
            )}
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>{props.t("config.advanced.title")}</CardTitle>
            <CardDescription>{props.t("config.advanced.description")}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-6">
            <section className="flex flex-col gap-3">
              <div className="flex items-start justify-between gap-3">
                <div className="flex min-w-0 flex-col gap-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <div className="text-sm font-medium">{props.t("config.cache.title")}</div>
                    {!props.builder.cacheEnabled ? <Badge variant="outline">{props.t("common.disabled")}</Badge> : null}
                  </div>
                  <div className="text-xs text-muted-foreground">{props.t("config.cache.description")}</div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  aria-label={props.t("actions.configure")}
                  title={props.t("actions.configure")}
                  onClick={() => setCacheDialogOpen(true)}
                >
                  <DatabaseZap />
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
                    : []
                }
              />
            </section>

            <Separator />

            <section className="flex flex-col gap-3">
              <div className="flex items-start justify-between gap-3">
                <div className="flex min-w-0 flex-col gap-1">
                  <div className="text-sm font-medium">{props.t("config.inputGuardrails.title", { count: props.builder.inputGuardrails.length })}</div>
                  <div className="text-xs text-muted-foreground">{props.t("config.inputGuardrails.description")}</div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  aria-label={props.t("actions.select")}
                  title={props.t("actions.select")}
                  onClick={() => setGuardrailDialogMode("input")}
                >
                  <ShieldCheck />
                </Button>
              </div>
              <ConfigSummaryPills pills={inputGuardrailItems.map((item) => item.name)} emptyLabel={props.t("config.inputGuardrails.empty")} />
            </section>

            <Separator />

            <section className="flex flex-col gap-3">
              <div className="flex items-start justify-between gap-3">
                <div className="flex min-w-0 flex-col gap-1">
                  <div className="text-sm font-medium">{props.t("config.outputGuardrails.title", { count: props.builder.outputGuardrails.length })}</div>
                  <div className="text-xs text-muted-foreground">{props.t("config.outputGuardrails.description")}</div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  aria-label={props.t("actions.select")}
                  title={props.t("actions.select")}
                  onClick={() => setGuardrailDialogMode("output")}
                >
                  <ShieldCheck />
                </Button>
              </div>
              <ConfigSummaryPills pills={outputGuardrailItems.map((item) => item.name)} emptyLabel={props.t("config.outputGuardrails.empty")} />
            </section>

            <Separator />

            <section className="flex flex-col gap-3">
              <div className="flex items-start justify-between gap-3">
                <div className="flex min-w-0 flex-col gap-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <div className="text-sm font-medium">{props.t("config.structuredOutput.title")}</div>
                    {!props.builder.structuredEnabled ? <Badge variant="outline">{props.t("common.disabled")}</Badge> : null}
                  </div>
                  <div className="text-xs text-muted-foreground">{props.t("config.structuredOutput.description")}</div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  aria-label={props.t("actions.configure")}
                  title={props.t("actions.configure")}
                  onClick={() => setStructuredDialogOpen(true)}
                >
                  <Braces />
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
                    : []
                }
              />
            </section>
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader className="gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0">
              <CardTitle>{props.t("config.metadataSection.title")}</CardTitle>
              <CardDescription>{props.t("config.metadataSection.description")}</CardDescription>
            </div>
            <Button type="button" variant="outline" className="shrink-0" onClick={props.onAddMetadataRow}>
              <Plus />
              {props.t("actions.addMetadata")}
            </Button>
          </CardHeader>
          <CardContent className="flex flex-col gap-6">
            <section className="flex flex-col gap-3">
              <div className="text-sm font-medium">{props.t("config.metadata.title")}</div>
              <div className="flex flex-col gap-2">
                {props.builder.metadataRows.length === 0 ? (
                  <div className="rounded-lg border border-dashed p-3 text-xs text-muted-foreground">{props.t("config.metadata.empty")}</div>
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
              </div>
            </section>

            <Separator />

            <section className="flex flex-col gap-2">
              <div className="text-sm font-medium">{props.t("config.logging.title")}</div>
              <div className="rounded-xl border bg-muted/20 p-3 text-xs text-muted-foreground">{props.t("config.logging.description")}</div>
            </section>
          </CardContent>
        </Card>

        {showVersionsPanel ? (
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
        ) : null}
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
