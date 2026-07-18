"use client";

import * as React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

export type AppSelectOption = {
  value: string;
  label: React.ReactNode;
  disabled?: boolean;
};

export type AppSelectGroup = {
  key?: React.Key;
  label?: React.ReactNode;
  options: AppSelectOption[];
};

export type AppSelectFieldProps = {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  onBlur?: () => void;
  options?: AppSelectOption[];
  groups?: AppSelectGroup[];
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
  optionTextDir?: "rtl" | "ltr" | "auto";
  size?: "sm" | "default";
  contentClassName?: string;
};

export function AppSelectField({
  value,
  defaultValue,
  onValueChange,
  onBlur,
  options = [],
  groups,
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
  dir,
  optionTextDir = "auto",
  size = "default",
  contentClassName,
}: AppSelectFieldProps) {
  const generatedID = React.useId();
  const triggerID = id ?? generatedID;
  const descriptionID = description ? `${triggerID}-description` : undefined;
  const errorID = error ? `${triggerID}-error` : undefined;
  const describedBy = [descriptionID, errorID].filter(Boolean).join(" ") || undefined;
  const invalid = Boolean(error);
  const optionGroups = groups ?? [{ options }];

  return (
    <Field className={cn("min-w-0", className)} data-invalid={invalid || undefined} data-disabled={disabled || undefined}>
      {label ? <FieldLabel htmlFor={triggerID}>{label}</FieldLabel> : null}
      <Select
        dir={dir}
        value={value}
        defaultValue={defaultValue}
        onValueChange={onValueChange}
        onOpenChange={(open) => {
          if (!open) onBlur?.();
        }}
        disabled={disabled}
        name={name}
        required={required}
      >
        <SelectTrigger
          id={triggerID}
          className={cn("min-w-40", size === "sm" ? "h-8 text-xs" : "h-10", triggerClassName)}
          aria-label={ariaLabel}
          aria-describedby={describedBy}
          aria-invalid={invalid || undefined}
        >
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent dir={dir} className={contentClassName}>
          {optionGroups.map((group, groupIndex) => (
            <React.Fragment key={group.key ?? groupIndex}>
              <SelectGroup>
                {group.label ? <SelectLabel>{group.label}</SelectLabel> : null}
                {group.options.map((option) => (
                  <SelectItem key={option.value} value={option.value} disabled={option.disabled}>
                    <span dir={optionTextDir}>{option.label}</span>
                  </SelectItem>
                ))}
              </SelectGroup>
              {groupIndex < optionGroups.length - 1 ? <SelectSeparator /> : null}
            </React.Fragment>
          ))}
        </SelectContent>
      </Select>
      {description ? <FieldDescription id={descriptionID}>{description}</FieldDescription> : null}
      {error ? <FieldError id={errorID}>{error}</FieldError> : null}
    </Field>
  );
}
