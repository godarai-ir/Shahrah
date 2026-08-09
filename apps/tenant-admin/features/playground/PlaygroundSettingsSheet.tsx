"use client";

import * as React from "react";
import { CircleHelp, Plus, Trash2, Wrench } from "lucide-react";

import { AppCodeTextareaField } from "@/components/app-ui/AppCodeTextareaField";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { HelpLink } from "@/components/app-ui/HelpLink";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { ButtonGroup } from "@/components/ui/button-group";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldContent, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

import {
  createPlaygroundMetadataRow,
  createPlaygroundTool,
  type PlaygroundColumnDraft,
  type PlaygroundMode,
} from "./domain";
import type { GuardrailItem } from "@/lib/prompt-builder";

type Translation = (key: string, values?: Record<string, string | number>) => string;

export type PlaygroundCacheEntitlements = {
  features?: Record<string, unknown>;
} | null;

type PlaygroundSettingsSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  column: PlaygroundColumnDraft | null;
  mode: PlaygroundMode;
  guardrails: GuardrailItem[];
  guardrailsLoading: boolean;
  cacheEntitlements: PlaygroundCacheEntitlements;
  readOnly: boolean;
  side: "left" | "right";
  onChange: (column: PlaygroundColumnDraft) => void;
  t: Translation;
};

