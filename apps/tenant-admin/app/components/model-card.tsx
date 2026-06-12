"use client";

import { type ReactNode, useState } from "react";
import Link from "next/link";
import {
  AlertTriangle,
  Brain,
  CheckCircle2,
  Clock,
  Code2,
  Copy,
  Database,
  DollarSign,
  ExternalLink,
  Eye,
  FileJson,
  Gauge,
  ImageIcon,
  Mic,
  ShieldCheck,
  Sparkles,
  Text,
  Wrench,
  Zap,
} from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
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
  tradeoffs?: string[];
  metadata?: {
    modelId: string;
    mode?: "chat" | "completion" | "embedding" | "image" | "audio";
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
  pricingModeBadge?: string;
  pricingAction?: ReactNode;
  labels?: {
    context?: string;
    maxOutput?: string;
    input?: string;
    output?: string;
    latency?: string;
    quality?: string;
    pricing?: string;
    pricingUnavailable?: string;
    technicalMetadata?: string;
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
    pricingUnavailable: "Pricing unavailable",
    technicalMetadata: "Technical metadata",
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

  const primaryCategory = model.category[0] ?? "chat";
  const isDeprecated = model.status === "deprecated";
  const isFast = model.latencyTier === "low";
  const isCheap = (model.pricing?.inputPerMillion ?? Number.POSITIVE_INFINITY) <= 0.25;

  const metrics = [
    { label: labels.context, value: formatTokenCount(model.contextWindow), hint: "Maximum prompt context window." },
    { label: labels.maxOutput, value: formatTokenCount(model.maxOutputTokens), hint: "Maximum output token generation." },
    {
      label: labels.input,
      value: formatPriceWithDisplay(model.pricing?.inputPerMillion, pricingDisplay),
      hint: "Cost by selected currency/scale.",
    },
    {
      label: labels.output,
      value: formatPriceWithDisplay(model.pricing?.outputPerMillion, pricingDisplay),
      hint: "Cost by selected currency/scale.",
    },
    { label: labels.latency, value: model.latencyTier ? capitalize(model.latencyTier) : "N/A", hint: "Relative response speed." },
    { label: labels.quality, value: model.qualityTier ? capitalize(model.qualityTier) : "N/A", hint: "Relative output quality." },
  ];

  async function handleCopyModelId() {
    const id = model.metadata?.modelId || model.id;
    try {
      await navigator.clipboard.writeText(id);
      setCopied(true);
      setTimeout(() => setCopied(false), 1400);
    } catch {
      setCopied(false);
    }
  }

  return (
    <Card
      className={cn(
        "group h-full rounded-xl border border-border/70 bg-card transition-colors hover:border-border",
        model.recommended && "ring-1 ring-primary/25",
        isDeprecated && "border-amber-500/50 bg-amber-500/5",
        props.className,
      )}
    >
      <CardHeader className="space-y-3 pb-3">
        <div className="flex items-start gap-3">
          <Avatar className="h-9 w-9 border">
            <AvatarImage src={model.provider.logoUrl} alt={model.provider.name} />
            <AvatarFallback>{initials(model.provider.name)}</AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <CardTitle className="truncate text-base font-semibold" title={model.name}>
              {model.name}
            </CardTitle>
            <p className="truncate text-xs text-muted-foreground" title={model.provider.name}>
              {model.provider.name}
            </p>
          </div>
          <div className="flex flex-wrap justify-end gap-1">
            <Badge variant="secondary">{labelCategory(primaryCategory, labels)}</Badge>
            <Badge variant={statusVariant(model.status)}>{labelStatus(model.status, labels)}</Badge>
            {model.recommended ? (
              <Badge className="bg-primary/15 text-primary hover:bg-primary/20">
                <Sparkles className="mr-1 h-3 w-3" />
                {labels.recommended}
              </Badge>
            ) : null}
            {model.popular ? <Badge variant="outline">{labels.popular}</Badge> : null}
            {typeof props.isSelected === "boolean" && props.onSelectedChange ? (
              <label className="inline-flex items-center gap-1 rounded-md border border-border/60 px-2 py-1 text-xs">
                <input
                  type="checkbox"
                  checked={props.isSelected}
                  onChange={(event) => props.onSelectedChange?.(event.target.checked)}
                  className="h-3.5 w-3.5"
                />
                {props.selectedCheckboxLabel ?? "Enabled"}
              </label>
            ) : null}
            {typeof props.isCompared === "boolean" && props.onComparedChange ? (
              <label className="inline-flex items-center gap-1 rounded-md border border-border/60 px-2 py-1 text-xs">
                <input
                  type="checkbox"
                  checked={props.isCompared}
                  onChange={(event) => props.onComparedChange?.(event.target.checked)}
                  className="h-3.5 w-3.5"
                />
                {props.compareCheckboxLabel ?? "مقایسه"}
              </label>
            ) : null}
          </div>
        </div>
        <p className="line-clamp-2 text-sm leading-5 text-foreground/90">{model.description}</p>
      </CardHeader>

      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
          {metrics.map((item) => (
            <div key={item.label} className="rounded-md border border-border/60 bg-muted/30 px-2.5 py-2" title={item.hint}>
              <p className="truncate text-[11px] text-muted-foreground">{item.label}</p>
              <p className="truncate text-sm font-medium">{item.value}</p>
            </div>
          ))}
        </div>

        <div className="flex flex-wrap gap-1.5">
          {model.capabilities.map((capability) => (
            <Badge key={capability} variant="outline" className="gap-1">
              {iconForCapability(capability)}
              <span className="text-[11px]">{capability}</span>
            </Badge>
          ))}
        </div>

        {model.tradeoffs && model.tradeoffs.length > 0 ? (
          <section className="rounded-md border border-amber-500/40 bg-amber-500/10 p-2.5">
            <p className="mb-1 flex items-center gap-1 text-xs font-medium text-amber-700 dark:text-amber-300">
              <AlertTriangle className="h-3.5 w-3.5" />
              {labels.tradeOffs}
            </p>
            <ul className="space-y-0.5 text-xs text-amber-800 dark:text-amber-200">
              {model.tradeoffs.slice(0, 3).map((item) => (
                <li key={item}>• {item}</li>
              ))}
            </ul>
          </section>
        ) : null}

        <section className="space-y-1.5">
          <div className="flex items-center justify-between gap-2">
            <p className="text-xs font-medium text-muted-foreground">{labels.pricing}</p>
            <div className="flex items-center gap-2">
              {props.pricingModeBadge ? <Badge variant="outline">{props.pricingModeBadge}</Badge> : null}
              {props.pricingAction}
            </div>
          </div>
          <div className="space-y-0.5 text-xs">
            <p>{priceRow(labels.input, model.pricing?.inputPerMillion, pricingDisplay)}</p>
            <p>{priceRow(labels.output, model.pricing?.outputPerMillion, pricingDisplay)}</p>
            {model.pricing?.cachedInputPerMillion !== undefined ? <p>{priceRow(labels.cachedInput, model.pricing.cachedInputPerMillion, pricingDisplay)}</p> : null}
            {model.pricing?.imagePerMillion !== undefined ? <p>{priceRow(labels.image, model.pricing.imagePerMillion, pricingDisplay)}</p> : null}
            {model.pricing?.audioPerMillion !== undefined ? <p>{priceRow(labels.audio, model.pricing.audioPerMillion, pricingDisplay)}</p> : null}
            {!hasAnyPricing(model.pricing) ? <p className="text-muted-foreground">{labels.pricingUnavailable}</p> : null}
          </div>
        </section>

        <details className="rounded-md border border-border/60 bg-muted/20 p-2.5">
          <summary className="cursor-pointer text-xs font-medium text-muted-foreground">{labels.technicalMetadata}</summary>
          <Separator className="my-2" />
          <div className="space-y-1 text-xs">
            <p><span className="text-muted-foreground">{labels.modelId}:</span> {model.metadata?.modelId ?? model.id}</p>
            <p><span className="text-muted-foreground">{labels.providerId}:</span> {model.provider.id}</p>
            <p><span className="text-muted-foreground">{labels.apiMode}:</span> {model.metadata?.mode ?? "N/A"}</p>
            <p><span className="text-muted-foreground">{labels.releaseDate}:</span> {model.metadata?.releaseDate ?? "N/A"}</p>
            <p><span className="text-muted-foreground">{labels.knowledgeCutoff}:</span> {model.metadata?.knowledgeCutoff ?? "N/A"}</p>
            <p><span className="text-muted-foreground">{labels.regions}:</span> {(model.metadata?.regions ?? []).join(", ") || "N/A"}</p>
            <p><span className="text-muted-foreground">{labels.rateLimitNotes}:</span> {model.metadata?.rateLimitNotes ?? "N/A"}</p>
            <p><span className="text-muted-foreground">{labels.license}:</span> {model.metadata?.license ?? "N/A"}</p>
          </div>
        </details>
      </CardContent>

      <CardFooter className="grid grid-cols-2 gap-2 pt-2 sm:grid-cols-3">
        {props.detailsHref ? (
          <Button asChild variant="outline" size="sm" className="gap-1">
            <Link href={props.detailsHref} target="_blank" rel="noreferrer">
              <ExternalLink className="h-3.5 w-3.5" />
              {labels.viewDetails}
            </Link>
          </Button>
        ) : (
          <Button variant="outline" size="sm" className="gap-1" onClick={() => props.onViewDetails?.(model)}>
            <ExternalLink className="h-3.5 w-3.5" />
            {labels.viewDetails}
          </Button>
        )}
        <Button variant="outline" size="sm" className="gap-1" onClick={handleCopyModelId}>
          {copied ? <CheckCircle2 className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
          {copied ? labels.copied : labels.copyId}
        </Button>
      </CardFooter>

      {(isFast || isCheap) && !isDeprecated ? (
        <div className="px-6 pb-4">
          <div className="flex flex-wrap gap-1.5">
            {isFast ? (
              <Badge variant="secondary" className="gap-1 text-[11px]">
                <Zap className="h-3 w-3" />
                Faster tier
              </Badge>
            ) : null}
            {isCheap ? (
              <Badge variant="secondary" className="gap-1 text-[11px]">
                <DollarSign className="h-3 w-3" />
                {labels.lowerCost}
              </Badge>
            ) : null}
          </div>
        </div>
      ) : null}
    </Card>
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

function labelStatus(value: ModelCardData["status"], labels: NonNullable<ModelCardProps["labels"]>) {
  switch (value) {
    case "stable":
      return labels.statusStable;
    case "preview":
      return labels.statusPreview;
    case "deprecated":
      return labels.statusDeprecated;
    case "new":
      return labels.statusNew;
    default:
      return value;
  }
}

function statusVariant(value: ModelCardData["status"]): "default" | "secondary" | "destructive" | "outline" {
  if (value === "deprecated") return "destructive";
  if (value === "preview") return "secondary";
  if (value === "new") return "default";
  return "outline";
}

function capitalize(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1);
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

function priceRow(
  label: string,
  value: number | undefined,
  display: { currency: "USD" | "TOMAN"; scale: "PER_MILLION_TOKENS" | "PER_TOKEN"; usdToIrr: number },
) {
  return (
    <>
      <span className="text-muted-foreground">{label}:</span>{" "}
      {formatPriceWithDisplay(value, display)}
    </>
  );
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

function hasAnyPricing(pricing: ModelCardData["pricing"]) {
  if (!pricing) return false;
  return [
    pricing.inputPerMillion,
    pricing.outputPerMillion,
    pricing.cachedInputPerMillion,
    pricing.imagePerMillion,
    pricing.audioPerMillion,
  ].some((item) => typeof item === "number" && Number.isFinite(item));
}

function trimZeros(value: number) {
  return value.toFixed(6).replace(/\.?0+$/, "");
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
