"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Check, ChevronsUpDown, Search } from "lucide-react";

import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { hideMobileIdentity } from "@/lib/auth/mobile-auth-ui";
import { cn } from "@/lib/utils";

export type PrincipalSearchKind = "user" | "team" | "org_unit" | "virtual_account" | "external_identity";
export type PrincipalSelectionType = PrincipalSearchKind | "everyone";

export type PrincipalSearchOption = {
  principal_id: string;
  display_name: string;
  secondary?: string;
  type?: PrincipalSearchKind;
};

export type PrincipalEntitySelection = {
  type: PrincipalSelectionType;
  principal_id?: string;
  display_name: string;
  secondary?: string;
};

export type PrincipalEntitySelectLabels = {
  users: string;
  teams: string;
  orgUnits: string;
  virtualAccounts: string;
  externalIdentities: string;
  everyone: string;
  searchPlaceholder: string;
  triggerPlaceholder: string;
  loading: string;
  empty: string;
  minSearchHint?: string;
  everyoneWarning?: string;
};

export function PrincipalEntitySelect(props: {
  id: string;
  selected: PrincipalEntitySelection | null;
  onSelectedChange: (value: PrincipalEntitySelection | null) => void;
  onSearch: (kind: PrincipalSearchKind, query: string) => Promise<PrincipalSearchOption[]>;
  labels: PrincipalEntitySelectLabels;
  kinds?: PrincipalSearchKind[];
  includeEveryone?: boolean;
  disabled?: boolean;
  className?: string;
  triggerClassName?: string;
  compact?: boolean;
  ariaLabel?: string;
}) {
  const kinds = useMemo(() => props.kinds?.length ? props.kinds : (["user", "team"] as PrincipalSearchKind[]), [props.kinds]);
  const [open, setOpen] = useState(false);
  const [activeKindState, setActiveKind] = useState<PrincipalSearchKind>(kinds[0] ?? "user");
  const activeKind = kinds.includes(activeKindState) ? activeKindState : (kinds[0] ?? "user");
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<PrincipalSearchOption[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const requestRef = useRef(0);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  function runSearch(nextKind: PrincipalSearchKind, nextQuery: string) {
    if (timerRef.current) clearTimeout(timerRef.current);
    const normalized = nextQuery.trim();
    if (normalized.length < 2) {
      setResults([]);
      setIsSearching(false);
      return;
    }

    timerRef.current = setTimeout(() => {
      const requestID = ++requestRef.current;
      setIsSearching(true);
      void props.onSearch(nextKind, normalized)
        .then((found) => {
          if (requestID !== requestRef.current) return;
          setResults(found.map((item) => ({ ...item, type: item.type ?? nextKind })));
        })
        .finally(() => {
          if (requestID === requestRef.current) setIsSearching(false);
        });
    }, 250);
  }

  function handleQueryChange(nextQuery: string) {
    setQuery(nextQuery);
    runSearch(activeKind, nextQuery);
  }

  function handleKindChange(nextKind: PrincipalSearchKind) {
    setActiveKind(nextKind);
    setResults([]);
    runSearch(nextKind, query);
  }

  function selectOption(option: PrincipalEntitySelection) {
    props.onSelectedChange(option);
    setOpen(false);
    setQuery("");
    setResults([]);
  }

  const selected = props.selected;
  const selectedLabel = hideMobileIdentity(selected?.display_name) || selected?.principal_id || "";
  const selectedMeta =
    selected?.type === "everyone"
      ? props.labels.users
      : selected
        ? getKindLabel(selected.type, props.labels)
        : "";
  const filteredEveryoneVisible =
    props.includeEveryone &&
    (!query.trim() || props.labels.everyone.toLowerCase().includes(query.trim().toLowerCase()));

  return (
    <div className={cn("flex min-w-0 flex-col gap-1", props.className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            id={props.id}
            type="button"
            variant="outline"
            role="combobox"
            aria-expanded={open}
            aria-label={props.ariaLabel}
            disabled={props.disabled}
            className={cn(
              "w-full justify-between rounded-md px-3 font-normal",
              props.compact ? "h-8 min-h-8 py-0 text-xs" : "h-auto min-h-11 py-2",
              props.triggerClassName,
            )}
          >
            <span className="flex min-w-0 flex-1 items-center gap-3 overflow-hidden">
              {selected && !props.compact ? (
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
                  {getAvatarText(selected)}
                </span>
              ) : null}
              <span className="min-w-0 flex-1 text-start">
                <span className={cn("block truncate", props.compact ? "text-xs" : "text-sm", !selected && "text-muted-foreground")}>
                  {selected ? selectedLabel : props.labels.triggerPlaceholder}
                </span>
                {selected && selectedMeta && !props.compact ? (
                  <span className="block truncate text-xs text-muted-foreground">{selectedMeta}</span>
                ) : null}
              </span>
            </span>
            <ChevronsUpDown className={cn("ms-2 shrink-0 opacity-60", props.compact ? "h-3.5 w-3.5" : "h-4 w-4")} />
          </Button>
        </PopoverTrigger>
        <PopoverContent
          align="start"
          sideOffset={4}
          className="w-[var(--radix-popover-trigger-width)] min-w-[min(22rem,calc(100vw-2rem))] max-w-[calc(100vw-2rem)] overflow-hidden p-0 sm:w-[min(36rem,calc(100vw-2rem))]"
        >
          <div className="border-b p-2 sm:p-3">
            <div className="relative">
              <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <DisplayNameField
                label={props.ariaLabel ?? props.labels.searchPlaceholder}
                value={query}
                onChange={(event) => handleQueryChange(event.target.value)}
                placeholder={props.labels.searchPlaceholder}
                fieldClassName="gap-0 [&>[data-slot=field-label]]:sr-only"
                className="h-9 ps-9 text-sm sm:h-10"
              />
            </div>
          </div>

          <div className="grid min-h-[190px] grid-rows-[auto_minmax(0,1fr)] sm:min-h-[230px] sm:grid-cols-[168px_minmax(0,1fr)] sm:grid-rows-none">
            <div className="grid auto-cols-fr grid-flow-col gap-1 overflow-x-auto border-b bg-muted/30 p-2 sm:flex sm:flex-col sm:border-b-0 sm:border-e">
              {kinds.map((kind) => (
                <button
                  key={kind}
                  type="button"
                  className={cn(
                    "min-w-0 whitespace-nowrap rounded-md px-2.5 py-2 text-center text-xs font-medium transition sm:w-full sm:px-3 sm:text-start sm:text-sm",
                    activeKind === kind
                      ? "bg-primary/10 text-primary ring-1 ring-primary/30"
                      : "text-muted-foreground hover:bg-background hover:text-foreground",
                  )}
                  onClick={() => handleKindChange(kind)}
                >
                  {getKindLabel(kind, props.labels)}
                </button>
              ))}
            </div>

            <div className="max-h-[260px] overflow-y-auto p-2 sm:max-h-[300px]">
              {filteredEveryoneVisible ? (
                <PrincipalOptionRow
                  option={{
                    type: "everyone",
                    display_name: props.labels.everyone,
                    secondary: props.labels.everyoneWarning,
                  }}
                  selected={selected}
                  labels={props.labels}
                  onSelect={selectOption}
                />
              ) : null}

              {isSearching ? (
                <p className="px-3 py-8 text-center text-sm text-muted-foreground">{props.labels.loading}</p>
              ) : results.length === 0 && !filteredEveryoneVisible ? (
                <p className="px-3 py-8 text-center text-sm text-muted-foreground">
                  {query.trim().length < 2 ? props.labels.minSearchHint ?? props.labels.empty : props.labels.empty}
                </p>
              ) : (
                results.map((item) => (
                  <PrincipalOptionRow
                    key={`${item.type ?? activeKind}-${item.principal_id}`}
                    option={{
                      type: item.type ?? activeKind,
                      principal_id: item.principal_id,
                      display_name: item.display_name,
                      secondary: item.secondary,
                    }}
                    selected={selected}
                    labels={props.labels}
                    onSelect={selectOption}
                  />
                ))
              )}
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}

function PrincipalOptionRow(props: {
  option: PrincipalEntitySelection;
  selected: PrincipalEntitySelection | null;
  labels: PrincipalEntitySelectLabels;
  onSelect: (option: PrincipalEntitySelection) => void;
}) {
  const isSelected =
    props.selected?.type === props.option.type &&
    (props.selected.principal_id ?? "") === (props.option.principal_id ?? "");

  return (
    <button
      type="button"
      className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-start transition hover:bg-accent"
      onClick={() => props.onSelect(props.option)}
    >
      <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
        {getAvatarText(props.option)}
      </span>
      <span className="min-w-0 flex-1">
        <span className="block truncate text-sm font-medium">
          {hideMobileIdentity(props.option.display_name) || props.option.principal_id || getKindLabel(props.option.type, props.labels)}
        </span>
        <span className="block truncate text-xs text-muted-foreground">
          {hideMobileIdentity(props.option.secondary) || getKindLabel(props.option.type, props.labels)}
        </span>
      </span>
      <Check className={cn("h-4 w-4 shrink-0", isSelected ? "opacity-100" : "opacity-0")} />
    </button>
  );
}

function getKindLabel(kind: PrincipalSelectionType, labels: PrincipalEntitySelectLabels): string {
  if (kind === "team") return labels.teams;
  if (kind === "org_unit") return labels.orgUnits;
  if (kind === "virtual_account") return labels.virtualAccounts;
  if (kind === "external_identity") return labels.externalIdentities;
  if (kind === "everyone") return labels.everyone;
  return labels.users;
}

function getAvatarText(selection: PrincipalEntitySelection): string {
  if (selection.type === "everyone") return "E";
  const label = hideMobileIdentity(selection.display_name) || selection.principal_id || selection.type;
  return label.trim().slice(0, 1).toUpperCase();
}