export function PlaygroundSettingsSheet(props: PlaygroundSettingsSheetProps) {
  if (!props.column) return null;
  const column = props.column;
  const config = column.config;
  const cacheLimits = resolveCacheLimits(props.cacheEntitlements);
  const cacheUnavailable = !cacheLimits.simpleEnabled && !cacheLimits.semanticEnabled;
  const showCachePlanLimit = cacheLimits.resolved && (!cacheLimits.simpleEnabled || !cacheLimits.semanticEnabled);
  const selectedCacheTypeUnavailable = config.cacheType === "simple" ? !cacheLimits.simpleEnabled : !cacheLimits.semanticEnabled;
  const cacheConfigDisabled = props.readOnly || cacheUnavailable || selectedCacheTypeUnavailable;
  const simpleCacheMaxTTL = config.cacheType === "simple" && cacheLimits.simpleTTLDays !== null
    ? cacheLimits.simpleTTLDays * 86_400
    : undefined;

  function changeConfig(patch: Partial<typeof config>) {
    props.onChange({ ...column, config: { ...config, ...patch }, dirty: true });
  }

  function toggleGuardrail(kind: "inputGuardrails" | "outputGuardrails", selector: string, checked: boolean) {
    const current = config[kind];
    changeConfig({ [kind]: checked ? [...current, selector] : current.filter((item) => item !== selector) });
  }

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent side={props.side} closeLabel={props.t("actions.close")} className="flex w-full flex-col gap-0 p-0 sm:max-w-xl" dir={props.side === "left" ? "rtl" : "ltr"}>
        <SheetHeader className="border-b border-border p-4 text-start">
          <SheetTitle>{props.t("settings.title")}</SheetTitle>
          <SheetDescription>{props.t("settings.description")}</SheetDescription>
        </SheetHeader>
        {props.mode === "image_generation" ? (
          <ImageGenerationSettings column={column} readOnly={props.readOnly} onChange={props.onChange} t={props.t} />
        ) : <Tabs defaultValue="model" className="min-h-0 flex-1 gap-0">
          <div className="flex justify-center overflow-x-auto border-b border-border p-3">
            <TabsList asChild className="h-auto bg-transparent p-0">
              <ButtonGroup aria-label={props.t("settings.title")} className="mx-auto">
                <TabsTrigger value="model" className="h-8 flex-none rounded-lg border border-border bg-background px-3 data-[state=active]:bg-accent data-[state=active]:text-accent-foreground">{props.t("settings.model")}</TabsTrigger>
                <TabsTrigger value="tools" className="h-8 flex-none rounded-lg border border-border bg-background px-3 data-[state=active]:bg-accent data-[state=active]:text-accent-foreground">{props.t("settings.tools")}</TabsTrigger>
                <TabsTrigger value="godarai" className="h-8 flex-none rounded-lg border border-border bg-background px-3 data-[state=active]:bg-accent data-[state=active]:text-accent-foreground">{props.t("settings.godarai")}</TabsTrigger>
                <TabsTrigger value="guardrails" className="h-8 flex-none rounded-lg border border-border bg-background px-3 data-[state=active]:bg-accent data-[state=active]:text-accent-foreground">{props.t("settings.guardrails")}</TabsTrigger>
                <TabsTrigger value="metadata" className="h-8 flex-none rounded-lg border border-border bg-background px-3 data-[state=active]:bg-accent data-[state=active]:text-accent-foreground">{props.t("settings.metadata")}</TabsTrigger>
              </ButtonGroup>
            </TabsList>
          </div>
          <ScrollArea className="min-h-0 flex-1">
            <TabsContent value="model" className="flex flex-col gap-5 p-4">
              <AppSwitchField surface={false} label={props.t("settings.stream")} checked={config.stream} disabled={props.readOnly} onCheckedChange={(stream) => changeConfig({ stream })} />
              <div className="grid gap-4 sm:grid-cols-2">
                <AppNumberField label={props.t("settings.temperature")} value={config.temperature} step={0.1} disabled={props.readOnly} onChange={(event) => changeConfig({ temperature: event.target.value })} />
                <AppNumberField label={props.t("settings.topP")} value={config.topP} step={0.05} disabled={props.readOnly} onChange={(event) => changeConfig({ topP: event.target.value })} />
                <AppNumberField label={props.t("settings.maxTokens")} value={config.maxTokens} min={1} step={1} disabled={props.readOnly} onChange={(event) => changeConfig({ maxTokens: event.target.value })} />
                <AppSelectField label={props.t("settings.reasoning")} value={config.reasoningEffort || "default"} disabled={props.readOnly} onValueChange={(value) => changeConfig({ reasoningEffort: value === "default" ? "" : value })} options={[
                  { value: "default", label: props.t("settings.reasoningNone") }, { value: "low", label: "low" }, { value: "medium", label: "medium" }, { value: "high", label: "high" },
                ]} optionTextDir="ltr" />
              </div>
              <AppCodeTextareaField
                label={props.t("settings.stop")}
                description={<span className="flex items-start gap-1">
                  <span>{props.t("settings.stopDescription")}</span>
                  <TooltipProvider delayDuration={150}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button type="button" variant="ghost" size="icon-xs" aria-label={props.t("settings.stopHelpLabel")}><CircleHelp data-icon="inline-start" /></Button>
                      </TooltipTrigger>
                      <TooltipContent side="top" className="max-w-xs text-start">{props.t("settings.stopTooltip")}</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </span>}
                value={config.stop}
                disabled={props.readOnly}
                onChange={(event) => changeConfig({ stop: event.target.value })}
                className="min-h-28"
              />
              {config.stream && config.outputGuardrails.length ? (
                <Alert variant="warning"><AlertDescription className="flex flex-wrap items-center justify-between gap-2"><span>{props.t("settings.guardrailConflict")}</span><Button size="sm" variant="outline" onClick={() => changeConfig({ stream: false })}>{props.t("settings.disableStream")}</Button></AlertDescription></Alert>
              ) : null}
              <Separator />
              <AppSwitchField surface={false} label={props.t("settings.structuredEnabled")} checked={config.structuredEnabled} disabled={props.readOnly} onCheckedChange={(structuredEnabled) => changeConfig({ structuredEnabled })} />
              {config.structuredEnabled ? (
                <div className="flex flex-col gap-4">
                  <AppSelectField label={props.t("settings.structuredFormat")} value={config.structuredFormat} disabled={props.readOnly} onValueChange={(value) => changeConfig({ structuredFormat: value as "json_object" | "json_schema" })} options={[{ value: "json_schema", label: "JSON Schema" }, { value: "json_object", label: "JSON Object" }]} optionTextDir="ltr" />
                  {config.structuredFormat === "json_schema" ? <>
                    <AppTextField label={props.t("settings.structuredName")} value={config.structuredName} disabled={props.readOnly} dir="ltr" onChange={(event) => changeConfig({ structuredName: event.target.value })} />
                    <AppSwitchField surface={false} label={props.t("settings.structuredStrict")} checked={config.structuredStrict} disabled={props.readOnly} onCheckedChange={(structuredStrict) => changeConfig({ structuredStrict })} />
                    <AppCodeTextareaField label={props.t("settings.structuredSchema")} value={config.structuredSchema} disabled={props.readOnly} onChange={(event) => changeConfig({ structuredSchema: event.target.value })} />
                  </> : null}
                </div>
              ) : null}
            </TabsContent>

            <TabsContent value="tools" className="flex flex-col gap-4 p-4">
              <Alert variant="info"><AlertDescription>{props.t("settings.toolSafety")}</AlertDescription></Alert>
              <div className="flex justify-end"><Button type="button" size="sm" variant="outline" disabled={props.readOnly} onClick={() => props.onChange({ ...column, tools: [...column.tools, createPlaygroundTool()], dirty: true })}><Wrench data-icon="inline-start" />{props.t("actions.addTool")}</Button></div>
              {column.tools.length === 0 ? <EmptyState variant="inline" density="compact" title={props.t("settings.tools")} description={props.t("settings.toolSafety")} /> : null}
              {column.tools.map((tool, index) => (
                <section key={tool.id} className="flex flex-col gap-3 border-b border-border pb-5 last:border-0">
                  <div className="flex items-center justify-between"><h3 className="text-sm font-medium">{tool.name || `${props.t("settings.tools")} ${index + 1}`}</h3><Button type="button" size="icon-sm" variant="ghost" aria-label={props.t("actions.remove")} disabled={props.readOnly} onClick={() => props.onChange({ ...column, tools: column.tools.filter((entry) => entry.id !== tool.id), dirty: true })}><Trash2 data-icon="inline-start" /></Button></div>
                  <AppTextField label={props.t("settings.toolName")} value={tool.name} dir="ltr" disabled={props.readOnly} onChange={(event) => props.onChange({ ...column, tools: column.tools.map((entry) => entry.id === tool.id ? { ...entry, name: event.target.value } : entry), dirty: true })} />
                  <AppTextField label={props.t("settings.toolDescription")} value={tool.description} disabled={props.readOnly} onChange={(event) => props.onChange({ ...column, tools: column.tools.map((entry) => entry.id === tool.id ? { ...entry, description: event.target.value } : entry), dirty: true })} />
                  <AppCodeTextareaField label={props.t("settings.toolSchema")} value={tool.schema} disabled={props.readOnly} onChange={(event) => props.onChange({ ...column, tools: column.tools.map((entry) => entry.id === tool.id ? { ...entry, schema: event.target.value } : entry), dirty: true })} />
                </section>
              ))}
              {column.tools.length ? <div className="grid gap-4 sm:grid-cols-2"><AppSelectField label={props.t("settings.toolChoice")} value={config.toolChoice} disabled={props.readOnly} onValueChange={(toolChoice) => changeConfig({ toolChoice })} options={[{ value: "auto", label: "auto" }, { value: "none", label: "none" }, { value: "required", label: "required" }]} optionTextDir="ltr" /><AppSwitchField surface={false} label={props.t("settings.parallelTools")} checked={config.parallelToolCalls} disabled={props.readOnly} onCheckedChange={(parallelToolCalls) => changeConfig({ parallelToolCalls })} /></div> : null}
            </TabsContent>

            <TabsContent value="godarai" className="flex flex-col gap-5 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <h3 className="text-sm font-semibold">{props.t("settings.cache")}</h3>
                <HelpLink href="/docs/guides/caching" />
              </div>
              {showCachePlanLimit ? (
                <Alert variant="info">
                  <AlertTitle>{props.t("settings.cachePlanLimitTitle")}</AlertTitle>
                  <AlertDescription>{getCachePlanLimitMessage(cacheLimits, props.t)}</AlertDescription>
                </Alert>
              ) : null}
              <AppSwitchField surface={false} label={props.t("settings.cacheEnabled")} checked={config.cacheEnabled} disabled={props.readOnly || cacheUnavailable} onCheckedChange={(cacheEnabled) => changeConfig({ cacheEnabled })} />
              {config.cacheEnabled ? <div className="grid gap-4 sm:grid-cols-2">
                <AppSelectField label={props.t("settings.cacheType")} value={config.cacheType} disabled={props.readOnly || cacheUnavailable} onValueChange={(value) => changeConfig({ cacheType: value as "simple" | "smart" })} options={[{ value: "simple", label: props.t("settings.cacheSimple"), disabled: !cacheLimits.simpleEnabled }, { value: "smart", label: props.t("settings.cacheSmart"), disabled: !cacheLimits.semanticEnabled }]} />
                <AppNumberField label={props.t("settings.cacheTTL")} value={config.cacheTTL} min={1} max={simpleCacheMaxTTL} step={1} disabled={cacheConfigDisabled} onChange={(event) => changeConfig({ cacheTTL: event.target.value })} />
                <AppTextField label={props.t("settings.cacheNamespace")} value={config.cacheNamespace} dir="ltr" disabled={cacheConfigDisabled} onChange={(event) => changeConfig({ cacheNamespace: event.target.value })} />
                {config.cacheType === "smart" ? <AppNumberField label={props.t("settings.cacheSimilarity")} value={config.cacheSimilarity} min={0.8} max={1} step={0.01} disabled={cacheConfigDisabled} onChange={(event) => changeConfig({ cacheSimilarity: event.target.value })} /> : null}
              </div> : null}
            </TabsContent>

            <TabsContent value="guardrails" className="flex flex-col gap-5 p-4">
              <h3 className="text-sm font-semibold">{props.t("settings.guardrails")}</h3>
              <AppSelectField label={props.t("settings.guardrailScope")} value={config.guardrailScope} disabled={props.readOnly} onValueChange={(value) => changeConfig({ guardrailScope: value as "all" | "last" })} options={[{ value: "last", label: props.t("settings.scopeLast") }, { value: "all", label: props.t("settings.scopeAll") }]} />
              {props.guardrailsLoading ? <p className="text-sm text-muted-foreground">{props.t("load.loading")}</p> : null}
              <GuardrailChecklist title={props.t("settings.inputGuardrails")} emptyLabel={props.t("settings.noGuardrails")} guardrails={props.guardrails} selected={config.inputGuardrails} disabled={props.readOnly} onToggle={(selector, checked) => toggleGuardrail("inputGuardrails", selector, checked)} />
              <GuardrailChecklist title={props.t("settings.outputGuardrails")} emptyLabel={props.t("settings.noGuardrails")} guardrails={props.guardrails} selected={config.outputGuardrails} disabled={props.readOnly} onToggle={(selector, checked) => toggleGuardrail("outputGuardrails", selector, checked)} />
            </TabsContent>

            <TabsContent value="metadata" className="flex flex-col gap-4 p-4">
              <div className="flex justify-end"><Button type="button" size="sm" variant="outline" disabled={props.readOnly} onClick={() => changeConfig({ metadata: [...config.metadata, createPlaygroundMetadataRow()] })}><Plus data-icon="inline-start" />{props.t("actions.addMetadata")}</Button></div>
              {config.metadata.length === 0 ? <EmptyState variant="inline" density="compact" title={props.t("settings.metadata")} /> : null}
              {config.metadata.map((row) => (
                <div key={row.id} className="grid gap-2 sm:grid-cols-[1fr_1fr_auto]">
                  <Input value={row.key} dir="ltr" aria-label={props.t("settings.metadataKey")} placeholder={props.t("settings.metadataKey")} disabled={props.readOnly} onChange={(event) => changeConfig({ metadata: config.metadata.map((entry) => entry.id === row.id ? { ...entry, key: event.target.value } : entry) })} />
                  <Input value={row.value} aria-label={props.t("settings.metadataValue")} placeholder={props.t("settings.metadataValue")} disabled={props.readOnly} onChange={(event) => changeConfig({ metadata: config.metadata.map((entry) => entry.id === row.id ? { ...entry, value: event.target.value } : entry) })} />
                  <Button type="button" size="icon" variant="ghost" aria-label={props.t("actions.remove")} disabled={props.readOnly} onClick={() => changeConfig({ metadata: config.metadata.filter((entry) => entry.id !== row.id) })}><Trash2 data-icon="inline-start" /></Button>
                </div>
              ))}
            </TabsContent>
          </ScrollArea>
        </Tabs>}
        <SheetFooter className="border-t border-border p-4"><Button type="button" variant="outline" onClick={() => props.onOpenChange(false)}>{props.t("actions.close")}</Button></SheetFooter>
      </SheetContent>
    </Sheet>
  );
}

