"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
  Check,
  ChevronLeft,
  ChevronRight,
  Copy,
  ExternalLink,
  Search,
  SearchX,
  X,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { copyTextToClipboard } from "@/lib/clipboard";
import { useDisplayFormatters } from "@/lib/use-display-formatters";
import { cn } from "@/lib/utils";
import {
  EndpointModeBadge,
  EndpointModeBadgeList,
} from "@/components/catalog/endpoint-mode-badge";
import { TokenScale } from "@/components/catalog/token-scale";
import { CompareToggleButton } from "@/components/compare/compare-toggle-button";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const ALL_MODES = "all";
const PAGE_SIZE = 24;

export type ProviderModel = {
  id: string;
  canonicalId?: string;
  slug: string;
  displayName: string;
  mode?: string[];
  contextWindow?: number | null;
  updatedAt?: string | null;
  pricing?: Array<{
    inputPricePerMillionTokens?: number | null;
    outputPricePerMillionTokens?: number | null;
    cachedInputPricePerMillionTokens?: number | null;
  }>;
  capabilities?: Array<{ capability?: { slug?: string; name?: string } }>;
};

export function ProviderModelsList({
  providerSlug,
  models,
}: {
  providerSlug: string;
  models: ProviderModel[];
}) {
  const t = useTranslations("ProviderDetail");
  const { settings, formatPriceAmount } = useDisplayFormatters();
  const usesPersianCurrency = settings.currencyUnit !== "USD";
  const [query, setQuery] = useState("");
  const [selectedMode, setSelectedMode] = useState(ALL_MODES);
  const [selectedEndpoint, setSelectedEndpoint] = useState("");
  const [page, setPage] = useState(1);
  const [copied, setCopied] = useState("");
  const [announcement, setAnnouncement] = useState("");

  const modes = useMemo(
    () =>
      Array.from(
        new Set(
          models
            .flatMap((model) => model.mode ?? [])
            .filter((mode) => mode !== "unknown"),
        ),
      ).sort(),
    [models],
  );

  const filteredModels = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase();

    return models.filter((model) => {
      const matchesQuery =
        !normalizedQuery ||
        `${model.displayName} ${model.canonicalId ?? ""} ${model.slug}`
          .toLocaleLowerCase()
          .includes(normalizedQuery);
      const matchesMode =
        selectedMode === ALL_MODES || model.mode?.includes(selectedMode);

      return matchesQuery && matchesMode;
    });
  }, [models, query, selectedMode]);

  const activeEndpoint = modes.includes(selectedEndpoint)
    ? selectedEndpoint
    : modes.includes("chat")
      ? "chat"
      : (modes[0] ?? "");
  const pageCount = Math.max(1, Math.ceil(filteredModels.length / PAGE_SIZE));
  const currentPage = Math.min(page, pageCount);
  const pageStart = (currentPage - 1) * PAGE_SIZE;
  const visibleModels = filteredModels.slice(pageStart, pageStart + PAGE_SIZE);
  const visibleStart = filteredModels.length ? pageStart + 1 : 0;
  const visibleEnd = Math.min(pageStart + PAGE_SIZE, filteredModels.length);
  const sampleModel =
    models.find((model) => model.mode?.includes(activeEndpoint)) ?? models[0];
  const sampleId =
    sampleModel?.canonicalId ??
    `${providerSlug}/${sampleModel?.slug ?? "model-id"}`;
  const endpointExample = getEndpointExample(activeEndpoint, sampleId);
  const filtersActive =
    query.trim().length > 0 || selectedMode !== ALL_MODES;

  function updateQuery(value: string) {
    setQuery(value);
    setPage(1);
  }

  function updateMode(value: string) {
    setSelectedMode(value);
    setPage(1);
  }

  function clearFilters() {
    setQuery("");
    setSelectedMode(ALL_MODES);
    setPage(1);
  }

  async function copy(value: string) {
    try {
      await copyTextToClipboard(value);
      setCopied(value);
      setAnnouncement(t("copySuccess"));
      window.setTimeout(() => {
        setCopied("");
        setAnnouncement("");
      }, 1800);
    } catch {
      setAnnouncement(t("copyFailed"));
    }
  }

  return (
    <Tabs dir="rtl" defaultValue="models" className="gap-5">
      <p className="sr-only" aria-live="polite" aria-atomic="true">
        {announcement}
      </p>

      <TabsList className="scrollbar-none mx-auto w-fit max-w-full overflow-x-auto p-1 group-data-horizontal/tabs:h-12">
        <TabsTrigger value="models" className="cursor-pointer px-4 text-base font-semibold sm:px-6">
          <bdi>{t("tabs.models")}</bdi>
        </TabsTrigger>
        <TabsTrigger value="coverage" className="cursor-pointer px-4 text-base font-semibold sm:px-6">
          <bdi>{t("tabs.coverage")}</bdi>
        </TabsTrigger>
      </TabsList>

      <TabsContent value="models" className="flex flex-col gap-4">
        <Card className="p-0">
          <CardContent className="flex flex-col gap-3 p-3 sm:p-4">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
              <div className="flex min-w-0 flex-1 gap-2">
                <div className="relative min-w-0 flex-1">
                  <Search
                    className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
                    aria-hidden="true"
                  />
                  <Input
                    value={query}
                    onChange={(event) => updateQuery(event.target.value)}
                    className="h-11 ps-10 sm:h-10"
                    placeholder={t("searchModelsPlaceholder")}
                    aria-label={t("searchModelsLabel")}
                  />
                </div>
                {query ? (
                  <Button
                    type="button"
                    size="icon"
                    variant="ghost"
                    className="size-11 sm:size-10"
                    onClick={() => updateQuery("")}
                    aria-label={t("clearSearch")}
                    title={t("clearSearch")}
                  >
                    <X aria-hidden="true" />
                  </Button>
                ) : null}
              </div>

              <div className="w-full lg:w-64">
                <Label className="sr-only" htmlFor="provider-mode-filter">
                  {t("modeFilterLabel")}
                </Label>
                <Select value={selectedMode} onValueChange={updateMode}>
                  <SelectTrigger
                    id="provider-mode-filter"
                    className="h-11 sm:h-10"
                    aria-label={t("modeFilterLabel")}
                  >
                    <SelectValue placeholder={t("allModes")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value={ALL_MODES}>{t("allModes")}</SelectItem>
                      {modes.map((mode) => (
                        <SelectItem key={mode} value={mode}>
                          <bdi dir="ltr">{mode}</bdi>
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {filtersActive ? (
              <div className="flex justify-end">
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={clearFilters}
                >
                  <X data-icon="inline-start" aria-hidden="true" />
                  {t("clearFilters")}
                </Button>
              </div>
            ) : null}
          </CardContent>
        </Card>

        {filteredModels.length ? (
          <>
            <div className="flex flex-col gap-3 md:hidden">
              {visibleModels.map((model) => {
                const price = model.pricing?.[0];
                const modelId = getModelId(providerSlug, model);

                return (
                  <Card key={model.id} className="p-0">
                    <CardHeader className="p-4">
                      <div className="flex min-w-0 items-start gap-3">
                        <div className="min-w-0 flex-1">
                          <div className="flex min-w-0 items-center gap-2">
                            <CompareToggleButton canonicalId={modelId} iconOnly />
                            <CardTitle className="min-w-0 text-base leading-6">
                              <Link
                                href={`/models/${providerSlug}/${model.slug}`}
                                className="rounded-sm hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                              >
                                <bdi dir="auto">{model.displayName}</bdi>
                              </Link>
                            </CardTitle>
                          </div>
                          <CardDescription
                            className="mt-2 overflow-x-auto text-left font-mono text-xs whitespace-nowrap"
                            dir="ltr"
                            title={modelId}
                          >
                            {modelId}
                          </CardDescription>
                        </div>
                        <Button
                          type="button"
                          size="icon"
                          variant="ghost"
                          className="size-10 shrink-0"
                          onClick={() => copy(modelId)}
                          aria-label={t("copyModelId")}
                          title={t("copyModelId")}
                        >
                          {copied === modelId ? (
                            <Check aria-hidden="true" />
                          ) : (
                            <Copy aria-hidden="true" />
                          )}
                        </Button>
                      </div>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-4 p-4 pt-0">
                      <dl className="grid grid-cols-3 gap-2">
                        <MobileMetric
                          label={t("table.inputShort")}
                          value={formatPriceAmount(
                            price?.inputPricePerMillionTokens,
                          )}
                          usesPersianCurrency={usesPersianCurrency}
                        />
                        <MobileMetric
                          label={t("table.outputShort")}
                          value={formatPriceAmount(
                            price?.outputPricePerMillionTokens,
                          )}
                          usesPersianCurrency={usesPersianCurrency}
                        />
                        <MobileMetric
                          label={t("table.cacheShort")}
                          value={formatPriceAmount(
                            price?.cachedInputPricePerMillionTokens,
                          )}
                          usesPersianCurrency={usesPersianCurrency}
                        />
                      </dl>

                      <Separator />

                      <dl className="flex flex-col gap-3">
                        <div className="flex items-center justify-between gap-3">
                          <dt className="text-xs text-muted-foreground">
                            {t("table.context")}
                          </dt>
                          <dd
                            className="font-mono text-sm tabular-nums"
                            dir="ltr"
                          >
                            {formatContextWindow(model.contextWindow)}
                          </dd>
                        </div>
                        <div className="flex items-start justify-between gap-3">
                          <dt className="shrink-0 text-xs text-muted-foreground">
                            {t("table.modes")}
                          </dt>
                          <dd className="flex flex-wrap justify-end gap-1">
                            <EndpointModeBadgeList
                              modes={model.mode ?? []}
                              limit={3}
                              className="justify-end"
                            />
                          </dd>
                        </div>
                      </dl>
                    </CardContent>
                  </Card>
                );
              })}
            </div>

            <Card className="hidden overflow-hidden p-0 md:block [&_td]:px-4 [&_th]:px-4 sm:[&_td]:px-5 sm:[&_th]:px-5">
              <Table dir="rtl">
                <TableHeader className="bg-muted/50">
                  <TableRow>
                    <TableHead className="min-w-48">
                      {t("table.model")}
                    </TableHead>
                    <TableHead>
                      <span className="inline-flex items-center gap-1">
                        {t("table.inputShort")}
                        <span aria-hidden="true">(</span>
                        <TokenScale scale={settings.priceScale} />
                        <span aria-hidden="true">)</span>
                      </span>
                    </TableHead>
                    <TableHead>
                      <span className="inline-flex items-center gap-1">
                        {t("table.outputShort")}
                        <span aria-hidden="true">(</span>
                        <TokenScale scale={settings.priceScale} />
                        <span aria-hidden="true">)</span>
                      </span>
                    </TableHead>
                    <TableHead>
                      <span className="inline-flex items-center gap-1">
                        {t("table.cacheShort")}
                        <span aria-hidden="true">(</span>
                        <TokenScale scale={settings.priceScale} />
                        <span aria-hidden="true">)</span>
                      </span>
                    </TableHead>
                    <TableHead className="min-w-44">
                      {t("table.modes")}
                    </TableHead>
                    <TableHead>{t("table.context")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visibleModels.map((model) => {
                    const price = model.pricing?.[0];
                    const modelId = getModelId(providerSlug, model);

                    return (
                      <TableRow key={model.id} className="group">
                        <TableCell>
                          <div className="flex min-w-0 items-center gap-2">
                            <CompareToggleButton canonicalId={modelId} iconOnly />
                            <Link
                              href={`/models/${providerSlug}/${model.slug}`}
                              className="min-w-0 rounded-sm font-medium hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            >
                              <bdi dir="auto">{model.displayName}</bdi>
                            </Link>
                            <Button
                              type="button"
                              size="icon-sm"
                              variant="ghost"
                              className="shrink-0 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"
                              onClick={() => copy(modelId)}
                              aria-label={t("copyModelId")}
                              title={t("copyModelId")}
                            >
                              {copied === modelId ? (
                                <Check aria-hidden="true" />
                              ) : (
                                <Copy aria-hidden="true" />
                              )}
                            </Button>
                          </div>
                        </TableCell>
                        <TableCell
                          className={cn(
                            "tabular-nums",
                            usesPersianCurrency ? "font-sans" : "font-mono",
                          )}
                          dir={usesPersianCurrency ? "rtl" : "ltr"}
                        >
                          {formatPriceAmount(price?.inputPricePerMillionTokens)}
                        </TableCell>
                        <TableCell
                          className={cn(
                            "tabular-nums",
                            usesPersianCurrency ? "font-sans" : "font-mono",
                          )}
                          dir={usesPersianCurrency ? "rtl" : "ltr"}
                        >
                          {formatPriceAmount(price?.outputPricePerMillionTokens)}
                        </TableCell>
                        <TableCell
                          className={cn(
                            "tabular-nums",
                            usesPersianCurrency ? "font-sans" : "font-mono",
                          )}
                          dir={usesPersianCurrency ? "rtl" : "ltr"}
                        >
                          {formatPriceAmount(
                            price?.cachedInputPricePerMillionTokens,
                          )}
                        </TableCell>
                        <TableCell>
                          <EndpointModeBadgeList
                            modes={model.mode ?? []}
                            limit={3}
                            className="max-w-64"
                          />
                        </TableCell>
                        <TableCell
                          className="font-mono tabular-nums"
                          dir="ltr"
                        >
                          {formatContextWindow(model.contextWindow)}
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </Card>

            {pageCount > 1 ? (
              <nav
                className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
                aria-label={t("paginationLabel")}
              >
                <p className="text-sm text-muted-foreground">
                  {t("paginationSummary", {
                    start: visibleStart,
                    end: visibleEnd,
                    total: filteredModels.length,
                  })}
                </p>
                <div className="grid grid-cols-2 gap-2 sm:flex">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="h-10 sm:h-8"
                    disabled={currentPage === 1}
                    onClick={() =>
                      setPage((current) => Math.max(1, current - 1))
                    }
                  >
                    <ChevronRight data-icon="inline-start" aria-hidden="true" />
                    {t("previousPage")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="h-10 sm:h-8"
                    disabled={currentPage === pageCount}
                    onClick={() =>
                      setPage((current) => Math.min(pageCount, current + 1))
                    }
                  >
                    {t("nextPage")}
                    <ChevronLeft data-icon="inline-end" aria-hidden="true" />
                  </Button>
                </div>
              </nav>
            ) : null}
          </>
        ) : (
          <Card className="p-0">
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <SearchX aria-hidden="true" />
                </EmptyMedia>
                <EmptyTitle>{t("noModelMatchesTitle")}</EmptyTitle>
                <EmptyDescription>{t("noModelMatches")}</EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button type="button" variant="outline" onClick={clearFilters}>
                  {t("clearFilters")}
                </Button>
              </EmptyContent>
            </Empty>
          </Card>
        )}
      </TabsContent>

      <TabsContent value="coverage">
        <section className="grid min-w-0 gap-5 lg:grid-cols-3" dir="rtl">
          <Card className="min-w-0 p-0">
            <CardHeader className="p-5">
              <CardTitle>{t("supportedEndpoints")}</CardTitle>
              <CardDescription className="mt-2">
                {t("supportedEndpointsDescription")}
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-2 p-3 pt-0 sm:p-4 sm:pt-0">
              {modes.map((mode) => {
                const selected = mode === activeEndpoint;

                return (
                  <Button
                    key={mode}
                    type="button"
                    variant="outline"
                    className={cn(
                      "h-auto min-h-12 w-full cursor-pointer justify-between gap-3 px-3 py-3",
                      selected &&
                        "border-primary bg-primary-soft text-primary-soft-foreground hover:bg-primary-soft",
                    )}
                    aria-pressed={selected}
                    onClick={() => setSelectedEndpoint(mode)}
                  >
                    <EndpointModeBadge mode={mode} />
                    <Badge className="shrink-0 border-success-soft bg-success-soft text-success-soft-foreground">
                      {t("supported")}
                    </Badge>
                  </Button>
                );
              })}
            </CardContent>
          </Card>

          <Card className="min-w-0 overflow-hidden p-0 lg:col-span-2">
            <CardHeader className="flex-row items-start justify-between gap-4 p-5">
              <div className="min-w-0">
                <CardTitle>
                  {t("endpointExample.title", { endpoint: activeEndpoint })}
                </CardTitle>
                <CardDescription className="mt-1">
                  {t("endpointExample.description")}
                </CardDescription>
              </div>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="size-10 shrink-0"
                onClick={() => copy(endpointExample.code)}
                aria-label={t("quickStart.copy")}
                title={t("quickStart.copy")}
              >
                {copied === endpointExample.code ? (
                  <Check aria-hidden="true" />
                ) : (
                  <Copy aria-hidden="true" />
                )}
              </Button>
            </CardHeader>
            <CardContent
              className="min-h-64 overflow-x-auto border-y border-border bg-muted/40 p-5 text-left font-mono text-xs leading-6 whitespace-pre"
              dir="ltr"
            >
              {endpointExample.code}
            </CardContent>
            <CardContent className="flex flex-col gap-2 p-5">
              <p className="text-sm text-muted-foreground">
                {t("endpointExample.docsDescription")}
              </p>
              <Button asChild variant="link" className="h-auto w-fit px-0">
                <a
                  href={endpointExample.docsHref}
                  target="_blank"
                  rel="noreferrer"
                >
                  {t("endpointExample.docsLink", {
                    endpoint: activeEndpoint,
                  })}
                  <ExternalLink data-icon="inline-end" aria-hidden="true" />
                </a>
              </Button>
            </CardContent>
          </Card>
        </section>
      </TabsContent>
    </Tabs>
  );
}

function MobileMetric({
  label,
  value,
  usesPersianCurrency,
}: {
  label: string;
  value: string;
  usesPersianCurrency: boolean;
}) {
  return (
    <div className="min-w-0 rounded-lg bg-muted p-2 text-center">
      <dt className="truncate text-xs text-muted-foreground">{label}</dt>
      <dd
        className={cn(
          "mt-1 truncate text-sm tabular-nums",
          usesPersianCurrency ? "font-sans" : "font-mono",
        )}
        dir={usesPersianCurrency ? "rtl" : "ltr"}
      >
        {value}
      </dd>
    </div>
  );
}

function getModelId(providerSlug: string, model: ProviderModel) {
  return model.canonicalId ?? `${providerSlug}/${model.slug}`;
}

function formatContextWindow(value?: number | null) {
  return value ? value.toLocaleString("en-US") : "—";
}

const WORKSPACE_BASE_URL = "https://YOUR-WORKSPACE-NAME.godarai.ir";
const DOCS_BASE_URL = "https://godarai.ir/docs/supported-apis";

function getEndpointExample(mode: string, modelId: string) {
  const normalized = mode.toLowerCase();

  if (normalized.includes("transcription")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/audio/transcriptions`,
      `${DOCS_BASE_URL}/speech-to-text`,
      `  -F "file=@audio.mp3" \\
  -F "model=${modelId}"`,
    );
  }
  if (normalized.includes("translation")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/audio/translations`,
      `${DOCS_BASE_URL}/audio-translation`,
      `  -F "file=@audio.mp3" \\
  -F "model=${modelId}"`,
    );
  }
  if (normalized.includes("text_to_speech") || normalized.includes("speech")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/audio/speech`,
      `${DOCS_BASE_URL}/text-to-speech`,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","input":"سلام دنیا","voice":"alloy"}' \\
  --output speech.mp3`,
    );
  }
  if (normalized.includes("embedding")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/embeddings`,
      `${DOCS_BASE_URL}/embeddings`,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","input":"متن نمونه برای بردارسازی"}'`,
    );
  }
  if (normalized.includes("image")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/images/generations`,
      `${DOCS_BASE_URL}/image-generation`,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","prompt":"یک شهر آینده‌نگر در غروب"}'`,
    );
  }
  if (normalized.includes("responses")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/responses`,
      `${DOCS_BASE_URL}/responses-api`,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","input":"سلام؛ خودت را معرفی کن."}'`,
    );
  }
  if (normalized.includes("moderation")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/moderations`,
      `${DOCS_BASE_URL}/moderation`,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","input":"متن مورد بررسی"}'`,
    );
  }
  if (normalized.includes("rerank")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/rerank`,
      `${DOCS_BASE_URL}/rerank`,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","query":"پرسش نمونه","documents":["سند اول","سند دوم"]}'`,
    );
  }
  if (normalized.includes("completion") && !normalized.includes("chat")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/completions`,
      `${DOCS_BASE_URL}/chat-completions/overview`,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","prompt":"روزی روزگاری"}'`,
    );
  }
  if (normalized.includes("realtime")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/realtime/sessions`,
      DOCS_BASE_URL,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","modalities":["text","audio"]}'`,
    );
  }
  if (normalized.includes("video")) {
    return createEndpointExample(
      `${WORKSPACE_BASE_URL}/v1/videos/generations`,
      DOCS_BASE_URL,
      `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","prompt":"یک نمای سینمایی از شهر در شب"}'`,
    );
  }

  return createEndpointExample(
    `${WORKSPACE_BASE_URL}/v1/chat/completions`,
    `${DOCS_BASE_URL}/chat-completions/overview`,
    `  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","messages":[{"role":"user","content":"سلام"}]}'`,
  );
}

function createEndpointExample(url: string, docsHref: string, body: string) {
  return {
    docsHref,
    code: `curl --request POST "${url}" \\
  -H "Authorization: Bearer $GODARAI_API_KEY" \\
${body}`,
  };
}
