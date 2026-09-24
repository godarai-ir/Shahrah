"use client";

import { useId } from "react";
import { useTranslations } from "next-intl";
import { Check } from "lucide-react";

import { EmptyState } from "@/components/app-ui/EmptyState";
import { SearchField } from "@/components/app-ui/SearchField";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { GUARDRAIL_CATALOG, GUARDRAIL_CATEGORIES, type GuardrailCatalogEntry } from "../catalog";
import type { GuardrailType } from "../types";
import type { TypeCategoryFilter } from "./state";

type StepTypeProps = {
  selected: GuardrailType | null;
  query: string;
  category: TypeCategoryFilter;
  headingRef?: React.Ref<HTMLHeadingElement>;
  onQueryChange: (query: string) => void;
  onCategoryChange: (category: TypeCategoryFilter) => void;
  onSelect: (type: GuardrailType) => void;
};

function TypeTile({ entry, selected, onSelect }: { entry: GuardrailCatalogEntry; selected: boolean; onSelect: () => void }) {
  const t = useTranslations("guardrails");
  const id = useId();
  const Icon = entry.icon;

  return (
    <button
      type="button"
      aria-pressed={selected}
      aria-labelledby={`${id}-name`}
      aria-describedby={`${id}-description`}
      onClick={onSelect}
      className="group flex min-h-42 flex-col gap-3 rounded-xl border border-card-border bg-card p-4 text-start shadow-xs transition-colors hover:border-primary-subtle-border focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-pressed:border-primary aria-pressed:bg-primary-soft aria-pressed:ring-4 aria-pressed:ring-primary/15"
    >
      <span className="flex items-center gap-2">
        <span
          aria-hidden="true"
          className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted text-primary group-aria-pressed:bg-primary group-aria-pressed:text-primary-foreground [&_svg]:size-5"
        >
          <Icon />
        </span>
        {entry.recommended ? (
          <Badge variant="outline" className="border-primary-subtle-border bg-primary-soft text-primary">{t("wizard.type.recommended")}</Badge>
        ) : null}
        <span
          aria-hidden="true"
          className="ms-auto flex size-5 shrink-0 items-center justify-center rounded-full border border-input-border text-primary-foreground group-aria-pressed:border-primary group-aria-pressed:bg-primary"
        >
          {selected ? <Check className="size-3" strokeWidth={3} /> : null}
        </span>
      </span>
      <span className="flex flex-col gap-1">
        <span id={`${id}-name`} className="text-sm font-semibold text-foreground">{t(`types.${entry.type}`)}</span>
        <span id={`${id}-description`} className="text-xs leading-5 text-muted-foreground">{t(`typeDescriptions.${entry.type}`)}</span>
      </span>
      <span className="mt-auto flex">
        <Badge variant="outline" className="font-normal text-muted-foreground">{t(`wizard.type.ops.${entry.ops}`)}</Badge>
      </span>
    </button>
  );
}

export function StepType({ selected, query, category, headingRef, onQueryChange, onCategoryChange, onSelect }: StepTypeProps) {
  const t = useTranslations("guardrails");
  const normalizedQuery = query.trim().toLowerCase();
  const matches = (entry: GuardrailCatalogEntry) => {
    if (category !== "all" && entry.category !== category) return false;
    if (!normalizedQuery) return true;

    return [t(`types.${entry.type}`), t(`typeDescriptions.${entry.type}`), entry.type]
      .some((text) => text.toLowerCase().includes(normalizedQuery));
  };
  const groups = GUARDRAIL_CATEGORIES
    .map((key) => ({ key, entries: GUARDRAIL_CATALOG.filter((entry) => entry.category === key && matches(entry)) }))
    .filter((group) => group.entries.length > 0);
  const filters: TypeCategoryFilter[] = ["all", ...GUARDRAIL_CATEGORIES];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h2 ref={headingRef} tabIndex={-1} className="text-base font-semibold outline-none md:text-lg">{t("wizard.type.heading")}</h2>
        <p className="text-sm leading-6 text-muted-foreground">{t("wizard.type.note")}</p>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <SearchField
          containerClassName="w-full sm:max-w-xs"
          placeholder={t("wizard.type.search")}
          aria-label={t("wizard.type.search")}
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
        />
        <div role="group" aria-label={t("wizard.type.categoriesLabel")} className="flex flex-wrap gap-2">
          {filters.map((filter) => (
            <Button
              key={filter}
              type="button"
              size="sm"
              variant="outline"
              aria-pressed={category === filter}
              onClick={() => onCategoryChange(filter)}
              className="rounded-full font-normal aria-pressed:border-primary aria-pressed:bg-primary-soft aria-pressed:font-semibold aria-pressed:text-primary"
            >
              {t(`categories.${filter}`)}
            </Button>
          ))}
        </div>
      </div>

      {groups.map((group) => (
        <section key={group.key} className="flex flex-col gap-3" aria-labelledby={`guardrail-category-${group.key}`}>
          <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
            <h3 id={`guardrail-category-${group.key}`} className="text-sm font-semibold text-foreground">{t(`categories.${group.key}`)}</h3>
            <p className="text-xs text-muted-foreground">{t(`categoryNotes.${group.key}`)}</p>
          </div>
          <div className="grid grid-cols-[repeat(auto-fill,minmax(13.75rem,1fr))] gap-3">
            {group.entries.map((entry) => (
              <TypeTile key={entry.type} entry={entry} selected={selected === entry.type} onSelect={() => onSelect(entry.type)} />
            ))}
          </div>
        </section>
      ))}

      {groups.length === 0 ? <EmptyState variant="surface" density="compact" title={t("wizard.type.empty")} /> : null}
    </div>
  );
}
