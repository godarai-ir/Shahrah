"use client";

import { type ReactNode, useState } from "react";
import Link from "next/link";
import {
  AlertTriangle,
  Brain,
  Braces,
  CheckCircle2,
  Clock,
  Code2,
  Copy,
  Database,
  ExternalLink,
  Eye,
  FileJson,
  FileText,
  Gauge,
  ImageIcon,
  MessageSquare,
  Mic,
  ShieldCheck,
  SlidersHorizontal,
  Sparkles,
  Text,
  Video,
  Volume2,
  Wrench,
  Zap,
} from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";

export interface ModelCardData {
  id: string;
  name: string;
  provider: {
    id: string;
    name: string;
    logoUrl?: string;
  };
  description: string;
  category: Array<"chat" | "reasoning" | "embedding" | "vision" | "audio" | "code" | "multimodal">;
  status: "stable" | "preview" | "deprecated" | "new";
  recommended?: boolean;
  popular?: boolean;
  contextWindow?: number;
  maxOutputTokens?: number;
  pricing?: {
    inputPerMillion?: number;
    outputPerMillion?: number;
    cachedInputPerMillion?: number;
    imagePerMillion?: number;
    audioPerMillion?: number;
  };
  latencyTier?: "low" | "medium" | "high";
  qualityTier?: "basic" | "good" | "high" | "frontier";
  capabilities: string[];
  inputModalities?: string[];
  outputModalities?: string[];
  invocationModes?: string[];
  tradeoffs?: string[];
  metadata?: {
    modelId: string;
    fqn?: string;
    mode?: string;
    releaseDate?: string;
    knowledgeCutoff?: string;
    regions?: string[];
    rateLimitNotes?: string;
    license?: string;
  };
}

type ModelCardProps = {
  model: ModelCardData;
  detailsHref?: string;
  pricingDisplay?: {
    currency: "USD" | "TOMAN";
    scale: "PER_MILLION_TOKENS" | "PER_TOKEN";
    usdToIrr: number;
  };
  onViewDetails?: (model: ModelCardData) => void;
  onCompare?: (model: ModelCardData) => void;
  isSelected?: boolean;
  onSelectedChange?: (checked: boolean) => void;
  selectedCheckboxLabel?: string;
  isCompared?: boolean;
  onComparedChange?: (checked: boolean) => void;
  compareCheckboxLabel?: string;
  stateBadge?: {
    label: string;
    tone: "active" | "inactive";
  };
  pricingModeBadge?: string;
  pricingAction?: ReactNode;
  hasCustomPricing?: boolean;
  footerActions?: ReactNode;
  extraContent?: ReactNode;
  labels?: {
    context?: string;
    maxOutput?: string;
    input?: string;
    output?: string;
    latency?: string;
    quality?: string;
    pricing?: string;
    priceInputOutput?: string;
    pricingUnavailable?: string;
    technicalMetadata?: string;
    inputFormats?: string;
    outputFormats?: string;
    capabilities?: string;
    invocationModes?: string;
    fqn?: string;
    customPricing?: string;
    pricingPerMillion?: string;
    modelId?: string;
    providerId?: string;
    apiMode?: string;
    releaseDate?: string;
    knowledgeCutoff?: string;
    regions?: string;
    rateLimitNotes?: string;
    license?: string;
    viewDetails?: string;
    copyId?: string;
    copied?: string;
    lowerCost?: string;
    recommended?: string;
    popular?: string;
    tradeOffs?: string;
    cachedInput?: string;
    image?: string;
    audio?: string;
    statusStable?: string;
    statusPreview?: string;
    statusDeprecated?: string;
    statusNew?: string;
    categoryChat?: string;
    categoryReasoning?: string;
    categoryEmbedding?: string;
    categoryVision?: string;
    categoryAudio?: string;
    categoryCode?: string;
    categoryMultimodal?: string;
  };
  className?: string;
};

