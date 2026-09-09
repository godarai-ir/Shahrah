"use client";

import { Check, ChevronDown, SearchX } from "lucide-react";
import { useMemo, useState } from "react";

import { EmptyState } from "@/components/app-ui/EmptyState";
import { ProviderLogo } from "@/components/app-ui/ProviderLogo";
import { SearchPickerDialog } from "@/components/app-ui/SearchPickerDialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

export type ProviderModelOption = {
  id: string;
  accountID: string;
  accountName: string;
  provider: string;
  providerDisplayName: string;
  logoURL?: string;
  model: string;
  fqn: string;
  modes: string[];
};

type ProviderModelPickerLabels = {
  title: string;
  placeholder: string;
  searchPlaceholder: string;
  groupCount: (count: number) => string;
  selected: string;
  emptyTitle: string;
  emptyDescription: string;
};

type ProviderModelPickerProps = {
  id: string;
  options: ProviderModelOption[];
  selected: ProviderModelOption | null;
  onSelectedChange: (value: ProviderModelOption) => void;
  labels: ProviderModelPickerLabels;
};

type ProviderModelGroup = {
  key: string;
  provider: string;
  providerDisplayName: string;
  logoURL?: string;
  accountName: string;
  options: ProviderModelOption[];
};

function shortModelName(option: ProviderModelOption): string {
  const providerPrefix = `${option.provider}/`;
  return option.model.startsWith(providerPrefix) ? option.model.slice(providerPrefix.length) : option.model;
}

export function ProviderModelPicker({ id, options, selected, onSelectedChange, labels }: ProviderModelPickerProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");

  const visibleGroups = useMemo(() => {
    const normalizedSearch = search.trim().toLocaleLowerCase();
    const visibleOptions = normalizedSearch
      ? options.filter((option) =>
          [shortModelName(option), option.model, option.fqn, option.provider, option.accountName]
            .join(" ")
            .toLocaleLowerCase()
            .includes(normalizedSearch),
        )
      : options;

    const groups = new Map<string, ProviderModelGroup>();
    for (const option of visibleOptions) {
      const key = `${option.provider}:${option.accountID}`;
      const group = groups.get(key) ?? {
        key,
        provider: option.provider,
        providerDisplayName: option.providerDisplayName,
        logoURL: option.logoURL,
        accountName: option.accountName,
        options: [],
      };
      group.options.push(option);
      groups.set(key, group);
    }

    return [...groups.values()]
      .map((group) => ({
        ...group,
        options: [...group.options].sort((first, second) => {
          if (first.id === selected?.id) return -1;
          if (second.id === selected?.id) return 1;
          return shortModelName(first).localeCompare(shortModelName(second));
        }),
      }))
      .sort((first, second) => {
        const selectedGroupKey = selected ? `${selected.provider}:${selected.accountID}` : "";
        if (first.key === selectedGroupKey) return -1;
        if (second.key === selectedGroupKey) return 1;
        return `${first.provider} ${first.accountName}`.localeCompare(`${second.provider} ${second.accountName}`);
      });
  }, [options, search, selected]);

  function changeOpen(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) setSearch("");
  }

  return (
    <>
      <Button
        id={id}
        type="button"
        variant="outline"
        className="h-auto min-h-11 w-full justify-between whitespace-normal px-3 py-2 text-start"
        aria-haspopup="dialog"
        aria-expanded={open}
        onClick={() => setOpen(true)}
      >
        {selected ? (
          <span className="flex min-w-0 flex-1 items-center gap-3">
            <ProviderLogo
              key={`${selected.provider}:${selected.logoURL ?? ""}`}
              provider={selected.provider}
              displayName={selected.providerDisplayName}
              logoURL={selected.logoURL}
              size={28}
              decorative
            />
            <span className="min-w-0 flex-1">
              <span className="block truncate font-medium" dir="ltr" title={shortModelName(selected)}>
                {shortModelName(selected)}
              </span>
              <span className="block truncate text-xs font-normal text-muted-foreground" dir="ltr">
                {selected.provider} / {selected.accountName}
              </span>
            </span>
          </span>
        ) : (
          <span className="text-muted-foreground">{labels.placeholder}</span>
        )}
        <ChevronDown className="shrink-0 text-muted-foreground" aria-hidden="true" />
      </Button>

      <SearchPickerDialog
        open={open}
        onOpenChange={changeOpen}
        title={labels.title}
        searchValue={search}
        onSearchValueChange={setSearch}
        searchPlaceholder={labels.searchPlaceholder}
        size="lg"
        contentClassName="overflow-hidden"
        bodyClassName="min-h-0"
      >
        {visibleGroups.length === 0 ? (
          <EmptyState
            icon={SearchX}
            title={labels.emptyTitle}
            description={labels.emptyDescription}
            variant="inline"
            density="compact"
          />
        ) : (
          <ScrollArea className="h-80 pe-3">
            <div className="space-y-3">
              {visibleGroups.map((group) => (
                <section key={group.key} className="overflow-hidden rounded-xl border border-border bg-card">
                  <div className="flex items-center justify-between gap-3 border-b border-border bg-muted/30 px-3 py-2">
                    <div className="flex min-w-0 items-center gap-2">
                      <ProviderLogo
                        key={`${group.provider}:${group.logoURL ?? ""}`}
                        provider={group.provider}
                        displayName={group.providerDisplayName}
                        logoURL={group.logoURL}
                        size={28}
                        decorative
                      />
                      <div className="min-w-0 flex-1 text-end">
                        <h3 className="truncate text-end text-sm font-semibold" dir="ltr">
                          {group.provider}
                        </h3>
                        <p className="truncate text-end text-xs text-muted-foreground" dir="ltr">
                          {group.accountName}
                        </p>
                      </div>
                    </div>
                    <Badge variant="secondary">{labels.groupCount(group.options.length)}</Badge>
                  </div>

                  <div className="divide-y divide-border">
                    {group.options.map((option) => {
                      const isSelected = option.id === selected?.id;
                      return (
                        <button
                          key={option.id}
                          type="button"
                          className={cn(
                            "flex w-full items-start gap-3 px-3 py-3 text-start transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring focus-visible:outline-none",
                            isSelected && "bg-primary-soft",
                          )}
                          onClick={() => {
                            onSelectedChange(option);
                            changeOpen(false);
                          }}
                        >
                          <span className="min-w-0 flex-1">
                            <span className="flex items-center gap-2">
                              <span className="min-w-0 flex-1 truncate font-mono text-sm font-semibold" title={shortModelName(option)}>
                                {shortModelName(option)}
                              </span>
                              {isSelected ? (
                                <Badge variant="info" className="shrink-0">
                                  <Check className="size-3.5" aria-hidden="true" />
                                  {labels.selected}
                                </Badge>
                              ) : null}
                            </span>
                            {option.modes.length > 0 ? (
                              <span className="mt-2 flex flex-wrap gap-1">
                                {option.modes.map((mode) => (
                                  <Badge key={`${option.id}-${mode}`} variant="outline">
                                    {mode}
                                  </Badge>
                                ))}
                              </span>
                            ) : null}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                </section>
              ))}
            </div>
          </ScrollArea>
        )}
      </SearchPickerDialog>
    </>
  );
}