function ImageGenerationSettings(props: {
  column: PlaygroundColumnDraft;
  readOnly: boolean;
  onChange: (column: PlaygroundColumnDraft) => void;
  t: Translation;
}) {
  const config = props.column.imageConfig;
  function changeConfig(patch: Partial<typeof config>) {
    props.onChange({ ...props.column, imageConfig: { ...config, ...patch }, dirty: true });
  }

  return (
    <ScrollArea className="min-h-0 flex-1">
      <FieldGroup className="p-4">
        <Alert variant="info"><AlertDescription>{props.t("image.settingsHelp")}</AlertDescription></Alert>
        <div className="grid gap-4 sm:grid-cols-2">
          <AppNumberField label={props.t("image.count")} description={props.t("image.countHelp")} value="1" min={1} max={1} disabled />
          <AppSelectField
            label={props.t("image.size")}
            value={config.size}
            disabled={props.readOnly}
            onValueChange={(size) => changeConfig({ size })}
            options={["auto", "1024x1024", "1536x1024", "1024x1536", "1792x1024", "1024x1792"].map((value) => ({ value, label: value }))}
            optionTextDir="ltr"
          />
          <AppSelectField
            label={props.t("image.quality")}
            value={config.quality}
            disabled={props.readOnly}
            onValueChange={(quality) => changeConfig({ quality })}
            options={["auto", "low", "medium", "high", "standard", "hd"].map((value) => ({ value, label: value }))}
            optionTextDir="ltr"
          />
          <AppSelectField
            label={props.t("image.responseFormat")}
            description={props.t("image.responseFormatHelp")}
            value={config.responseFormat || "default"}
            disabled={props.readOnly}
            onValueChange={(value) => changeConfig({ responseFormat: value === "default" ? "" : value as "url" | "b64_json" })}
            options={[{ value: "default", label: props.t("image.modelDefault") }, { value: "url", label: "URL" }, { value: "b64_json", label: "Base64 JSON" }]}
            optionTextDir="ltr"
          />
          <AppSelectField
            label={props.t("image.outputFormat")}
            value={config.outputFormat || "default"}
            disabled={props.readOnly}
            onValueChange={(value) => changeConfig({ outputFormat: value === "default" ? "" : value as "png" | "jpeg" | "webp" })}
            options={[{ value: "default", label: props.t("image.modelDefault") }, { value: "png", label: "PNG" }, { value: "jpeg", label: "JPEG" }, { value: "webp", label: "WebP" }]}
            optionTextDir="ltr"
          />
          {(config.outputFormat === "jpeg" || config.outputFormat === "webp") ? (
            <AppNumberField label={props.t("image.outputCompression")} value={config.outputCompression} min={0} max={100} step={1} disabled={props.readOnly} onChange={(event) => changeConfig({ outputCompression: event.target.value })} />
          ) : null}
          <AppSelectField
            label={props.t("image.background")}
            value={config.background}
            disabled={props.readOnly}
            onValueChange={(value) => changeConfig({ background: value as "auto" | "transparent" | "opaque" })}
            options={["auto", "transparent", "opaque"].map((value) => ({ value, label: value }))}
            optionTextDir="ltr"
          />
          <AppSelectField
            label={props.t("image.moderation")}
            value={config.moderation}
            disabled={props.readOnly}
            onValueChange={(value) => changeConfig({ moderation: value as "auto" | "low" })}
            options={["auto", "low"].map((value) => ({ value, label: value }))}
            optionTextDir="ltr"
          />
        </div>
      </FieldGroup>
    </ScrollArea>
  );
}