export function ModelCard(props: ModelCardProps) {
  const { model } = props;
  const pricingDisplay = props.pricingDisplay ?? {
    currency: "USD" as const,
    scale: "PER_MILLION_TOKENS" as const,
    usdToIrr: 0,
  };
  const [copied, setCopied] = useState(false);
  const labels = {
    context: "Context",
    maxOutput: "Max Output",
    input: "Input",
    output: "Output",
    latency: "Latency",
    quality: "Quality",
    pricing: "Pricing",
    priceInputOutput: "Input cost / Output cost",
    pricingUnavailable: "Pricing unavailable",
    technicalMetadata: "Technical metadata",
    inputFormats: "Input formats",
    outputFormats: "Output formats",
    capabilities: "Capabilities",
    invocationModes: "Invocation modes",
    fqn: "Model FQN",
    customPricing: "Custom pricing",
    pricingPerMillion: "Cost (1M tokens)",
    modelId: "Model ID",
    providerId: "Provider ID",
    apiMode: "API mode",
    releaseDate: "Release date",
    knowledgeCutoff: "Knowledge cutoff",
    regions: "Regions",
    rateLimitNotes: "Rate limit notes",
    license: "License",
    viewDetails: "View Details",
    copyId: "Copy ID",
    copied: "Copied",
    lowerCost: "Lower cost",
    recommended: "Recommended",
    popular: "Popular",
    tradeOffs: "Trade-offs",
    cachedInput: "Cached input",
    image: "Image",
    audio: "Audio",
    statusStable: "Stable",
    statusPreview: "Preview",
    statusDeprecated: "Deprecated",
    statusNew: "New",
    categoryChat: "Chat",
    categoryReasoning: "Reasoning",
    categoryEmbedding: "Embedding",
    categoryVision: "Vision",
    categoryAudio: "Audio",
    categoryCode: "Code",
    categoryMultimodal: "Multimodal",
    ...props.labels,
  };

  const isDeprecated = model.status === "deprecated";
  const isFast = model.latencyTier === "low";
  const modelFqn = model.metadata?.fqn || model.metadata?.modelId || model.id;
  const invocationModes = model.invocationModes?.length
    ? model.invocationModes
    : model.metadata?.mode?.split(",").map((item) => item.trim()).filter(Boolean) ?? [];

  const perMillionPricingDisplay = {
    ...pricingDisplay,
    scale: "PER_MILLION_TOKENS" as const,
  };

  async function handleCopyModelId() {
    try {
      await navigator.clipboard.writeText(modelFqn);
      setCopied(true);
      setTimeout(() => setCopied(false), 1400);
    } catch {
      setCopied(false);
    }
  }

  return (
    <Card
      dir="rtl"
      className={cn(
        "group w-full min-w-0 max-w-full overflow-hidden rounded-lg border border-border/70 bg-card transition-colors hover:border-border",
        model.recommended && "ring-1 ring-primary/25",
        isDeprecated && "border-[var(--warning-soft-border)] bg-[var(--warning-soft)]",
        props.stateBadge?.tone === "active" && "border-[var(--success-soft-border)]",
        props.stateBadge?.tone === "inactive" && "border-border/70 bg-card",
        props.className,
      )}
    >
      <CardHeader className="p-2.5 pb-1.5">
        <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 items-start gap-2">
            <Avatar className="size-7 shrink-0 border bg-background">
              <AvatarImage src={model.provider.logoUrl} alt={model.provider.name} />
              <AvatarFallback>{initials(model.provider.name)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <div className="flex min-w-0 items-center gap-1">
                <CardTitle className="min-w-0 truncate text-[15px] font-semibold" title={model.name} dir="ltr">
                  {model.name}
                </CardTitle>
                {props.detailsHref ? (
                  <Button asChild variant="ghost" size="icon" className="size-6 shrink-0" aria-label={labels.viewDetails} title={labels.viewDetails}>
                    <Link href={props.detailsHref} target="_blank" rel="noreferrer">
                      <ExternalLink className="h-3 w-3" />
                    </Link>
                  </Button>
                ) : (
                  <Button variant="ghost" size="icon" className="size-6 shrink-0" onClick={() => props.onViewDetails?.(model)} aria-label={labels.viewDetails} title={labels.viewDetails}>
                    <ExternalLink className="h-3 w-3" />
                  </Button>
                )}
              </div>
            </div>
          </div>
          <div className="flex max-w-full shrink-0 flex-wrap items-center justify-start gap-2 sm:justify-end">
            {props.pricingModeBadge ? <Badge variant="secondary">{props.pricingModeBadge}</Badge> : null}
            {typeof props.isSelected === "boolean" && props.onSelectedChange ? (
              <label className="inline-flex min-h-7 cursor-pointer items-center gap-1.5 px-1 text-[13px] font-medium">
                <span>{props.selectedCheckboxLabel ?? "Enabled"}</span>
                <Switch
                  checked={props.isSelected}
                  onCheckedChange={(checked) => props.onSelectedChange?.(checked)}
                  aria-label={props.selectedCheckboxLabel ?? "Enabled"}
                />
              </label>
            ) : null}
            {typeof props.isCompared === "boolean" && props.onComparedChange ? (
              <label className="inline-flex min-h-8 cursor-pointer items-center gap-2 rounded-md border border-border/70 bg-background px-2.5 text-sm">
                <input
                  type="checkbox"
                  checked={props.isCompared}
                  onChange={(event) => props.onComparedChange?.(event.target.checked)}
                  className="h-3.5 w-3.5"
                />
                {props.compareCheckboxLabel ?? "Compare"}
              </label>
            ) : null}
          </div>
        </div>
        <div className="flex flex-wrap items-center justify-between gap-2 pt-0.5">
          <div className="flex min-w-0 flex-wrap gap-1">
            {model.category.map((category) => (
              <Badge key={category} variant="secondary" className="px-2 py-0 text-[11px]">{labelCategory(category, labels)}</Badge>
            ))}
            {props.hasCustomPricing ? (
              <Badge className="gap-1 border-[var(--warning-soft-border)] bg-[var(--warning-soft)] px-2 py-0 text-[11px] text-[var(--warning-soft-foreground)] hover:bg-[color-mix(in_oklab,var(--warning-soft)_88%,black)]" variant="outline">
                <SlidersHorizontal className="h-3 w-3" />
                {labels.customPricing}
              </Badge>
            ) : null}
            {model.recommended ? (
              <Badge className="surface-primary-soft gap-1 px-2 py-0 text-[11px] hover:bg-[var(--primary-soft-hover)]">
                <Sparkles className="h-3 w-3" />
                {labels.recommended}
              </Badge>
            ) : null}
            {model.popular ? <Badge variant="outline" className="px-2 py-0 text-[11px]">{labels.popular}</Badge> : null}
          </div>
          <div className="flex min-w-0 flex-wrap items-center gap-1.5">
            <Button variant="outline" size="sm" className="h-8 gap-1 px-2 text-[12px]" onClick={handleCopyModelId}>
              {copied ? <CheckCircle2 className="h-3 w-3 text-[var(--success-soft-foreground)]" /> : <Copy className="h-3 w-3" />}
              {copied ? labels.copied : labels.copyId}
            </Button>
            {props.footerActions}
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex flex-col gap-1.5 px-2.5 py-1.5">
        <div className="grid grid-cols-1 gap-1.5 sm:grid-cols-2 xl:grid-cols-4">
          <MetricCell label={labels.context} value={formatTokenCount(model.contextWindow)} />
          <PriceMetricCell
            label={labels.priceInputOutput}
            inputValue={formatPriceValueOnly(model.pricing?.inputPerMillion, perMillionPricingDisplay)}
            outputValue={formatPriceValueOnly(model.pricing?.outputPerMillion, perMillionPricingDisplay)}
            inputLabel={labels.input}
            outputLabel={labels.output}
            action={props.pricingAction}
          />
          <IconListCell label={labels.inputFormats} values={model.inputModalities} emptyLabel="N/A" />
          <IconListCell label={labels.outputFormats} values={model.outputModalities} emptyLabel="N/A" />
        </div>

        <section className="flex flex-col gap-0.5">
          <p className="text-xs font-medium text-muted-foreground">{labels.capabilities}</p>
          <div className="flex flex-wrap gap-1">
            {model.capabilities.length > 0 ? (
              model.capabilities.map((capability) => (
                <Badge key={capability} variant="outline" className="gap-1 px-2 py-0">
                  {iconForCapability(capability)}
                  <span className="text-[11px]">{capability}</span>
                </Badge>
              ))
            ) : (
              <span className="text-xs text-muted-foreground">N/A</span>
            )}
          </div>
        </section>

        <section className="flex flex-col gap-0.5">
          <p className="text-xs font-medium text-muted-foreground">{labels.invocationModes}</p>
          <div className="flex flex-wrap gap-1">
            {invocationModes.length > 0 ? (
              invocationModes.map((mode) => (
                <Badge key={mode} variant="secondary" className="gap-1 px-2 py-0 text-[11px]">
                  {iconForMode(mode)}
                  {mode}
                </Badge>
              ))
            ) : (
              <span className="text-xs text-muted-foreground">N/A</span>
            )}
          </div>
        </section>

        {model.tradeoffs && model.tradeoffs.length > 0 ? (
          <section className="rounded-md border border-[var(--warning-soft-border)] bg-[var(--warning-soft)] p-2">
            <p className="mb-1 flex items-center gap-1 text-xs font-medium text-[var(--warning-soft-foreground)]">
              <AlertTriangle className="h-3.5 w-3.5" />
              {labels.tradeOffs}
            </p>
            <ul className="flex flex-col gap-0.5 text-xs text-[color-mix(in_oklab,var(--warning-soft-foreground)_85%,black)] dark:text-[color-mix(in_oklab,var(--warning-soft-foreground)_88%,white)]">
              {model.tradeoffs.slice(0, 3).map((item) => (
                <li key={item}>- {item}</li>
              ))}
            </ul>
          </section>
        ) : null}
      </CardContent>

      {isFast && !isDeprecated ? (
        <div className="px-3 pb-3">
          <div className="flex flex-wrap gap-1">
            <Badge variant="secondary" className="gap-1 text-[11px]">
              <Zap className="h-3 w-3" />
              Faster tier
            </Badge>
          </div>
        </div>
      ) : null}
      {props.extraContent ? <div className="px-2.5 pb-2.5">{props.extraContent}</div> : null}
    </Card>
  );
}

function IconListCell(props: { label: string; values?: string[]; emptyLabel: string }) {
  return (
    <div className="min-w-0 rounded-md border border-border/60 bg-muted/20 px-2 py-1.5">
      <IconList label={props.label} values={props.values} emptyLabel={props.emptyLabel} />
    </div>
  );
}

function PriceMetricCell(props: {
  label: string;
  inputLabel: string;
  outputLabel: string;
  inputValue: string;
  outputValue: string;
  action?: ReactNode;
}) {
  return (
    <div className="min-w-0 rounded-md border border-border/60 bg-muted/20 px-2 py-1.5">
      <div className="flex min-w-0 items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p className="truncate text-[11px] text-muted-foreground">{props.label}</p>
          <div className="mt-1 grid min-w-0 gap-x-3 gap-y-1 sm:grid-cols-2">
            <div className="min-w-0">
              <p className="text-[10px] text-muted-foreground">{props.inputLabel}</p>
              <p className="truncate text-[13px] font-semibold" dir="ltr">{props.inputValue}</p>
            </div>
            <div className="min-w-0">
              <p className="text-[10px] text-muted-foreground">{props.outputLabel}</p>
              <p className="truncate text-[13px] font-semibold" dir="ltr">{props.outputValue}</p>
            </div>
          </div>
        </div>
        {props.action ? <div className="shrink-0">{props.action}</div> : null}
      </div>
    </div>
  );
}

function IconList(props: { label: string; values?: string[]; emptyLabel: string }) {
  return (
    <div className="flex min-w-0 flex-col gap-1">
      <p className="truncate text-[11px] font-medium text-muted-foreground">{props.label}</p>
      <div className="flex flex-wrap gap-1">
        {props.values?.length ? (
          props.values.map((value) => (
            <Badge key={`${props.label}-${value}`} variant="outline" className="gap-1 px-2 py-0 text-[11px]">
              {iconForModality(value)}
              <span>{value}</span>
            </Badge>
          ))
        ) : (
          <span className="text-xs text-muted-foreground">{props.emptyLabel}</span>
        )}
      </div>
    </div>
  );
}

function MetricCell(props: { label: string; value: string }) {
  return (
    <div className="min-w-0 rounded-md border border-border/60 bg-muted/20 px-2 py-1.5">
      <p className="truncate text-[11px] text-muted-foreground">{props.label}</p>
      <p className="truncate text-[13px] font-semibold" dir="ltr">{props.value}</p>
    </div>
  );
}

function labelCategory(value: ModelCardData["category"][number], labels: NonNullable<ModelCardProps["labels"]>) {
  switch (value) {
    case "chat":
      return labels.categoryChat;
    case "reasoning":
      return labels.categoryReasoning;
    case "embedding":
      return labels.categoryEmbedding;
    case "vision":
      return labels.categoryVision;
    case "audio":
      return labels.categoryAudio;
    case "code":
      return labels.categoryCode;
    case "multimodal":
      return labels.categoryMultimodal;
    default:
      return value;
  }
}

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((item) => item[0]?.toUpperCase() ?? "")
    .join("");
}

