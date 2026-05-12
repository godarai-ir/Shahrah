"use client";

import { useMemo, useState } from "react";
import { Check, ChevronsUpDown, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

export type ModelSelectionKind = "provider_allowed" | "virtual_model";
export type ModelSelection = { id: string; name: string; kind: ModelSelectionKind; secondary?: string };
export type ModelEntitySelectLabels = {
  triggerPlaceholder: string;
  searchPlaceholder: string;
  providerAllowed: string;
  virtualModels: string;
  loading: string;
  empty: string;
};

export function ModelEntitySelect(props: {
  id: string;
  selected: ModelSelection | null;
  onSelectedChange: (value: ModelSelection | null) => void;
  providerAllowedModels: ModelSelection[];
  virtualModels: ModelSelection[];
  labels: ModelEntitySelectLabels;
  showVirtualModels?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const showVirtualModels = props.showVirtualModels ?? true;
  const [requestedKind, setRequestedKind] = useState<ModelSelectionKind>("provider_allowed");
  const activeKind: ModelSelectionKind = showVirtualModels ? requestedKind : "provider_allowed";

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const source = activeKind === "provider_allowed" ? props.providerAllowedModels : props.virtualModels;
    if (!q) return source;
    return source.filter((item) => `${item.name} ${item.secondary ?? ""}`.toLowerCase().includes(q));
  }, [activeKind, props.providerAllowedModels, props.virtualModels, query]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button id={props.id} type="button" variant="outline" role="combobox" aria-expanded={open} className="h-11 w-full justify-between rounded-md px-3 font-normal">
          <span className={cn("truncate text-start", !props.selected && "text-muted-foreground")}>
            {props.selected ? props.selected.name : props.labels.triggerPlaceholder}
          </span>
          <ChevronsUpDown className="ms-2 h-4 w-4 shrink-0 opacity-60" />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-[min(640px,calc(100vw-2rem))] overflow-hidden p-0">
        <div className="border-b p-3">
          <div className="relative">
            <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder={props.labels.searchPlaceholder} className="h-10 ps-9" />
          </div>
        </div>
        <div className="grid min-h-[230px] sm:grid-cols-[168px_minmax(0,1fr)]">
          <div className="flex gap-1 overflow-x-auto border-b bg-muted/30 p-2 sm:block sm:space-y-1 sm:border-b-0 sm:border-e">
            <button type="button" className={cn("w-full whitespace-nowrap rounded-md px-3 py-2 text-start text-sm font-medium transition", activeKind === "provider_allowed" ? "bg-primary/10 text-primary ring-1 ring-primary/30" : "text-muted-foreground hover:bg-background hover:text-foreground")} onClick={() => setRequestedKind("provider_allowed")}>
              {props.labels.providerAllowed}
            </button>
            {showVirtualModels ? (
              <button type="button" className={cn("w-full whitespace-nowrap rounded-md px-3 py-2 text-start text-sm font-medium transition", activeKind === "virtual_model" ? "bg-primary/10 text-primary ring-1 ring-primary/30" : "text-muted-foreground hover:bg-background hover:text-foreground")} onClick={() => setRequestedKind("virtual_model")}>
                {props.labels.virtualModels}
              </button>
            ) : null}
          </div>
          <div className="max-h-[300px] overflow-y-auto p-2">
            {filtered.length === 0 ? (
              <p className="px-3 py-8 text-center text-sm text-muted-foreground">{props.labels.empty}</p>
            ) : (
              filtered.map((item) => {
                const isSelected = props.selected?.id === item.id && props.selected?.kind === item.kind;
                return (
                  <button key={`${item.kind}:${item.id}`} type="button" className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-start transition hover:bg-accent" onClick={() => { props.onSelectedChange(item); setOpen(false); setQuery(""); }}>
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-medium">{item.name}</span>
                      <span className="block truncate text-xs text-muted-foreground">{item.secondary || (item.kind === "provider_allowed" ? props.labels.providerAllowed : props.labels.virtualModels)}</span>
                    </span>
                    <Check className={cn("h-4 w-4 shrink-0", isSelected ? "opacity-100" : "opacity-0")} />
                  </button>
                );
              })
            )}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
