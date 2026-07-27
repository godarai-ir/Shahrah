"use client";

import { useMemo, useState, type ReactNode } from "react";
import {
  Banknote,
  BookOpen,
  Bot,
  Cpu,
  ExternalLink,
  Gauge,
  Link2,
  ListTree,
  type LucideIcon,
  Network,
  ShieldCheck,
  SlidersHorizontal,
  Sparkles,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { CompareToggleButton } from "@/components/compare/compare-toggle-button";
import { CopyPageLink } from "@/components/catalog/copy-page-link";
import { DisplayPrice } from "@/components/catalog/display-price";
import { ModalityFlow } from "@/components/catalog/modality-icon-list";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatCompactNumber } from "@/lib/number-format";
import { cn } from "@/lib/utils";
import { useDisplayFormatters } from "@/lib/use-display-formatters";

type Provider = {
  slug?: string | null;
  name?: string | null;
  status?: string | null;
  docsUrl?: string | null;
  websiteUrl?: string | null;
  privacyPolicyUrl?: string | null;
  supportsByok?: boolean | null;
  supportsRegionalRouting?: boolean | null;
};

type Pricing = {
  id: string;
  currency?: string | null;
  inputPricePerMillionTokens?: number | null;
  outputPricePerMillionTokens?: number | null;
  cachedInputPricePerMillionTokens?: number | null;
  effectiveFrom?: string | null;
  effectiveTo?: string | null;
  sourceUrl?: string | null;
  region?: { slug?: string | null; name?: string | null } | null;
};

type ModelDetail = {
  canonicalId: string;
  slug: string;
  displayName: string;
  shortDescription?: string | null;
  longDescription?: string | null;
  status: string;
  lifecycleStatus?: string | null;
  contextWindow?: number | null;
  maxOutputTokens?: number | null;
  knowledgeCutoff?: string | null;
  releaseDate?: string | null;
  deprecatedAt?: string | null;
  tokenizer?: string | null;
  license?: string | null;
  officialDocsUrl?: string | null;
  provider?: Provider | null;
  pricing?: Pricing[] | null;
  inputModalities?: Array<{ modality?: { slug?: string | null; name?: string | null } | null }> | null;
  outputModalities?: Array<{ modality?: { slug?: string | null; name?: string | null } | null }> | null;
  supportedParams?: Array<{ supportedParameter?: { slug?: string | null; name?: string | null } | null }> | null;
  capabilities?: Array<{ capability?: { slug?: string | null; name?: string | null } | null }> | null;
  regions?: Array<{ region?: { slug?: string | null; name?: string | null } | null }> | null;
  sources?: Array<{ id?: string; title?: string | null; url?: string | null; sourceType?: string | null; notes?: string | null }> | null;
  changelog?: Array<{ id?: string; changeType?: string | null; title?: string | null; description?: string | null; changedAt?: string | null; createdBy?: string | null }> | null;
  benchmarks?: Array<{ id?: string; score?: number | null; percentile?: number | null; evaluatedAt?: string | null; benchmark?: { slug?: string | null; name?: string | null } | null }> | null;
};

type StatusKey =
  | "active"
  | "inactive"
  | "disabled"
  | "maintenance"
  | "deprecated"
  | "preview"
  | "available"
  | "unavailable"
  | "draft"
  | "review"
  | "published"
  | "archived";
const EMPTY = "-";

function getStatusKey(status: string): StatusKey | null {
  const normalized = status.toLowerCase().replaceAll("_", "-");
  const keys: StatusKey[] = [
    "active",
    "inactive",
    "disabled",
    "maintenance",
    "deprecated",
    "preview",
    "available",
    "unavailable",
    "draft",
    "review",
    "published",
    "archived",
  ];
  return keys.includes(normalized as StatusKey) ? (normalized as StatusKey) : null;
}

function safeUrl(url?: string | null): string | null {
  if (!url) return null;

  try {
    const parsed = new URL(url);
    return parsed.protocol === "http:" || parsed.protocol === "https:" ? parsed.toString() : null;
  } catch {
    return null;
  }
}

