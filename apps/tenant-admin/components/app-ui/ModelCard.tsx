"use client";

import { type ReactNode, useState } from "react";
import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
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
} from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ButtonGroup } from "@/components/ui/button-group";
import { CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
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
  mobileDensity?: "compact" | "detailed";
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
  selectedControlLabel?: ReactNode;
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
  showCopyAction?: boolean;
  footerActions?: ReactNode;
  extraContent?: ReactNode;
  formatCapabilityLabel?: (value: string) => string;
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
    pricingPerToken?: string;
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
  const mobileDensity = props.mobileDensity ?? "detailed";
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
    pricingPerMillion: "Cost per 1M tokens (some models use tiered pricing)",
    pricingPerToken: "Cost per token",
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
  const isCompactMobile = mobileDensity === "compact";
  const modelFqn = model.metadata?.fqn || model.metadata?.modelId || model.id;
  const invocationModes = model.invocationModes?.length
    ? model.invocationModes
    : model.metadata?.mode?.split(",").map((item) => item.trim()).filter(Boolean) ?? [];

  const visibleCategories = isCompactMobile ? model.category.slice(0, 2) : model.category;
  const visibleCapabilities = isCompactMobile ? model.capabilities.slice(0, 3) : model.capabilities;
  const visibleModes = isCompactMobile ? invocationModes.slice(0, 2) : invocationModes;
  const remainingCapabilityCount = Math.max(0, model.capabilities.length - visibleCapabilities.length);
  const remainingModeCount = Math.max(0, invocationModes.length - visibleModes.length);
  const showCopyAction = props.showCopyAction ?? true;
  const showSelectedControl = typeof props.isSelected === "boolean" && Boolean(props.onSelectedChange);
  const hasUtilityActions = showCopyAction || Boolean(props.footerActions);

  async function handleCopyModelId() {
    try {
      await copyTextToClipboard(modelFqn);
      setCopied(true);
      setTimeout(() => setCopied(false), 1400);
    } catch {
      setCopied(false);
    }
  }

  function renderUtilityActions() {
    if (!hasUtilityActions) return null;

    return (
      <ButtonGroup aria-label={`${model.name} ${labels.copyId}`}>
        {showCopyAction ? (
          <Button type="button" variant="outline" size="sm" onClick={handleCopyModelId}>
            {copied ? <CheckCircle2 className="text-success" data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
            {copied ? labels.copied : labels.copyId}
          </Button>
        ) : null}
        {props.footerActions}
      </ButtonGroup>
    );
  }

  function renderSelectedControl(className?: string) {
    if (!showSelectedControl) return null;

    if (props.selectedControlLabel === null) {
      return (
        <Switch
          checked={props.isSelected === true}
          onCheckedChange={(checked) => props.onSelectedChange?.(checked)}
          aria-label={props.selectedCheckboxLabel ?? "Enabled"}
          className={cn(
            "border-muted-foreground/50 shadow-sm data-[state=checked]:border-primary data-[state=unchecked]:bg-muted",
            className,
          )}
        />
      );
    }

    return (
      <label
        className={cn(
          "inline-flex min-h-9 cursor-pointer items-center gap-2 rounded-md border px-2.5 text-sm font-medium transition-colors",
          props.isSelected
            ? "border-success-soft-border bg-success-soft text-success-soft-foreground"
            : "border-border bg-background text-foreground",
          className,
        )}
      >
        <span className="whitespace-nowrap">
          {props.selectedControlLabel ?? props.selectedCheckboxLabel ?? "Enabled"}
        </span>
        <Switch
          checked={props.isSelected === true}
          onCheckedChange={(checked) => props.onSelectedChange?.(checked)}
          aria-label={props.selectedCheckboxLabel ?? "Enabled"}
        />
      </label>
    );
  }

  function renderCategoryBadges() {
    return (
      <>
        {visibleCategories.map((category) => (
          <Badge key={category} variant={categoryBadgeVariant(category)}>
            {labelCategory(category, labels)}
          </Badge>
        ))}
        {props.pricingModeBadge ? <Badge variant="secondary">{props.pricingModeBadge}</Badge> : null}
        {model.recommended ? (
          <Badge className="gap-1" variant="success">
            <Sparkles className="size-3.5" aria-hidden="true" />
            {labels.recommended}
          </Badge>
        ) : null}
        {model.popular ? <Badge variant="info">{labels.popular}</Badge> : null}
      </>
    );
  }

  return (
    <article
      className={cn(
        "group w-full min-w-0 max-w-full overflow-hidden rounded-lg border border-card-border bg-card text-card-foreground shadow-xs transition-[border-color,box-shadow] hover:border-border hover:shadow-sm",
        model.recommended && "ring-1 ring-primary/30",
        isDeprecated && "border-warning/40 bg-warning/5",
        props.stateBadge?.tone === "active" && "border-success-soft-border",
        props.stateBadge?.tone === "inactive" && "border-border bg-card",
        props.className,
      )}
    >
      <CardHeader className="flex flex-col gap-3 p-4">
        <div data-slot="model-card-title-row" className="flex w-full min-w-0 items-start justify-between gap-3">
          <div className="flex min-w-0 flex-1 items-center gap-3">
            <Avatar className="size-11 shrink-0 border border-border bg-muted">
              <AvatarImage src={model.provider.logoUrl} alt="" aria-hidden="true" />
              <AvatarFallback className="font-semibold">{initials(model.provider.name)}</AvatarFallback>
            </Avatar>
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <div className="flex min-w-0 items-center gap-2">
                <CardTitle className="min-w-0 truncate text-xl leading-7" title={model.name} dir="ltr">
                  {model.name}
                </CardTitle>
                {props.detailsHref ? (
                  <Button asChild variant="ghost" size="icon-sm" className="shrink-0" aria-label={labels.viewDetails} title={labels.viewDetails}>
                    <Link href={props.detailsHref} target="_blank" rel="noreferrer">
                      <ExternalLink aria-hidden="true" />
                    </Link>
                  </Button>
                ) : (
                  <Button variant="ghost" size="icon-sm" className="shrink-0" onClick={() => props.onViewDetails?.(model)} aria-label={labels.viewDetails} title={labels.viewDetails}>
                    <ExternalLink aria-hidden="true" />
                  </Button>
                )}
              </div>
              <div data-slot="model-card-tag-row" className="flex min-w-0 flex-wrap items-center gap-1.5">
                {renderCategoryBadges()}
                {props.hasCustomPricing ? (
                  <Badge className="gap-1" variant="warning">
                    <SlidersHorizontal className="size-3.5" aria-hidden="true" />
                    {labels.customPricing}
                  </Badge>
                ) : null}
                {typeof props.isCompared === "boolean" && props.onComparedChange ? (
                  <label className="inline-flex min-h-8 cursor-pointer items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={props.isCompared}
                      onChange={(event) => props.onComparedChange?.(event.target.checked)}
                      className="size-4"
                    />
                    {props.compareCheckboxLabel ?? "Compare"}
                  </label>
                ) : null}
              </div>
            </div>
          </div>
          {showSelectedControl ? (
            <div className="shrink-0">
              {renderSelectedControl()}
            </div>
          ) : null}
        </div>
      </CardHeader>

      <CardContent className="flex flex-col gap-3 px-4 pb-4 pt-0">
        <div className="grid min-w-0 grid-cols-1 gap-2.5 sm:grid-cols-3">
          <MetricPanel className="sm:col-span-1" label={labels.context} value={formatTokenCount(model.contextWindow)} />
          <PricePanel
            className="sm:col-span-2"
            label={pricingDisplay.scale === "PER_TOKEN" ? labels.pricingPerToken : labels.pricingPerMillion}
            inputValue={formatPriceValueOnly(model.pricing?.inputPerMillion, pricingDisplay)}
            outputValue={formatPriceValueOnly(model.pricing?.outputPerMillion, pricingDisplay)}
            inputLabel={labels.input}
            outputLabel={labels.output}
            action={props.pricingAction}
          />
        </div>

        <div className="grid min-w-0 grid-cols-1 gap-2.5 sm:grid-cols-2">
          <BadgePanel label={labels.inputFormats} values={model.inputModalities} kind="modality" tone="info" />
          <BadgePanel label={labels.outputFormats} values={model.outputModalities} kind="modality" tone="success" />
          <BadgePanel
            label={labels.capabilities}
            values={visibleCapabilities}
            kind="capability"
            moreCount={remainingCapabilityCount}
            formatValue={props.formatCapabilityLabel}
          />
          <BadgePanel
            label={labels.invocationModes}
            values={visibleModes}
            kind="mode"
            moreCount={remainingModeCount}
          />
        </div>

        {model.tradeoffs && model.tradeoffs.length > 0 ? (
          <section className="rounded-xl border border-warning/40 bg-warning/10 p-3">
            <p className="mb-2 flex items-center gap-2 text-sm font-medium text-warning">
              <AlertTriangle aria-hidden="true" />
              {labels.tradeOffs}
            </p>
            <ul className="grid gap-1 text-sm text-muted-foreground">
              {model.tradeoffs.slice(0, 3).map((item) => (
                <li key={item}>- {item}</li>
              ))}
            </ul>
          </section>
        ) : null}

        {props.extraContent}
      </CardContent>

      {hasUtilityActions ? (
        <>
          <Separator />
          <CardFooter data-slot="model-card-actions" className="justify-end p-4 pt-4 md:p-5 md:pt-5">
            {renderUtilityActions()}
          </CardFooter>
        </>
      ) : null}
    </article>
  );
}

function BadgePanel(props: {
  label: string;
  values?: string[];
  kind: "capability" | "mode" | "modality";
  tone?: "neutral" | "info" | "success";
  moreCount?: number;
  formatValue?: (value: string) => string;
}) {
  const tone = props.tone ?? "neutral";

  return (
    <section
      aria-label={props.label}
      className={cn(
        "flex min-w-0 flex-col gap-2.5 rounded-lg border p-3",
        tone === "neutral" && "border-border/60 bg-muted/15",
        tone === "info" && "border-info-soft-border bg-info-soft",
        tone === "success" && "border-success-soft-border bg-success-soft",
      )}
    >
      <p className="text-xs font-medium text-muted-foreground">{props.label}</p>
      <div className="flex min-h-7 flex-wrap gap-2">
        {props.values?.length ? (
          props.values.map((value) => (
            <Badge
              key={`${props.label}-${value}`}
              variant={badgeVariantForValue(props.kind, tone, value)}
              className="gap-1.5"
              title={value}
            >
              {props.kind === "capability" ? iconForCapability(value) : props.kind === "mode" ? iconForMode(value) : iconForModality(value)}
              <bdi dir="auto">{props.formatValue?.(value) ?? value}</bdi>
            </Badge>
          ))
        ) : (
          <span className="text-xs text-muted-foreground">N/A</span>
        )}
        {props.moreCount ? <Badge variant="secondary" className="text-xs">{`+${props.moreCount}`}</Badge> : null}
      </div>
    </section>
  );
}

function PricePanel(props: {
  label: string;
  inputLabel: string;
  outputLabel: string;
  inputValue: string;
  outputValue: string;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <section aria-label={props.label} className={cn("flex min-w-0 flex-col gap-2.5 rounded-lg border border-border/60 bg-muted/15 p-3", props.className)}>
      <div className="flex min-w-0 items-center justify-between gap-2">
        <p className="min-w-0 text-xs font-medium leading-5 text-muted-foreground">{props.label}</p>
        <div className="flex shrink-0 items-center justify-center">
          {props.action ?? <SlidersHorizontal className="size-5 text-muted-foreground" aria-hidden="true" />}
        </div>
      </div>
      <div className="grid min-w-0 grid-cols-2 gap-2">
        <div className="min-w-0 rounded-md bg-background/50 px-2 py-1.5 text-center">
          <p className="text-xs text-muted-foreground">{props.inputLabel}</p>
          <p className="mt-1 truncate text-sm font-semibold" dir="ltr" title={props.inputValue}>{props.inputValue}</p>
        </div>
        <div className="min-w-0 rounded-md bg-background/50 px-2 py-1.5 text-center">
          <p className="text-xs text-muted-foreground">{props.outputLabel}</p>
          <p className="mt-1 truncate text-sm font-semibold" dir="ltr" title={props.outputValue}>{props.outputValue}</p>
        </div>
      </div>
    </section>
  );
}

function MetricPanel(props: { label: string; value: string; className?: string }) {
  return (
    <section aria-label={props.label} className={cn("flex min-h-20 min-w-0 flex-col items-center justify-center rounded-lg border border-border/60 bg-muted/15 p-3 text-center", props.className)}>
      <p className="text-xs font-medium text-muted-foreground">{props.label}</p>
      <p className="mt-2 truncate text-xl font-semibold" dir="ltr" title={props.value}>{props.value}</p>
    </section>
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

function categoryBadgeVariant(
  value: ModelCardData["category"][number],
): "secondary" | "info" {
  return value === "vision" || value === "audio" || value === "multimodal"
    ? "info"
    : "secondary";
}

function badgeVariantForValue(
  kind: "capability" | "mode" | "modality",
  tone: "neutral" | "info" | "success",
  value: string,
): "secondary" | "info" | "success" | "warning" {
  if (kind === "modality") return tone === "success" ? "success" : "info";
  if (kind === "mode") return "secondary";

  const key = value.toLowerCase();
  if (key.includes("json") || key.includes("structured")) return "success";
  if (key.includes("cache")) return "warning";
  if (key.includes("tool") || key.includes("function") || key.includes("code")) return "info";
  return "secondary";
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
  if (typeof usdPerMillion !== "number" || !Number.isFinite(usdPerMillion)) return "N/A";

  const value = display.scale === "PER_TOKEN" ? usdPerMillion / 1_000_000 : usdPerMillion;
  if (display.currency === "TOMAN") {
    const toman = (value * display.usdToIrr) / 10;
    return `${new Intl.NumberFormat("en-US", {
      maximumFractionDigits: display.scale === "PER_TOKEN" ? 2 : 0,
    }).format(toman)} toman`;
  }

  return `$${trimZeros(value)}`;
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
