"use client";

import * as React from "react";
import { X } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";

export type ChipInputProps = {
  id?: string;
  value: readonly string[];
  onValueChange: (value: string[]) => void;
  parseValues?: (raw: string) => string[];
  getRemoveLabel: (value: string) => string;
  placeholder?: string;
  disabled?: boolean;
};

const defaultParseValues = (raw: string) => raw.split(/[,\s]+/).map((item) => item.trim()).filter(Boolean);

export function ChipInput({ id, value, onValueChange, parseValues = defaultParseValues, getRemoveLabel, placeholder, disabled }: ChipInputProps) {
  const [draft, setDraft] = React.useState("");
  function remove(item: string) { onValueChange(value.filter((current) => current !== item)); }
  function commit() {
    const nextItems = parseValues(draft);
    if (nextItems.length) onValueChange(Array.from(new Set([...value, ...nextItems])));
    setDraft("");
  }
  return (
    <div className="rounded-md border border-input bg-background px-2 py-2">
      <div className="flex flex-wrap items-center gap-2">
        {value.map((item) => <Badge key={item} variant="secondary" className="gap-1 pe-1"><span>{item}</span><button type="button" className="rounded-sm text-muted-foreground hover:text-foreground" onClick={() => remove(item)} aria-label={getRemoveLabel(item)} disabled={disabled}><X aria-hidden="true" /></button></Badge>)}
        <Input id={id} value={draft} onChange={(event) => setDraft(event.target.value)} onBlur={commit} onKeyDown={(event) => { if (event.key === "Enter" || event.key === ",") { event.preventDefault(); commit(); } else if (event.key === "Backspace" && !draft && value.length) remove(value[value.length - 1]); }} placeholder={placeholder} className="h-8 min-w-32 flex-1 border-0 p-0 ps-1 shadow-none focus-visible:ring-0" disabled={disabled} />
      </div>
    </div>
  );
}
