"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { Check, Copy, ExternalLink, Search } from "lucide-react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

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
  docsUrl,
}: {
  providerSlug: string;
  models: ProviderModel[];
  docsUrl?: string;
}) {
  const t = useTranslations("ProviderDetail");
  const [query, setQuery] = useState("");
  const [copied, setCopied] = useState("");

  const filteredModels = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase();
    if (!normalizedQuery) return models;
    return models.filter((model) =>
      `${model.displayName} ${model.canonicalId ?? ""} ${model.slug}`
        .toLocaleLowerCase()
        .includes(normalizedQuery),
    );
  }, [models, query]);

  const modes = useMemo(
    () => Array.from(new Set(models.flatMap((model) => model.mode ?? []).filter((mode) => mode !== "unknown"))).sort(),
    [models],
  );
  const sampleModel = models.find((model) => model.mode?.includes("chat")) ?? models[0];
  const sampleId = sampleModel?.canonicalId ?? `${providerSlug}/${sampleModel?.slug ?? "model-id"}`;

  async function copy(value: string) {
    await navigator.clipboard.writeText(value);
    setCopied(value);
    window.setTimeout(() => setCopied(""), 1400);
  }

  return (
    <Tabs defaultValue="models" className="gap-5">
      <TabsList className="w-fit max-w-full overflow-x-auto">
        <TabsTrigger value="models">
          {t("tabs.models")}
          <Badge>{new Intl.NumberFormat("fa-IR").format(models.length)}</Badge>
        </TabsTrigger>
        <TabsTrigger value="coverage">{t("tabs.coverage")}</TabsTrigger>
      </TabsList>

      <TabsContent value="models" className="space-y-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative w-full max-w-xl">
            <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
            <Input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              className="h-10 ps-10"
              placeholder={t("searchModelsPlaceholder")}
              aria-label={t("searchModelsLabel")}
            />
          </div>
          <span className="text-sm text-muted-foreground">{t("modelsShown", { count: filteredModels.length })}</span>
        </div>

        {filteredModels.length ? (
          <Card className="overflow-hidden p-0">
            <Table dir="rtl">
              <TableHeader className="bg-muted/50">
                <TableRow>
                  <TableHead>{t("table.model")}</TableHead>
                  <TableHead>{t("table.modelId")}</TableHead>
                  <TableHead>{t("table.input")}</TableHead>
                  <TableHead>{t("table.output")}</TableHead>
                  <TableHead>{t("table.cacheRead")}</TableHead>
                  <TableHead>{t("table.modes")}</TableHead>
                  <TableHead>{t("table.context")}</TableHead>
                  <TableHead><span className="sr-only">{t("table.actions")}</span></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredModels.map((model) => {
                  const price = model.pricing?.[0];
                  const modelId = model.canonicalId ?? `${providerSlug}/${model.slug}`;
                  return (
                    <TableRow key={model.id}>
                      <TableCell>
                        <Link href={`/models/${providerSlug}/${model.slug}`} className="font-medium hover:text-primary">
                          {model.displayName}
                        </Link>
                      </TableCell>
                      <TableCell>
                        <code className="text-xs" dir="ltr">{modelId}</code>
                      </TableCell>
                      <TableCell dir="ltr">{formatPrice(price?.inputPricePerMillionTokens)}</TableCell>
                      <TableCell dir="ltr">{formatPrice(price?.outputPricePerMillionTokens)}</TableCell>
                      <TableCell dir="ltr">{formatPrice(price?.cachedInputPricePerMillionTokens)}</TableCell>
                      <TableCell>
                        <div className="flex max-w-64 flex-wrap gap-1">
                          {(model.mode ?? []).slice(0, 3).map((mode) => <Badge key={mode} className="bg-background">{mode}</Badge>)}
                        </div>
                      </TableCell>
                      <TableCell dir="ltr">{model.contextWindow ? model.contextWindow.toLocaleString("en-US") : "—"}</TableCell>
                      <TableCell>
                        <Button size="icon-sm" variant="ghost" onClick={() => copy(modelId)} aria-label={t("copyModelId")}>
                          {copied === modelId ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </Card>
        ) : (
          <Card><p className="text-sm text-muted-foreground">{t("noModelMatches")}</p></Card>
        )}
      </TabsContent>

      <TabsContent value="coverage" className="space-y-5">
        <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_20rem]">
          <Card className="overflow-hidden p-0">
            <div className="flex items-center justify-between border-b border-border px-5 py-4">
              <div>
                <h2 className="font-semibold">{t("quickStart.title")}</h2>
                <p className="mt-1 text-sm text-muted-foreground">{t("quickStart.description")}</p>
              </div>
              <Button size="icon-sm" variant="ghost" onClick={() => copy(quickStartCode(sampleId))} aria-label={t("quickStart.copy")}>
                {copied === quickStartCode(sampleId) ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
              </Button>
            </div>
            <pre className="overflow-x-auto bg-muted/40 p-5 text-start text-xs leading-6" dir="ltr">
              <code>{quickStartCode(sampleId)}</code>
            </pre>
          </Card>
          <Card>
            <h2 className="font-semibold">{t("coverageSummary.title")}</h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              {t("coverageSummary.description", { count: modes.length })}
            </p>
            {docsUrl ? (
              <a className="mt-4 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline" href={docsUrl} target="_blank" rel="noreferrer">
                {t("viewApiDocs")} <ExternalLink className="size-4" aria-hidden="true" />
              </a>
            ) : null}
          </Card>
        </div>
        <section>
          <h2 className="text-lg font-semibold">{t("supportedEndpoints")}</h2>
          <p className="mt-1 text-sm text-muted-foreground">{t("supportedEndpointsDescription")}</p>
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {modes.map((mode) => (
              <Card key={mode} className="flex items-center justify-between gap-3">
                <div>
                  <strong className="block text-sm">{mode}</strong>
                  <code className="mt-1 block text-xs text-muted-foreground" dir="ltr">{mode}</code>
                </div>
                <Badge className="border-success-soft bg-success-soft text-success-soft-foreground">{t("supported")}</Badge>
              </Card>
            ))}
          </div>
        </section>
      </TabsContent>
    </Tabs>
  );
}

function formatPrice(value?: number | null) {
  return typeof value === "number" ? `$${value.toLocaleString("en-US", { maximumFractionDigits: 6 })}` : "—";
}

function quickStartCode(modelId: string) {
  return `curl https://api.godarai.ir/v1/chat/completions \\
  -H "Authorization: Bearer $GODARAI_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"${modelId}","messages":[{"role":"user","content":"Hello"}]}'`;
}
