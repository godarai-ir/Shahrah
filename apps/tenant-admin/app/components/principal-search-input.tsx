"use client";

import { useRef, useState } from "react";
import { Check, ChevronsUpDown } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

export type PrincipalSearchKind = "user" | "team";

export type PrincipalSearchOption = {
  principal_id: string;
  display_name: string;
  secondary?: string;
};

export function PrincipalSearchInput(props: {
  id: string;
  kind: PrincipalSearchKind;
  placeholder: string;
  loadingLabel: string;
  emptyLabel: string;
  selectedID: string;
  onSelectedIDChange: (value: string) => void;
  onSelectedOptionChange?: (value: PrincipalSearchOption | null) => void;
  onSearch: (kind: PrincipalSearchKind, query: string) => Promise<PrincipalSearchOption[]>;
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  const [results, setResults] = useState<PrincipalSearchOption[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [selectedName, setSelectedName] = useState("");

  function scheduleSearch(nextQuery: string) {
    if (timerRef.current) clearTimeout(timerRef.current);
    const normalized = nextQuery.trim();
    if (normalized.length < 2) {
      setResults([]);
      setIsSearching(false);
      return;
    }
    timerRef.current = setTimeout(() => {
      void (async () => {
        setIsSearching(true);
        try {
          const found = await props.onSearch(props.kind, normalized);
          setResults(found);
        } finally {
          setIsSearching(false);
        }
      })();
    }, 250);
  }

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
            className="h-10 w-full justify-between font-normal"
          >
            <span className="truncate text-start">
              {props.selectedID ? (selectedName || props.selectedID) : props.placeholder}
            </span>
            <ChevronsUpDown className="ms-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0">
          <Command shouldFilter={false}>
            <CommandInput
              placeholder={props.placeholder}
              onValueChange={(value) => {
                if (!value.trim()) {
                  props.onSelectedIDChange("");
                  setSelectedName("");
                  props.onSelectedOptionChange?.(null);
                }
                scheduleSearch(value);
              }}
            />
            <CommandList>
              <CommandEmpty>{isSearching ? props.loadingLabel : props.emptyLabel}</CommandEmpty>
              <CommandGroup>
                {results.map((item) => (
                  <CommandItem
                    key={`${props.kind}-${item.principal_id}`}
                    value={`${item.display_name} ${item.principal_id}`}
                    onSelect={() => {
                      props.onSelectedIDChange(item.principal_id);
                      setSelectedName(item.display_name);
                      props.onSelectedOptionChange?.(item);
                      setResults([]);
                      setOpen(false);
                    }}
                  >
                    <Check
                      className={cn(
                        "mr-2 h-4 w-4",
                        props.selectedID === item.principal_id ? "opacity-100" : "opacity-0",
                      )}
                    />
                    <div className="min-w-0">
                      <div className="truncate font-medium">{item.display_name}</div>
                      <div className="truncate text-xs text-muted-foreground">{item.principal_id}</div>
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      {props.selectedID ? <p className="text-xs text-muted-foreground">ID: {props.selectedID}</p> : null}
    </div>
  );
}