function SectionCard({
  id,
  label,
  icon: Icon,
  children,
  className,
}: {
  id: string;
  label: string;
  icon: LucideIcon;
  children: React.ReactNode;
  className?: string;
}) {
  const t = useTranslations("ModelDetail");

  return (
    <Card id={id} className={cn("scroll-mt-24 p-0", className)}>
      <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-3 sm:px-5">
        <div className="flex min-w-0 items-center gap-2.5">
          <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary-soft-foreground">
            <Icon className="size-4" aria-hidden="true" />
          </span>
          <h2 className="text-base font-semibold sm:text-lg">{label}</h2>
        </div>
        <a
          href={`#${id}`}
          aria-label={t("sectionAnchor", { section: label })}
          className="rounded-md px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          #
        </a>
      </div>
      <div className="p-4 sm:p-5">{children}</div>
    </Card>
  );
}

function KeyValue({ label, value, technical = false }: { label: string; value: React.ReactNode; technical?: boolean }) {
  return (
    <div className="grid min-w-0 gap-1 py-2.5 text-sm sm:grid-cols-5 sm:items-start sm:gap-4">
      <dt className="text-muted-foreground sm:col-span-2">{label}</dt>
      <dd
        dir={technical ? "ltr" : "auto"}
        data-technical={technical || undefined}
        className={cn(
          "min-w-0 font-medium sm:col-span-3",
          technical && "overflow-x-auto whitespace-nowrap text-xs leading-6",
        )}
      >
        {value}
      </dd>
    </div>
  );
}

function ChipList({ values, emptyLabel, initial = 10 }: { values: string[]; emptyLabel: string; initial?: number }) {
  const t = useTranslations("ModelDetail");
  const [expanded, setExpanded] = useState(false);

  if (values.length === 0) return <p className="text-sm text-muted-foreground">{emptyLabel}</p>;

  const shown = expanded ? values : values.slice(0, initial);

  return (
    <div className="flex flex-col items-start gap-3">
      <div className="flex flex-wrap gap-2">
        {shown.map((value) => (
          <Badge key={value} className="px-2.5 py-1">
            <bdi dir="auto">{value}</bdi>
          </Badge>
        ))}
      </div>
      {values.length > initial ? (
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => setExpanded((value) => !value)}
        >
          {expanded ? t("showLess") : t("showMore", { count: values.length - initial })}
        </Button>
      ) : null}
    </div>
  );
}

