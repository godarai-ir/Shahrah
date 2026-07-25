"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
  ArrowLeft,
  Boxes,
  Cable,
  CircleAlert,
  KeyRound,
  Map,
  Search,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { ProviderMark } from "@/components/catalog/provider-mark";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

export type ProviderDirectoryItem = {
  id?: string;
  slug: string;
  name: string;
  description?: string | null;
  docsUrl?: string | null;
  logoUrl?: string | null;
  status?: string;
  modelCount?: number;
  modes: string[];
  supportsByok?: boolean;
  supportsRegionalRouting?: boolean;
};

const MODE_GROUPS = [
  {
    key: "inference",
    modes: ["chat", "responses", "completion", "embedding", "rerank", "moderation"],
  },
  {
    key: "multimodal",
    modes: ["image", "video", "text_to_speech", "audio_transcription", "audio_translation", "realtime"],
  },
] as const;

const MODE_LABEL_KEYS = {
  chat: "modes.chat",
  responses: "modes.responses",
  completion: "modes.completion",
  embedding: "modes.embedding",
  rerank: "modes.rerank",
  moderation: "modes.moderation",
  image: "modes.image",
  video: "modes.video",
  text_to_speech: "modes.text_to_speech",
  audio_transcription: "modes.audio_transcription",
  audio_translation: "modes.audio_translation",
  realtime: "modes.realtime",
} as const;

