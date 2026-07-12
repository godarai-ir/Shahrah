import * as React from "react";

import { EmptyState } from "@/components/app-ui/EmptyState";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldContent, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from "@/components/ui/field";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export type CheckboxOption = { value: string; label: React.ReactNode; description?: React.ReactNode; disabled?: boolean; textDir?: "rtl" | "ltr" | "auto" };
export type CheckboxOptionListProps = {
  label: React.ReactNode;
  description?: React.ReactNode;
  options: readonly CheckboxOption[];
  value: readonly string[];
  onValueChange: (value: string[]) => void;
  emptyText: React.ReactNode;
  loading?: boolean;
  loadingLabel?: string;
  disabled?: boolean;
  className?: string;
};

export function CheckboxOptionList({ label, description, options, value, onValueChange, emptyText, loading = false, loadingLabel = "Loading", disabled, className }: CheckboxOptionListProps) {
  const generatedID = React.useId();
  const selected = new Set(value);
  return (
    <FieldSet className={cn("min-w-0", className)} disabled={disabled}>
      <FieldLegend variant="label">{label}</FieldLegend>
      {description ? <FieldDescription>{description}</FieldDescription> : null}
      <div className="max-h-64 overflow-auto rounded-md border border-border p-2">
        {loading ? (
          <div className="flex flex-col gap-2" role="status" aria-label={loadingLabel}><Skeleton className="h-10 w-full" /><Skeleton className="h-10 w-full" /></div>
        ) : options.length === 0 ? (
          <EmptyState variant="inline" density="compact" title={emptyText} />
        ) : (
          <FieldGroup data-slot="checkbox-group" className="gap-1">
            {options.map((option, index) => {
              const id = `${generatedID}-${index}`;
              return (
                <Field key={option.value} orientation="horizontal" data-disabled={disabled || option.disabled || undefined} className="rounded-md px-2 py-2 hover:bg-muted/40">
                  <Checkbox id={id} checked={selected.has(option.value)} disabled={disabled || option.disabled} onCheckedChange={(checked) => onValueChange(checked === true ? [...value, option.value] : value.filter((item) => item !== option.value))} />
                  <FieldContent>
                    <FieldLabel htmlFor={id} dir={option.textDir}>{option.label}</FieldLabel>
                    {option.description ? <FieldDescription>{option.description}</FieldDescription> : null}
                  </FieldContent>
                </Field>
              );
            })}
          </FieldGroup>
        )}
      </div>
    </FieldSet>
  );
}
