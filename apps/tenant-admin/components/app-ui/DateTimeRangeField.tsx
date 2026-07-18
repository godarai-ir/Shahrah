import * as React from "react";
import { Calendar } from "lucide-react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group";
import { fromDateTimeLocalValue, toDateTimeLocalValue } from "@/lib/date-time";
import { cn } from "@/lib/utils";

export type DateTimeFieldProps = {
  id?: string;
  label: React.ReactNode;
  value: string;
  onValueChange: (value: string) => void;
  description?: React.ReactNode;
  error?: React.ReactNode;
  disabled?: boolean;
  fieldClassName?: string;
  inputClassName?: string;
};

export function DateTimeField({ id, label, value, onValueChange, description, error, disabled, fieldClassName, inputClassName }: DateTimeFieldProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  const descriptionID = description ? `${controlID}-description` : undefined;
  const errorID = error ? `${controlID}-error` : undefined;
  const describedBy = [descriptionID, errorID].filter(Boolean).join(" ") || undefined;

  return (
    <Field data-invalid={Boolean(error) || undefined} data-disabled={disabled || undefined} className={fieldClassName}>
      <FieldLabel htmlFor={controlID}>{label}</FieldLabel>
      <InputGroup data-disabled={disabled || undefined}>
        <InputGroupAddon><Calendar aria-hidden="true" /></InputGroupAddon>
        <InputGroupInput id={controlID} type="datetime-local" dir="ltr" value={toDateTimeLocalValue(value)} onChange={(event) => onValueChange(fromDateTimeLocalValue(event.target.value))} disabled={disabled} aria-invalid={Boolean(error) || undefined} aria-describedby={describedBy} className={cn("text-start font-mono text-xs", inputClassName)} />
      </InputGroup>
      {description ? <FieldDescription id={descriptionID}>{description}</FieldDescription> : null}
      {error ? <FieldError id={errorID}>{error}</FieldError> : null}
    </Field>
  );
}

export type DateTimeRangeFieldProps = {
  from: string;
  to: string;
  fromLabel: React.ReactNode;
  toLabel: React.ReactNode;
  onFromChange: (value: string) => void;
  onToChange: (value: string) => void;
  fromID?: string;
  toID?: string;
  disabled?: boolean;
  className?: string;
};

export function DateTimeRangeField({ from, to, fromLabel, toLabel, onFromChange, onToChange, fromID, toID, disabled, className }: DateTimeRangeFieldProps) {
  return (
    <div className={cn("grid min-w-0 gap-3 md:grid-cols-2", className)}>
      <DateTimeField id={fromID} label={fromLabel} value={from} onValueChange={onFromChange} disabled={disabled} />
      <DateTimeField id={toID} label={toLabel} value={to} onValueChange={onToChange} disabled={disabled} />
    </div>
  );
}