type ResolvedCacheLimits = {
  resolved: boolean;
  simpleEnabled: boolean;
  semanticEnabled: boolean;
  simpleTTLDays: number | null;
};

function resolveCacheLimits(entitlements: PlaygroundCacheEntitlements): ResolvedCacheLimits {
  const features = entitlements?.features;
  if (!features) {
    return { resolved: false, simpleEnabled: true, semanticEnabled: true, simpleTTLDays: null };
  }

  return {
    resolved: true,
    simpleEnabled: readBooleanEntitlement(features["cache.simple.enabled"], true),
    semanticEnabled: readBooleanEntitlement(features["cache.semantic.enabled"], false),
    simpleTTLDays: readPositiveNumberEntitlement(features["cache.simple.ttl_days"]),
  };
}

function getCachePlanLimitMessage(limits: ResolvedCacheLimits, t: Translation): string {
  if (!limits.simpleEnabled && !limits.semanticEnabled) return t("settings.cachePlanUnavailable");
  if (limits.simpleEnabled && !limits.semanticEnabled && limits.simpleTTLDays !== null) {
    return t("settings.cachePlanSimpleOnlyWithTTL", { days: limits.simpleTTLDays });
  }
  if (limits.simpleEnabled && !limits.semanticEnabled) return t("settings.cachePlanSimpleOnly");
  if (!limits.simpleEnabled && limits.semanticEnabled) return t("settings.cachePlanSemanticOnly");
  if (limits.simpleTTLDays !== null) return t("settings.cachePlanBothWithTTL", { days: limits.simpleTTLDays });
  return t("settings.cachePlanBoth");
}

