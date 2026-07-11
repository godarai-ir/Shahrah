"use client";

import * as React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

type AppSelectOption = {
  value: string;
  label: React.ReactNode;
  disabled?: boolean;
};

type AppSelectFieldProps = {
  value: string;
  onValueChange: (value: string) => void;
  options: AppSelectOption[];
  placeholder?: string;
  label?: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  id?: string;
  name?: string;
  required?: boolean;
  className?: string;
  triggerClassName?: string;
  ariaLabel?: string;
  disabled?: boolean;
  dir?: "rtl" | "ltr";
};

export function AppSelectField({
  value,
  onValueChange,
  options,
  placeholder,
  label,
  description,
  error,
  id,
  name,
  required,
  className,
  triggerClassName,
  ariaLabel,
  disabled,
  dir = "rtl",
}: AppSelectFieldProps) {
  const generatedID = React.useId();
  const triggerID = id ?? generatedID;
  const descriptionID = description ? `${triggerID}-description` : undefined;
  const errorID = error ? `${triggerID}-error` : undefined;
  const describedBy = [descriptionID, errorID].filter(Boolean).join(" ") || undefined;
  const invalid = Boolean(error);

  return (
    <Field className={cn("min-w-0", className)} data-invalid={invalid || undefined} data-disabled={disabled || undefined}>
      {label ? <FieldLabel htmlFor={triggerID}>{label}</FieldLabel> : null}
      <Select dir={dir} value={value} onValueChange={onValueChange} disabled={disabled} name={name} required={required}>
        <SelectTrigger id={triggerID} className={cn("min-w-40", triggerClassName)} aria-label={ariaLabel} aria-describedby={describedBy} aria-invalid={invalid || undefined}>
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent dir={dir}>
          <SelectGroup>
            {options.map((option) => (
              <SelectItem key={option.value} value={option.value} disabled={option.disabled}>
                {option.label}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
      {description ? <FieldDescription id={descriptionID}>{description}</FieldDescription> : null}
      {error ? <FieldError id={errorID}>{error}</FieldError> : null}
    </Field>
  );
}