export function ProvidersExplorer({
  providers,
  hasError,
}: {
  providers: ProviderDirectoryItem[];
  hasError: boolean;
}) {
  const t = useTranslations("Providers");
  const formatter = useMemo(() => new Intl.NumberFormat("fa-IR"), []);
  const availableModes = useMemo(
    () => Array.from(new Set(providers.flatMap((provider) => provider.modes))).sort(),
    [providers],
  );
  const defaultMode = availableModes.includes("chat") ? "chat" : (availableModes[0] ?? "");
  const [section, setSection] = useState("endpoints");
  const [selectedMode, setSelectedMode] = useState(defaultMode);
  const [query, setQuery] = useState("");

  const normalizedQuery = query.trim().toLocaleLowerCase();
  const filteredProviders = providers.filter((provider) => {
    if (!normalizedQuery) return true;
    return `${provider.name} ${provider.slug} ${provider.description ?? ""}`
      .toLocaleLowerCase()
      .includes(normalizedQuery);
  });
  const endpointProviders = filteredProviders.filter((provider) =>
    provider.modes.includes(selectedMode),
  );
  const endpointCount = providers.reduce((sum, provider) => sum + provider.modes.length, 0);
  const modelCount = providers.reduce((sum, provider) => sum + (provider.modelCount ?? 0), 0);
  const byokCount = providers.filter((provider) => provider.supportsByok).length;

  if (hasError) {
    return (
      <Empty className="border border-warning-soft bg-warning-soft text-warning-soft-foreground">
        <EmptyHeader>
          <EmptyMedia variant="icon"><CircleAlert aria-hidden="true" /></EmptyMedia>
          <EmptyTitle>{t("loadError")}</EmptyTitle>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="flex min-w-0 flex-col gap-8">
      <section aria-label={t("metricsLabel")} className="grid overflow-hidden rounded-xl border border-border bg-card sm:grid-cols-2 lg:grid-cols-4">
        <Metric
          label={t("metrics.providers")}
          value={formatter.format(providers.length)}
          icon={Boxes}
          active={section === "providers"}
          onClick={() => setSection("providers")}
        />
        <Metric
          label={t("metrics.models")}
          value={formatter.format(modelCount)}
          icon={Boxes}
          href="/models"
        />
        <Metric
          label={t("metrics.endpointCombinations")}
          value={formatter.format(endpointCount)}
          icon={Cable}
          active={section === "endpoints"}
          onClick={() => setSection("endpoints")}
        />
        <Metric
          label={t("metrics.byokProviders")}
          value={formatter.format(byokCount)}
          icon={KeyRound}
        />
      </section>

      <Tabs value={section} onValueChange={setSection} className="gap-5">
        <TabsList variant="default" className="w-fit max-w-full overflow-x-auto">
          <TabsTrigger value="endpoints">{t("sections.endpoints")}</TabsTrigger>
          <TabsTrigger value="providers">
            {t("sections.providers")}
            <Badge>{formatter.format(providers.length)}</Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="endpoints">
          <div className="grid min-w-0 gap-6 lg:grid-cols-[15rem_minmax(0,1fr)]">
            <aside className="min-w-0">
              <div className="rounded-xl border border-border bg-card p-2 lg:sticky lg:top-20">
                {MODE_GROUPS.map((group) => {
                  const modes = group.modes.filter((mode) => availableModes.includes(mode));
                  if (!modes.length) return null;
                  return (
                    <div key={group.key} className="mb-3 last:mb-0">
                      <p className="px-2 py-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                        {t(`modeGroups.${group.key}`)}
                      </p>
                      <div className="flex gap-1 overflow-x-auto lg:flex-col">
                        {modes.map((mode) => {
                          const count = providers.filter((provider) => provider.modes.includes(mode)).length;
                          return (
                            <Button
                              key={mode}
                              type="button"
                              variant={selectedMode === mode ? "secondary" : "ghost"}
                              className="justify-between"
                              onClick={() => setSelectedMode(mode)}
                            >
                              <span className="truncate">{modeLabel(mode, t)}</span>
                              <span className="text-xs tabular-nums text-muted-foreground">{formatter.format(count)}</span>
                            </Button>
                          );
                        })}
                      </div>
                    </div>
                  );
                })}
              </div>
            </aside>

            <section className="flex min-w-0 flex-col gap-4">
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <h2 className="text-xl font-semibold">{modeLabel(selectedMode, t)}</h2>
                  <code className="rounded-md bg-muted px-2 py-1 text-xs" dir="ltr">{selectedMode}</code>
                </div>
                <p className="mt-1 text-sm leading-6 text-muted-foreground">
                  {t("endpointSupport", {
                    count: endpointProviders.length,
                    total: providers.length,
                  })}
                </p>
              </div>
              <SearchField value={query} onValueChange={setQuery} placeholder={t("filterProviders")} />
              {endpointProviders.length ? (
                <div className="grid gap-3 sm:grid-cols-2">
                  {endpointProviders.map((provider) => (
                    <ProviderEndpointCard key={provider.slug} provider={provider} selectedMode={selectedMode} />
                  ))}
                </div>
              ) : (
                <Empty className="border border-border bg-card">
                  <EmptyHeader>
                    <EmptyMedia variant="icon"><Cable aria-hidden="true" /></EmptyMedia>
                    <EmptyTitle>{t("noEndpointProviders")}</EmptyTitle>
                    <EmptyDescription>{t("noEndpointProvidersDescription")}</EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </section>
          </div>
        </TabsContent>

        <TabsContent value="providers">
          <section className="flex min-w-0 flex-col gap-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <SearchField value={query} onValueChange={setQuery} placeholder={t("searchProviders")} />
              <span className="text-sm text-muted-foreground">
                {t("providerResults", { count: filteredProviders.length, total: providers.length })}
              </span>
            </div>
            {filteredProviders.length ? (
              <div className="grid gap-3 sm:grid-cols-2">
                {filteredProviders.map((provider) => (
                  <ProviderDirectoryCard key={provider.slug} provider={provider} />
                ))}
              </div>
            ) : (
              <Empty className="border border-border bg-card">
                <EmptyHeader>
                  <EmptyMedia variant="icon"><Search aria-hidden="true" /></EmptyMedia>
                  <EmptyTitle>{t("noProviderMatches")}</EmptyTitle>
                </EmptyHeader>
              </Empty>
            )}
          </section>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function Metric({
  label,
  value,
  icon: Icon,
  href,
  active,
  onClick,
}: {
  label: string;
  value: string;
  icon: typeof Boxes;
  href?: string;
  active?: boolean;
  onClick?: () => void;
}) {
  const className = cn(
    "group flex min-h-28 items-center justify-between gap-4 border-b border-border p-5 text-start transition-colors sm:border-e lg:border-b-0",
    active ? "bg-primary-soft" : "hover:bg-muted/40",
  );
  const content = (
    <>
      <div>
        <strong className="block text-2xl font-bold tabular-nums">{value}</strong>
        <span className="mt-1 block text-sm text-muted-foreground">{label}</span>
      </div>
      <Icon className="size-5 text-muted-foreground transition-colors group-hover:text-primary" aria-hidden="true" />
    </>
  );

  if (href) return <Link href={href} className={className}>{content}</Link>;
  if (onClick) return <button type="button" className={className} onClick={onClick}>{content}</button>;
  return <div className={className}>{content}</div>;
}

function SearchField({
  value,
  onValueChange,
  placeholder,
}: {
  value: string;
  onValueChange: (value: string) => void;
  placeholder: string;
}) {
  return (
    <div className="relative w-full max-w-xl">
      <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
      <Input className="h-10 ps-10" value={value} onChange={(event) => onValueChange(event.target.value)} placeholder={placeholder} />
    </div>
  );
}

function ProviderEndpointCard({
  provider,
  selectedMode,
}: {
  provider: ProviderDirectoryItem;
  selectedMode: string;
}) {
  const t = useTranslations("Providers");
  return (
    <Link href={`/providers/${provider.slug}`} className="group rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
      <Card className="h-full p-0 transition-colors group-hover:border-primary/50">
        <CardContent className="flex items-center gap-3 p-4 sm:p-4">
          <ProviderMark name={provider.name} logoUrl={provider.logoUrl} />
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <strong className="truncate">{provider.name}</strong>
              <Badge className="border-success-soft bg-success-soft text-success-soft-foreground">{t("native")}</Badge>
            </div>
            <p className="mt-1 truncate text-xs text-muted-foreground" dir="ltr">{selectedMode}</p>
          </div>
          <ArrowLeft className="size-4 text-muted-foreground transition-transform group-hover:-translate-x-1 group-hover:text-primary" aria-hidden="true" />
        </CardContent>
      </Card>
    </Link>
  );
}

function ProviderDirectoryCard({ provider }: { provider: ProviderDirectoryItem }) {
  const t = useTranslations("Providers");
  const formatter = useMemo(() => new Intl.NumberFormat("fa-IR"), []);
  const visibleModes = provider.modes.slice(0, 7);
  const remaining = provider.modes.length - visibleModes.length;

  return (
    <Link href={`/providers/${provider.slug}`} className="group rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
      <Card className="h-full p-0 transition-colors group-hover:border-primary/50">
        <CardContent className="p-5 sm:p-5">
          <div className="flex items-start gap-3">
            <ProviderMark name={provider.name} logoUrl={provider.logoUrl} />
            <div className="min-w-0 flex-1">
              <strong className="block truncate text-base">{provider.name}</strong>
              <span className="text-xs text-muted-foreground">{t("activeModels", { count: provider.modelCount ?? 0 })}</span>
            </div>
            <ArrowLeft className="size-4 text-muted-foreground transition-transform group-hover:-translate-x-1 group-hover:text-primary" aria-hidden="true" />
          </div>
          <div className="mt-4 flex flex-wrap gap-1.5">
            {visibleModes.map((mode) => <Badge key={mode} className="bg-background">{modeLabel(mode, t)}</Badge>)}
            {remaining > 0 ? <Badge className="bg-background">+{formatter.format(remaining)}</Badge> : null}
          </div>
          <div className="mt-4 flex flex-wrap gap-3 text-xs text-muted-foreground">
            {provider.supportsByok ? <span className="inline-flex items-center gap-1"><KeyRound className="size-3.5" aria-hidden="true" />{t("byok")}</span> : null}
            {provider.supportsRegionalRouting ? <span className="inline-flex items-center gap-1"><Map className="size-3.5" aria-hidden="true" />{t("regionalRouting")}</span> : null}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

function modeLabel(
  mode: string,
  t: ReturnType<typeof useTranslations<"Providers">>,
) {
  const key = MODE_LABEL_KEYS[mode as keyof typeof MODE_LABEL_KEYS];
  return key ? t(key) : mode;
}