function readBooleanEntitlement(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}

function readPositiveNumberEntitlement(value: unknown): number | null {
  if (value === null || value === undefined || value === "") return null;
  const parsed = typeof value === "number" ? value : Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function GuardrailChecklist(props: { title: string; emptyLabel: string; guardrails: GuardrailItem[]; selected: string[]; disabled: boolean; onToggle: (selector: string, checked: boolean) => void }) {
  const idPrefix = React.useId();
  return (
    <FieldSet>
      <FieldLegend variant="label">{props.title}</FieldLegend>
      <FieldGroup className="gap-1 rounded-lg border border-border p-2">
        {props.guardrails.map((guardrail) => {
          const selector = guardrail.selector || `guardrails/${guardrail.slug}`;
          const controlID = `${idPrefix}-${guardrail.id}`;
          return (
            <Field key={guardrail.id} orientation="horizontal" data-disabled={props.disabled || undefined} className="rounded-md px-2 py-1 hover:bg-muted/50">
              <Checkbox id={controlID} checked={props.selected.includes(selector)} disabled={props.disabled} onCheckedChange={(checked) => props.onToggle(selector, checked === true)} />
              <FieldContent><FieldLabel htmlFor={controlID}>{guardrail.name}</FieldLabel>{guardrail.description ? <FieldDescription>{guardrail.description}</FieldDescription> : null}</FieldContent>
            </Field>
          );
        })}
        {props.guardrails.length === 0 ? <FieldDescription className="px-2 py-1">{props.emptyLabel}</FieldDescription> : null}
      </FieldGroup>
    </FieldSet>
  );
}
