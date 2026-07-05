"use client";

import * as React from "react";

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
  className,
  triggerClassName,
  ariaLabel,
  disabled,
  dir = "rtl",
}: AppSelectFieldProps) {
  return (
    <div className={cn("flex min-w-0 flex-col gap-1.5", className)}>
      {label ? <span className="text-sm font-medium text-foreground">{label}</span> : null}
      <Select dir={dir} value={value} onValueChange={onValueChange} disabled={disabled}>
        <SelectTrigger className={cn("min-w-40", triggerClassName)} aria-label={ariaLabel}>
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
      {description ? <p className="text-xs leading-5 text-muted-foreground">{description}</p> : null}
    </div>
  );
}
