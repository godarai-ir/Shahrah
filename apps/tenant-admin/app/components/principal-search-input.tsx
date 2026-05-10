"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Check, ChevronsUpDown, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

export type PrincipalSearchKind = "user" | "team" | "external_identity";
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
  externalIdentities: string;
  everyone: string;
  searchPlaceholder: string;
  triggerPlaceholder: string;
  loading: string;
  empty: string;
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
  showSelectedMeta?: boolean;
}) {
  const kinds = useMemo(() => props.kinds?.length ? props.kinds : (["user", "team"] as PrincipalSearchKind[]), [props.kinds]);
  const [open, setOpen] = useState(false);
  const [activeKind, setActiveKind] = useState<PrincipalSearchKind>(kinds[0] ?? "user");
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<PrincipalSearchOption[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const requestRef = useRef(0);

  useEffect(() => {
    if (!kinds.includes(activeKind)) {
      setActiveKind(kinds[0] ?? "user");
    }
  }, [activeKind, kinds]);

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
  const selectedLabel = selected?.display_name || selected?.principal_id || "";
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
    <div className={cn("min-w-0 space-y-1", props.className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            id={props.id}
            type="button"
            variant="outline"
            role="combobox"
            aria-expanded={open}
            disabled={props.disabled}
            className="h-11 w-full justify-between rounded-md px-3 font-normal"
          >
            <span className="flex min-w-0 items-center gap-3">
              {selected ? (
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
                  {getAvatarText(selected)}
                </span>
              ) : null}
              <span className="min-w-0 text-start">
                <span className={cn("block truncate text-sm", !selected && "text-muted-foreground")}>
                  {selected ? selectedLabel : props.labels.triggerPlaceholder}
                </span>
                {selected && props.showSelectedMeta !== false && selectedMeta ? (
                  <span className="block truncate text-xs text-muted-foreground">{selectedMeta}</span>
                ) : null}
              </span>
            </span>
            <ChevronsUpDown className="ms-2 h-4 w-4 shrink-0 opacity-60" />
          </Button>
        </PopoverTrigger>
        <PopoverContent align="start" className="w-[min(640px,calc(100vw-2rem))] overflow-hidden p-0">
          <div className="border-b p-3">
            <div className="relative">
              <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(event) => handleQueryChange(event.target.value)}
                placeholder={props.labels.searchPlaceholder}
                className="h-10 ps-9"
              />
            </div>
          </div>

          <div className="grid min-h-[230px] sm:grid-cols-[168px_minmax(0,1fr)]">
            <div className="flex gap-1 overflow-x-auto border-b bg-muted/30 p-2 sm:block sm:space-y-1 sm:border-b-0 sm:border-e">
              {kinds.map((kind) => (
                <button
                  key={kind}
                  type="button"
                  className={cn(
                    "w-full whitespace-nowrap rounded-md px-3 py-2 text-start text-sm font-medium transition",
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

            <div className="max-h-[300px] overflow-y-auto p-2">
              {filteredEveryoneVisible ? (
                <PrincipalOptionRow
                  option={{
                    type: "everyone",
                    display_name: props.labels.everyone,
                  }}
                  selected={selected}
                  labels={props.labels}
                  onSelect={selectOption}
                />
              ) : null}

              {isSearching ? (
                <p className="px-3 py-8 text-center text-sm text-muted-foreground">{props.labels.loading}</p>
              ) : results.length === 0 && !filteredEveryoneVisible ? (
                <p className="px-3 py-8 text-center text-sm text-muted-foreground">{props.labels.empty}</p>
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
      {selected?.principal_id && props.showSelectedMeta !== false ? (
        <p className="break-all text-xs text-muted-foreground">ID: {selected.principal_id}</p>
      ) : null}
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
        <span className="block truncate text-sm font-medium">{props.option.display_name}</span>
        <span className="block truncate text-xs text-muted-foreground">
          {props.option.secondary || props.option.principal_id || getKindLabel(props.option.type, props.labels)}
        </span>
      </span>
      <Check className={cn("h-4 w-4 shrink-0", isSelected ? "opacity-100" : "opacity-0")} />
    </button>
  );
}

function getKindLabel(kind: PrincipalSelectionType, labels: PrincipalEntitySelectLabels): string {
  if (kind === "team") return labels.teams;
  if (kind === "external_identity") return labels.externalIdentities;
  if (kind === "everyone") return labels.everyone;
  return labels.users;
}

function getAvatarText(selection: PrincipalEntitySelection): string {
  if (selection.type === "everyone") return "E";
  const label = selection.display_name || selection.principal_id || selection.type;
  return label.trim().slice(0, 1).toUpperCase();
}
