"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Search } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Card, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

type ProviderModel = {
  id: string;
  slug: string;
  displayName: string;
};

const PAGE_SIZE = 24;

export function ProviderModelsList({
  providerSlug,
  models,
}: {
  providerSlug: string;
  models: ProviderModel[];
}) {
  const t = useTranslations("ProviderDetail");
  const [query, setQuery] = useState("");
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE);

  const filteredModels = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase();
    if (!normalizedQuery) return models;

    return models.filter((model) =>
      `${model.displayName} ${model.slug}`.toLocaleLowerCase().includes(normalizedQuery),
    );
  }, [models, query]);

  const visibleModels = filteredModels.slice(0, visibleCount);
  const remaining = Math.max(filteredModels.length - visibleModels.length, 0);

  return (
    <div className="flex flex-col gap-4">
      <div className="relative max-w-xl">
        <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <Input
          value={query}
          onChange={(event) => {
            setQuery(event.target.value);
            setVisibleCount(PAGE_SIZE);
          }}
          className="ps-10"
          placeholder={t("searchModelsPlaceholder")}
          aria-label={t("searchModelsLabel")}
        />
      </div>

      {visibleModels.length === 0 ? (
        <Card role="status">
          <p className="text-sm text-muted-foreground">{t("noModelMatches")}</p>
        </Card>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          {visibleModels.map((model) => (
            <Card key={model.id} className="group overflow-hidden p-0 transition hover:-translate-y-0.5 hover:border-primary hover:shadow-md">
              <Link
                href={`/models/${providerSlug}/${model.slug}`}
                className="flex h-full flex-col focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                <CardHeader className="flex-1">
                  <CardTitle className="text-sm leading-6 transition-colors group-hover:text-primary">
                    <bdi dir="auto">{model.displayName}</bdi>
                  </CardTitle>
                  <p className="truncate text-xs text-muted-foreground" dir="ltr" data-technical>
                    /{model.slug}
                  </p>
                </CardHeader>
                <CardFooter className="justify-between border-t border-border bg-muted/40">
                  <span className="text-xs font-medium text-primary">{t("viewModelDetails")}</span>
                  <ArrowLeft className="size-4 text-primary transition-transform group-hover:-translate-x-1" aria-hidden="true" />
                </CardFooter>
              </Link>
            </Card>
          ))}
        </div>
      )}

      {remaining > 0 ? (
        <Button
          type="button"
          variant="outline"
          className="self-center"
          onClick={() => setVisibleCount((count) => count + PAGE_SIZE)}
        >
          {t("showMoreModels", { count: remaining })}
        </Button>
      ) : null}
    </div>
  );
}