export function formatTokenCount(value?: number) {
  if (!value || !Number.isFinite(value)) return "N/A";
  if (value >= 1_000_000) return `${trimZeros(value / 1_000_000)}M`;
  if (value >= 1_000) return `${trimZeros(value / 1_000)}K`;
  return `${value}`;
}

export function formatPriceWithDisplay(
  usdPerMillion: number | undefined,
  display: { currency: "USD" | "TOMAN"; scale: "PER_MILLION_TOKENS" | "PER_TOKEN"; usdToIrr: number },
) {
  if (typeof usdPerMillion !== "number" || !Number.isFinite(usdPerMillion)) return "N/A";

  let value = usdPerMillion;
  if (display.scale === "PER_TOKEN") value = value / 1_000_000;

  if (display.currency === "TOMAN") {
    const toman = (value * display.usdToIrr) / 10;
    const formatted = new Intl.NumberFormat("en-US", {
      maximumFractionDigits: display.scale === "PER_TOKEN" ? 2 : 0,
    }).format(toman);
    return display.scale === "PER_TOKEN" ? `${formatted} toman / token` : `${formatted} toman / 1M tokens`;
  }

  return display.scale === "PER_TOKEN"
    ? `$${trimZeros(value)} / token`
    : `$${trimZeros(value)} / 1M tokens`;
}

