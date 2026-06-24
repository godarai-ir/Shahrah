"use client";

import { useMemo, useState } from "react";
import {
  AudioLines,
  Code2,
  FileText,
  Image as ImageIcon,
  type LucideIcon,
  Video,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  formatDisplayDate,
  formatDisplayPrice,
} from "@/lib/display-settings";
import { useDisplaySettings } from "@/components/display-settings-provider";
import { CompareToggleButton } from "@/components/compare/compare-toggle-button";

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

const EMPTY = "-";

function fmtNumber(value?: number | null): string {
  if (value === null || value === undefined) return EMPTY;
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 2 }).format(value);
}

function safeUrl(url?: string | null): string | null {
  if (!url) return null;
  try {
    const u = new URL(url);
    if (u.protocol !== "http:" && u.protocol !== "https:") return null;
    return u.toString();
  } catch {
    return null;
  }
}

function SectionTitle({ id, children }: { id: string; children: React.ReactNode }) {
  return (
    <div id={id} className="mb-3 flex items-center justify-between">
      <h2 className="text-base font-semibold text-zinc-100">{children}</h2>
      <a href={`#${id}`} className="text-xs text-zinc-500 hover:text-zinc-300">
        #
      </a>
    </div>
  );
}

function KeyValue({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[160px_1fr] gap-3 py-1 text-sm">
      <span className="text-zinc-500">{label}</span>
      <span className="text-zinc-200 break-all">{value}</span>
    </div>
  );
}

function ChipList({ values, emptyLabel, initial = 10 }: { values: string[]; emptyLabel: string; initial?: number }) {
  const [expanded, setExpanded] = useState(false);
  if (values.length === 0) return <p className="text-sm text-zinc-500">{emptyLabel}</p>;
  const shown = expanded ? values : values.slice(0, initial);
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-1.5">
        {shown.map((value) => (
          <Badge key={value} className="border-zinc-700 text-zinc-300">
            {value}
          </Badge>
        ))}
      </div>
      {values.length > initial ? (
        <Button type="button" className="h-8 bg-zinc-800 px-2 py-1 text-xs text-zinc-100 hover:bg-zinc-700" onClick={() => setExpanded((v) => !v)}>
          {expanded ? "Show less" : `Show more (${values.length - initial})`}
        </Button>
      ) : null}
    </div>
  );
}

type ModalityMeta = {
  icon: LucideIcon;
  className: string;
};

function getModalityMeta(value: string): ModalityMeta {
  const key = value.toLowerCase();
  if (key.includes("text")) return { icon: FileText, className: "border-blue-500/40 bg-blue-500/10 text-blue-300" };
  if (key.includes("image")) return { icon: ImageIcon, className: "border-fuchsia-500/40 bg-fuchsia-500/10 text-fuchsia-300" };
  if (key.includes("audio") || key.includes("speech") || key.includes("voice")) {
    return { icon: AudioLines, className: "border-amber-500/40 bg-amber-500/10 text-amber-300" };
  }
  if (key.includes("video")) return { icon: Video, className: "border-rose-500/40 bg-rose-500/10 text-rose-300" };
  if (key.includes("embedding")) return { icon: Code2, className: "border-emerald-500/40 bg-emerald-500/10 text-emerald-300" };
  return { icon: Code2, className: "border-zinc-600 bg-zinc-800/70 text-zinc-300" };
}