export function ModelDetailView({ model, requestedCanonicalId, fetchError }: { model: ModelDetail | null; requestedCanonicalId: string; fetchError?: string | null }) {
  const t = useTranslations("ModelDetail");
  const tCommon = useTranslations("Common");
  const tStatus = useTranslations("Statuses");
  const { formatDate } = useDisplayFormatters();
  const numberFormatter = useMemo(() => new Intl.NumberFormat("fa-IR", { maximumFractionDigits: 2 }), []);

  const normalized = useMemo(() => {
    const inputModalities = (model?.inputModalities ?? []).map((entry) => entry.modality?.name || entry.modality?.slug).filter((value): value is string => Boolean(value));
    const outputModalities = (model?.outputModalities ?? []).map((entry) => entry.modality?.name || entry.modality?.slug).filter((value): value is string => Boolean(value));
    const capabilities = (model?.capabilities ?? []).map((entry) => entry.capability?.name || entry.capability?.slug).filter((value): value is string => Boolean(value));
    const supportedParams = (model?.supportedParams ?? []).map((entry) => {
      const slug = entry.supportedParameter?.slug;
      const name = entry.supportedParameter?.name;
      if (!slug && !name) return null;
      return slug && name ? `${name} (${slug})` : (name || slug) as string;
    }).filter((value): value is string => Boolean(value));
    const sortedPricing = [...(model?.pricing ?? [])].sort((a, b) => (a.region?.slug || "global").localeCompare(b.region?.slug || "global"));
    return { inputModalities, outputModalities, capabilities, supportedParams, sortedPricing };
  }, [model]);

  const formatStatus = (status?: string | null) => {
    if (!status) return EMPTY;
    const key = getStatusKey(status);
    return key ? tStatus(key) : status;
  };
  const booleanLabel = (value?: boolean | null) => value === null || value === undefined ? EMPTY : value ? tCommon("yes") : tCommon("no");
  const formatCompactMetric = (value?: number | null) => value === null || value === undefined
    ? EMPTY
    : (
        <bdi dir="ltr" title={numberFormatter.format(value)}>
          {formatCompactNumber(value)}
        </bdi>
      );

  if (!model) {
    const notFound = fetchError?.includes("API 404");
    return (
      <Card className="mx-auto flex max-w-2xl flex-col items-start gap-4 p-6 sm:p-8">
        <span className="flex size-11 items-center justify-center rounded-xl bg-destructive-soft text-destructive-soft-foreground">
          <Bot className="size-5" aria-hidden="true" />
        </span>
        <div className="flex flex-col gap-2">
          <h1 className="text-xl font-semibold sm:text-2xl">{notFound ? t("notFound") : t("loadError")}</h1>
          <p className="text-sm leading-6 text-muted-foreground">
            {t("requestedCanonicalId")}: <bdi dir="ltr" data-technical className="break-all">{requestedCanonicalId}</bdi>
          </p>
        </div>
      </Card>
    );
  }

  const sections = [
    ["core-specs", t("sections.coreSpecs")],
    ["provider-governance", t("sections.providerGovernance")],
    ["modalities", t("sections.modalities")],
    ["capabilities", t("sections.capabilities")],
    ["supported-parameters", t("sections.supportedParameters")],
    ["pricing", t("sections.pricing")],
    ["sources", t("sections.sources")],
  ] as const;

  const modelDocsUrl = safeUrl(model.officialDocsUrl);
  const providerDocsUrl = safeUrl(model.provider?.docsUrl);
  const modelDescription = model.longDescription?.trim() || model.shortDescription?.trim();
  const modelStatusKey = getStatusKey(model.status);
  const statusClassName = modelStatusKey === "active" || modelStatusKey === "available"
    ? "border-success-soft bg-success-soft text-success-soft-foreground"
    : undefined;

  const metrics = [
    { label: t("contextWindow"), value: formatCompactMetric(model.contextWindow), icon: Gauge },
    { label: t("maxOutputTokens"), value: formatCompactMetric(model.maxOutputTokens), icon: Cpu },
    { label: t("baselineInputPrice"), value: <DisplayPrice value={model.pricing?.[0]?.inputPricePerMillionTokens} />, icon: Banknote },
    { label: t("baselineOutputPrice"), value: <DisplayPrice value={model.pricing?.[0]?.outputPricePerMillionTokens} />, icon: Banknote },
  ];

  return (
    <div className="grid min-w-0 gap-5 xl:grid-cols-4 xl:items-start">
      <div className="flex min-w-0 flex-col gap-5 xl:col-span-3">
        <Card className="overflow-hidden p-0">
          <div className="flex flex-col gap-5 p-4 sm:p-6">
            <div className="flex min-w-0 flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
              <div className="flex min-w-0 flex-1 flex-col gap-3">
                <div className="flex min-w-0 items-center gap-3">
                  <span className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary-soft-foreground">
                    <Bot className="size-5" aria-hidden="true" />
                  </span>
                  <div className="flex min-w-0 items-center gap-2">
                    <h1 className="min-w-0 text-2xl font-bold leading-tight tracking-tight sm:text-3xl">
                      <bdi dir="auto">{model.displayName}</bdi>
                    </h1>
                    <CopyPageLink
                      label={t("copyLink")}
                      copiedLabel={t("copySuccess")}
                      failedLabel={t("copyError")}
                    />
                  </div>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm font-semibold">
                    <bdi dir="auto">{model.provider?.name || model.provider?.slug || t("unknownProvider")}</bdi>
                  </span>
                  <Badge className={statusClassName}>{formatStatus(model.status)}</Badge>
                  {model.lifecycleStatus ? <Badge>{formatStatus(model.lifecycleStatus)}</Badge> : null}
                </div>
                {modelDescription ? (
                  <p dir="auto" className="max-w-3xl text-sm leading-7 text-muted-foreground sm:text-base">
                    {modelDescription}
                  </p>
                ) : null}
              </div>

              <div className="flex shrink-0 flex-wrap items-center gap-2">
                {modelDocsUrl ? (
                  <Button asChild variant="outline">
                    <a href={modelDocsUrl} target="_blank" rel="noreferrer">
                      <ExternalLink data-icon="inline-start" aria-hidden="true" />
                      {t("officialDocs")}
                    </a>
                  </Button>
                ) : null}
                <CompareToggleButton canonicalId={model.canonicalId} compact />
              </div>
            </div>

          </div>

          <dl className="grid grid-cols-2 border-t border-border sm:grid-cols-4">
            {metrics.map(({ label, value, icon: Icon }, index) => (
              <div
                key={label}
                className={cn(
                  "flex min-w-0 flex-col gap-2 border-border p-4 sm:p-5",
                  index % 2 === 0 && "border-e",
                  index >= 2 && "border-t",
                  index < 3 && "sm:border-e",
                  index >= 2 && "sm:border-t-0",
                )}
              >
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Icon className="size-4 shrink-0" aria-hidden="true" />
                  <dt>{label}</dt>
                </div>
                <dd dir="auto" className="min-w-0 text-sm font-semibold leading-6 tabular-nums sm:text-base">{value}</dd>
              </div>
            ))}
          </dl>
        </Card>

        <Card className="sticky top-16 z-20 bg-card/95 p-2 shadow-sm backdrop-blur xl:hidden">
          <nav aria-label={t("onThisPage")} className="scrollbar-none flex min-w-0 items-center gap-1 overflow-x-auto">
            <span className="shrink-0 px-2 text-xs font-semibold text-foreground">{t("onThisPage")}</span>
            {sections.map(([id, label]) => (
              <a
                key={id}
                href={`#${id}`}
                className="shrink-0 rounded-md px-2.5 py-1.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {label}
              </a>
            ))}
          </nav>
        </Card>

        <div className="grid min-w-0 gap-5 lg:grid-cols-2">
          <SectionCard id="core-specs" label={t("sections.coreSpecs")} icon={ListTree}>
            <dl className="divide-y divide-border">
              <KeyValue label={t("fields.status")} value={formatStatus(model.status)} />
              <KeyValue label={t("fields.lifecycle")} value={formatStatus(model.lifecycleStatus)} />
              <KeyValue label={t("fields.releaseDate")} value={formatDate(model.releaseDate)} />
              <KeyValue label={t("fields.deprecatedAt")} value={formatDate(model.deprecatedAt)} />
              <KeyValue label={t("fields.knowledgeCutoff")} value={formatDate(model.knowledgeCutoff)} />
              <KeyValue label={t("fields.license")} value={model.license || EMPTY} />
            </dl>
          </SectionCard>

          <SectionCard id="provider-governance" label={t("sections.providerGovernance")} icon={ShieldCheck}>
            <dl className="divide-y divide-border">
              <KeyValue label={t("fields.providerStatus")} value={formatStatus(model.provider?.status)} />
              <KeyValue label={t("fields.byokSupport")} value={booleanLabel(model.provider?.supportsByok)} />
              <KeyValue label={t("fields.regionalRouting")} value={booleanLabel(model.provider?.supportsRegionalRouting)} />
              <KeyValue label={t("fields.modelDocs")} value={modelDocsUrl ? <ExternalResource href={modelDocsUrl} label={tCommon("open")} /> : EMPTY} />
              <KeyValue label={t("fields.providerDocs")} value={providerDocsUrl ? <ExternalResource href={providerDocsUrl} label={tCommon("open")} /> : EMPTY} />
            </dl>
          </SectionCard>

          <SectionCard id="modalities" label={t("sections.modalities")} icon={Network}>
            <ModalityFlow
              inputValues={normalized.inputModalities}
              outputValues={normalized.outputModalities}
              inputLabel={t("input")}
              outputLabel={t("output")}
              emptyLabel={EMPTY}
            />
          </SectionCard>

          <SectionCard id="capabilities" label={t("sections.capabilities")} icon={Sparkles}>
            <ChipList values={normalized.capabilities} emptyLabel={t("noCapabilities")} />
          </SectionCard>

          <SectionCard id="supported-parameters" label={t("sections.supportedParameters")} icon={SlidersHorizontal} className="lg:col-span-2">
            <ChipList values={normalized.supportedParams} emptyLabel={t("noSupportedParameters")} initial={14} />
          </SectionCard>

          <SectionCard id="pricing" label={t("sections.pricing")} icon={Banknote} className="lg:col-span-2">
            {normalized.sortedPricing.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("noPricing")}</p>
            ) : (
              <div className="grid gap-3">
                {normalized.sortedPricing.map((price) => {
                  const sourceUrl = safeUrl(price.sourceUrl);
                  const priceValues: Array<[string, ReactNode]> = [
                    [t("input"), <DisplayPrice key="input-price" value={price.inputPricePerMillionTokens} />],
                    [t("output"), <DisplayPrice key="output-price" value={price.outputPricePerMillionTokens} />],
                    [t("cachedInput"), <DisplayPrice key="cached-input-price" value={price.cachedInputPricePerMillionTokens} />],
                    [t("effectiveFrom"), formatDate(price.effectiveFrom)],
                  ];

                  return (
                    <article key={price.id} className="flex min-w-0 flex-col gap-4 rounded-lg border border-border bg-background/60 p-4">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                          <span className="flex size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                            <Network className="size-4" aria-hidden="true" />
                          </span>
                          <div>
                            <p className="text-xs text-muted-foreground">{t("region")}</p>
                            <p className="text-sm font-semibold"><bdi dir="auto">{price.region?.name || price.region?.slug || tCommon("global")}</bdi></p>
                          </div>
                        </div>
                        {sourceUrl ? <ExternalResource href={sourceUrl} label={t("source")} /> : null}
                      </div>
                      <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                        {priceValues.map(([label, value]) => (
                          <div key={label} className="min-w-0 rounded-md bg-muted/60 p-3">
                            <dt className="text-xs text-muted-foreground">{label}</dt>
                            <dd className="mt-1 text-sm font-semibold leading-6">{value}</dd>
                          </div>
                        ))}
                      </dl>
                    </article>
                  );
                })}
              </div>
            )}
          </SectionCard>

          <SectionCard id="sources" label={t("sections.sources")} icon={BookOpen} className="lg:col-span-2">
            {(model.sources ?? []).length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("noSources")}</p>
            ) : (
              <div className="grid gap-3 sm:grid-cols-2">
                {(model.sources ?? []).map((source, index) => {
                  const sourceUrl = safeUrl(source.url);
                  return (
                    <article
                      key={source.id || `${source.url || "source"}-${index}`}
                      dir="rtl"
                      className="flex min-w-0 flex-col gap-2 rounded-lg bg-background/60 p-4 text-start"
                    >
                      <div className="flex items-start gap-3">
                        <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                          <Link2 className="size-4" aria-hidden="true" />
                        </span>
                        <div className="min-w-0 flex-1">
                          <h3 className="font-medium">{source.title || tCommon("source")}</h3>
                          <p className="text-xs text-muted-foreground">{source.sourceType || EMPTY}</p>
                        </div>
                      </div>
                      {sourceUrl ? <ExternalResource href={sourceUrl} label={tCommon("open")} showUrl /> : <p>{EMPTY}</p>}
                      {source.notes ? <p className="text-sm leading-6 text-muted-foreground">{source.notes}</p> : null}
                    </article>
                  );
                })}
              </div>
            )}
          </SectionCard>
        </div>
      </div>

      <aside className="hidden self-start xl:sticky xl:top-24 xl:block">
        <Card className="p-3">
          <div className="flex items-center gap-2 px-2 pb-2">
            <ListTree className="size-4 text-muted-foreground" aria-hidden="true" />
            <h2 className="text-sm font-semibold">{t("onThisPage")}</h2>
          </div>
          <nav className="flex flex-col gap-1" aria-label={t("onThisPage")}>
            {sections.map(([id, label]) => (
              <a
                key={id}
                href={`#${id}`}
                className="rounded-md px-2.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {label}
              </a>
            ))}
          </nav>
        </Card>
      </aside>
    </div>
  );
}

function ExternalResource({ href, label, showUrl = false }: { href: string; label: string; showUrl?: boolean }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      dir={showUrl ? "ltr" : undefined}
      className={cn(
        "inline-flex min-w-0 items-center gap-1.5 text-sm font-medium text-primary transition-colors hover:text-primary/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        showUrl && "max-w-full overflow-x-auto whitespace-nowrap text-xs",
      )}
    >
      <span>{showUrl ? href : label}</span>
      <ExternalLink className="size-3.5 shrink-0" aria-hidden="true" />
    </a>
  );
}
