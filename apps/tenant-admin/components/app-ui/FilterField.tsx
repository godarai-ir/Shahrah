"use client";

import * as React from "react";

import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Slider } from "@/components/ui/slider";
import { cn } from "@/lib/utils";

export type FilterFieldProps = React.ComponentProps<typeof Field> & { label: React.ReactNode; htmlFor?: string };

export function FilterField({ label, htmlFor, className, children, ...props }: FilterFieldProps) {
  return <Field className={cn("min-w-0 gap-1", className)} {...props}><FieldLabel htmlFor={htmlFor} className="text-xs text-muted-foreground">{label}</FieldLabel>{children}</Field>;
}

export type DebouncedTextFilterProps = Omit<React.ComponentProps<typeof Input>, "value" | "onChange"> & { label: React.ReactNode; value: string; onValueChange: (value: string) => void; debounceMs?: number; mono?: boolean };

export function DebouncedTextFilter({ label, value, onValueChange, debounceMs = 350, mono, id, className, ...props }: DebouncedTextFilterProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  const [draftState, setDraftState] = React.useState({ source: value, draft: value });
  const draft = draftState.source === value ? draftState.draft : value;
  const onValueChangeRef = React.useRef(onValueChange);
  React.useEffect(() => { onValueChangeRef.current = onValueChange; }, [onValueChange]);
  React.useEffect(() => { if (draft === value) return; const timer = window.setTimeout(() => onValueChangeRef.current(draft), debounceMs); return () => window.clearTimeout(timer); }, [debounceMs, draft, value]);
  return <FilterField label={label} htmlFor={controlID}><Input id={controlID} value={draft} onChange={(event) => setDraftState({ source: value, draft: event.target.value })} dir={mono ? "ltr" : props.dir} className={cn("h-8 text-xs", mono && "text-start font-mono", className)} {...props} /></FilterField>;
}

export type RangeFilterFieldProps = { label: React.ReactNode; value: [number, number]; min: number; max: number; step?: number; disabled?: boolean; minLabel: string; maxLabel: string; formatValue: (value: number) => React.ReactNode; onValueChange?: (value: [number, number]) => void; onValueCommit: (value: [number, number]) => void };

export function RangeFilterField({ label, value, min, max, step = 1, disabled, minLabel, maxLabel, formatValue, onValueChange, onValueCommit }: RangeFilterFieldProps) {
  const id = React.useId();
  return (
    <FilterField label={label} htmlFor={id}>
      <div className="flex min-h-20 flex-col justify-center gap-3 rounded-md border border-border px-3 py-2">
        <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground" dir="ltr"><span title={minLabel}>{formatValue(value[0])}</span><span title={maxLabel}>{formatValue(value[1])}</span></div>
        <Slider id={id} value={value} min={min} max={max} step={step} disabled={disabled} onValueChange={(next) => onValueChange?.([next[0] ?? min, next[1] ?? max])} onValueCommit={(next) => onValueCommit([next[0] ?? min, next[1] ?? max])} />
        <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground" dir="ltr"><span>{formatValue(min)}</span><span>{formatValue(max)}</span></div>
      </div>
    </FilterField>
  );
}