function formatPriceValueOnly(
  usdPerMillion: number | undefined,
  display: { currency: "USD" | "TOMAN"; scale: "PER_MILLION_TOKENS" | "PER_TOKEN"; usdToIrr: number },
) {
  return formatPriceWithDisplay(usdPerMillion, display);
}

function trimZeros(value: number) {
  return value.toFixed(6).replace(/\.?0+$/, "");
}

function iconForModality(value: string) {
  const key = value.toLowerCase();
  if (key.includes("image") || key.includes("vision")) return <ImageIcon className="h-3 w-3" />;
  if (key.includes("audio") || key.includes("speech")) return <Volume2 className="h-3 w-3" />;
  if (key.includes("video")) return <Video className="h-3 w-3" />;
  if (key.includes("json")) return <Braces className="h-3 w-3" />;
  if (key.includes("file") || key.includes("pdf")) return <FileText className="h-3 w-3" />;
  return <Text className="h-3 w-3" />;
}

function iconForMode(value: string) {
  const key = value.toLowerCase();
  if (key.includes("chat")) return <MessageSquare className="h-3 w-3" />;
  if (key.includes("embed")) return <Database className="h-3 w-3" />;
  if (key.includes("image")) return <ImageIcon className="h-3 w-3" />;
  if (key.includes("video")) return <Video className="h-3 w-3" />;
  if (key.includes("rank")) return <Gauge className="h-3 w-3" />;
  if (key.includes("code")) return <Code2 className="h-3 w-3" />;
  return <Sparkles className="h-3 w-3" />;
}

function iconForCapability(capability: string) {
  const key = capability.toLowerCase();
  if (key.includes("text")) return <Text className="h-3 w-3" />;
  if (key.includes("vision") || key.includes("image")) return <Eye className="h-3 w-3" />;
  if (key.includes("tool") || key.includes("function")) return <Wrench className="h-3 w-3" />;
  if (key.includes("json") || key.includes("structured")) return <FileJson className="h-3 w-3" />;
  if (key.includes("stream")) return <Gauge className="h-3 w-3" />;
  if (key.includes("reason")) return <Brain className="h-3 w-3" />;
  if (key.includes("code")) return <Code2 className="h-3 w-3" />;
  if (key.includes("embed")) return <Database className="h-3 w-3" />;
  if (key.includes("cache")) return <Clock className="h-3 w-3" />;
  if (key.includes("pdf")) return <ShieldCheck className="h-3 w-3" />;
  if (key.includes("audio")) return <Mic className="h-3 w-3" />;
  if (key.includes("multimodal")) return <ImageIcon className="h-3 w-3" />;
  return <Sparkles className="h-3 w-3" />;
}