function ModalityIconList({ values, emptyLabel }: { values: string[]; emptyLabel: string }) {
  if (values.length === 0) return <p className="text-sm text-zinc-500">{emptyLabel}</p>;

  return (
    <div className="flex flex-wrap gap-2">
      {values.map((value) => {
        const { icon: Icon, className } = getModalityMeta(value);
        return (
          <span key={value} className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs ${className}`}>
            <Icon className="h-3.5 w-3.5" />
            <span>{value}</span>
          </span>
        );
      })}
    </div>
  );
}

export function ModelDetailView({
  model,
  requestedCanonicalId,
  fetchError,
}: {
  model: ModelDetail | null;
  requestedCanonicalId: string;
  fetchError?: string | null;
}) {
  const [copiedField, setCopiedField] = useState<"canonical" | "slug" | null>(null);
  const { settings } = useDisplaySettings();

  const normalized = useMemo(() => {
    const inputModalities = (model?.inputModalities ?? [])
      .map((entry) => entry.modality?.name || entry.modality?.slug)
      .filter((x): x is string => Boolean(x));
    const outputModalities = (model?.outputModalities ?? [])
      .map((entry) => entry.modality?.name || entry.modality?.slug)
      .filter((x): x is string => Boolean(x));
    const capabilities = (model?.capabilities ?? [])
      .map((entry) => entry.capability?.name || entry.capability?.slug)
      .filter((x): x is string => Boolean(x));
    const supportedParams = (model?.supportedParams ?? [])
      .map((entry) => {
        const slug = entry.supportedParameter?.slug;
        const name = entry.supportedParameter?.name;
        if (!slug && !name) return null;
        return slug && name ? `${name} (${slug})` : (name || slug) as string;
      })
      .filter((x): x is string => Boolean(x));
    const sortedPricing = [...(model?.pricing ?? [])].sort((a, b) =>
      (a.region?.slug || "global").localeCompare(b.region?.slug || "global"),
    );

    return { inputModalities, outputModalities, capabilities, supportedParams, sortedPricing };
  }, [model]);

  async function copyText(text: string, field: "canonical" | "slug") {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 1200);
    } catch {
      setCopiedField(null);
    }
  }

  if (!model) {
    const notFound = fetchError?.includes("API 404");
    return (
      <Card className="space-y-3">
        <h1 className="text-xl font-semibold">{notFound ? "Model not found" : "Unable to load model details"}</h1>
        <p className="text-sm text-zinc-300">Requested canonical ID: <code>{requestedCanonicalId}</code></p>
        {fetchError ? <pre className="overflow-x-auto rounded-lg border border-zinc-800 bg-zinc-950 p-3 text-xs text-amber-300">{fetchError}</pre> : null}
      </Card>
    );
  }

  const sections = [
    ["core-specs", "Core Specs"],
    ["provider-governance", "Provider & Governance"],
    ["modalities", "Modalities"],
    ["capabilities", "Capabilities"],
    ["supported-parameters", "Supported Parameters"],
    ["pricing", "Pricing"],
    ["sources", "Sources"],
  ] as const;

  return (
    <div className="grid gap-4 lg:grid-cols-[1fr_240px]">
      <div className="space-y-4">
        <Card className="space-y-4">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h1 className="text-2xl font-semibold">{model.displayName}</h1>
              <p className="mt-1 text-sm text-zinc-400">{model.longDescription || model.shortDescription || "No description available."}</p>
            </div>
            <div className="flex flex-wrap gap-1.5">
              <Badge className="border-zinc-700 text-zinc-200">{model.status || EMPTY}</Badge>
              <Badge className="border-zinc-700 text-zinc-200">{model.lifecycleStatus || EMPTY}</Badge>
              <Badge className="border-zinc-700 text-zinc-200">{model.provider?.name || model.provider?.slug || "Unknown provider"}</Badge>
            </div>
          </div>

          <div className="flex flex-wrap gap-2 text-sm text-zinc-300">
            <span className="rounded border border-zinc-800 bg-zinc-950 px-2 py-1">
              canonicalId: <code>{model.canonicalId}</code>
            </span>
            <CompareToggleButton canonicalId={model.canonicalId} />
            <Button type="button" className="h-8 bg-zinc-800 px-2 py-1 text-xs text-zinc-100 hover:bg-zinc-700" onClick={() => copyText(model.canonicalId, "canonical")}>
              {copiedField === "canonical" ? "Copied" : "Copy canonicalId"}
            </Button>
            <span className="rounded border border-zinc-800 bg-zinc-950 px-2 py-1">
              slug: <code>{model.slug}</code>
            </span>
            <Button type="button" className="h-8 bg-zinc-800 px-2 py-1 text-xs text-zinc-100 hover:bg-zinc-700" onClick={() => copyText(model.slug, "slug")}>
              {copiedField === "slug" ? "Copied" : "Copy slug"}
            </Button>
          </div>

          <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
            <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs text-zinc-500">Context window</p>
              <p className="text-base font-medium">{fmtNumber(model.contextWindow)}</p>
            </div>
            <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs text-zinc-500">Max output tokens</p>
              <p className="text-base font-medium">{fmtNumber(model.maxOutputTokens)}</p>
            </div>
            <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs text-zinc-500">Baseline input price / 1M</p>
              <p className="text-base font-medium">{formatDisplayPrice(model.pricing?.[0]?.inputPricePerMillionTokens, settings)}</p>
            </div>
            <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs text-zinc-500">Baseline output price / 1M</p>
              <p className="text-base font-medium">{formatDisplayPrice(model.pricing?.[0]?.outputPricePerMillionTokens, settings)}</p>
            </div>
            <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs text-zinc-500">Modalities</p>
              <p className="text-base font-medium">{normalized.inputModalities.length} in / {normalized.outputModalities.length} out</p>
            </div>
            <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-3">
              <p className="text-xs text-zinc-500">Capabilities</p>
              <p className="text-base font-medium">{normalized.capabilities.length}</p>
            </div>
          </div>
        </Card>

        <div className="grid gap-4 xl:grid-cols-2">
          <Card>
            <SectionTitle id="core-specs">Core Specs</SectionTitle>
            <KeyValue label="Canonical ID" value={model.canonicalId} />
            <KeyValue label="Slug" value={model.slug} />
            <KeyValue label="Status" value={model.status || EMPTY} />
            <KeyValue label="Lifecycle" value={model.lifecycleStatus || EMPTY} />
            <KeyValue label="Release Date" value={formatDisplayDate(model.releaseDate, settings)} />
            <KeyValue label="Deprecated At" value={formatDisplayDate(model.deprecatedAt, settings)} />
            <KeyValue label="Knowledge Cutoff" value={formatDisplayDate(model.knowledgeCutoff, settings)} />
            <KeyValue label="Tokenizer" value={model.tokenizer || EMPTY} />
            <KeyValue label="License" value={model.license || EMPTY} />
          </Card>

          <Card>
            <SectionTitle id="provider-governance">Provider & Governance</SectionTitle>
            <KeyValue label="Provider Name" value={model.provider?.name || EMPTY} />
            <KeyValue label="Provider Slug" value={model.provider?.slug || EMPTY} />
            <KeyValue label="Provider Status" value={model.provider?.status || EMPTY} />
            <KeyValue label="BYOK Support" value={model.provider?.supportsByok ? "Yes" : "No"} />
            <KeyValue label="Regional Routing" value={model.provider?.supportsRegionalRouting ? "Yes" : "No"} />
            <KeyValue
              label="Model Docs"
              value={safeUrl(model.officialDocsUrl) ? <a className="text-cyan-400 hover:text-cyan-300" href={safeUrl(model.officialDocsUrl) || "#"} target="_blank" rel="noreferrer">Open</a> : EMPTY}
            />
            <KeyValue
              label="Provider Docs"
              value={safeUrl(model.provider?.docsUrl) ? <a className="text-cyan-400 hover:text-cyan-300" href={safeUrl(model.provider?.docsUrl) || "#"} target="_blank" rel="noreferrer">Open</a> : EMPTY}
            />
            <KeyValue
              label="Privacy Policy"
              value={safeUrl(model.provider?.privacyPolicyUrl) ? <a className="text-cyan-400 hover:text-cyan-300" href={safeUrl(model.provider?.privacyPolicyUrl) || "#"} target="_blank" rel="noreferrer">Open</a> : EMPTY}
            />
          </Card>

          <Card>
            <SectionTitle id="modalities">Modalities</SectionTitle>
            <p className="mb-2 text-xs uppercase tracking-wide text-zinc-500">Input</p>
            <ModalityIconList values={normalized.inputModalities} emptyLabel="No input modalities." />
            <p className="mb-2 mt-4 text-xs uppercase tracking-wide text-zinc-500">Output</p>
            <ModalityIconList values={normalized.outputModalities} emptyLabel="No output modalities." />
          </Card>

          <Card>
            <SectionTitle id="capabilities">Capabilities</SectionTitle>
            <ChipList values={normalized.capabilities} emptyLabel="No capabilities listed." />
          </Card>

          <Card className="xl:col-span-2">
            <SectionTitle id="supported-parameters">Supported Parameters</SectionTitle>
            <ChipList values={normalized.supportedParams} emptyLabel="No supported parameters listed." initial={14} />
          </Card>

          <Card className="xl:col-span-2">
            <SectionTitle id="pricing">Pricing</SectionTitle>
            {normalized.sortedPricing.length === 0 ? (
              <p className="text-sm text-zinc-500">No pricing entries.</p>
            ) : (
              <div className="space-y-2">
                {normalized.sortedPricing.map((p) => (
                  <div key={p.id} className="grid gap-2 rounded-lg border border-zinc-800 bg-zinc-950 p-3 text-sm md:grid-cols-5">
                    <div><span className="text-zinc-500">Region:</span> {p.region?.name || p.region?.slug || "global"}</div>
                    <div><span className="text-zinc-500">Input:</span> {formatDisplayPrice(p.inputPricePerMillionTokens, settings)}</div>
                    <div><span className="text-zinc-500">Output:</span> {formatDisplayPrice(p.outputPricePerMillionTokens, settings)}</div>
                    <div><span className="text-zinc-500">Cached:</span> {formatDisplayPrice(p.cachedInputPricePerMillionTokens, settings)}</div>
                    <div><span className="text-zinc-500">Effective:</span> {formatDisplayDate(p.effectiveFrom, settings)}</div>
                    <div className="md:col-span-5">
                      <span className="text-zinc-500">Source:</span>{" "}
                      {safeUrl(p.sourceUrl) ? (
                        <a href={safeUrl(p.sourceUrl) || "#"} target="_blank" rel="noreferrer" className="text-cyan-400 hover:text-cyan-300">
                          {p.sourceUrl}
                        </a>
                      ) : (
                        EMPTY
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>

          <Card>
            <SectionTitle id="sources">Sources</SectionTitle>
            {(model.sources ?? []).length === 0 ? (
              <p className="text-sm text-zinc-500">No sources.</p>
            ) : (
              <div className="space-y-2 text-sm">
                {(model.sources ?? []).map((source, i) => (
                  <div key={source.id || `${source.url || "source"}-${i}`} className="rounded border border-zinc-800 bg-zinc-950 p-2">
                    <div className="font-medium">{source.title || "Source"}</div>
                    <div className="text-zinc-500">{source.sourceType || EMPTY}</div>
                    {safeUrl(source.url) ? <a href={safeUrl(source.url) || "#"} target="_blank" rel="noreferrer" className="break-all text-cyan-400 hover:text-cyan-300">{source.url}</a> : <div>{EMPTY}</div>}
                    {source.notes ? <div className="mt-1 text-zinc-400">{source.notes}</div> : null}
                  </div>
                ))}
              </div>
            )}
          </Card>

        </div>
      </div>

      <aside className="hidden lg:block">
        <Card className="sticky top-20">
          <h3 className="mb-2 text-sm font-semibold text-zinc-200">On this page</h3>
          <nav className="space-y-1 text-sm">
            {sections.map(([id, label]) => (
              <a key={id} href={`#${id}`} className="block rounded px-2 py-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200">
                {label}
              </a>
            ))}
          </nav>
        </Card>
      </aside>
    </div>
  );
}

