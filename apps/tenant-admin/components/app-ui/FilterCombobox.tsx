"use client";

import * as React from "react";

import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { cn } from "@/lib/utils";

type FilterComboboxOption = {
  value: string;
  label: string;
};

type FilterComboboxProps = {
  id: string;
  options: FilterComboboxOption[];
  value: string;
  onValueChange: (value: string) => void;
  placeholder: string;
  emptyText: string;
  className?: string;
  mono?: boolean;
  searchable?: boolean;
};

export function FilterCombobox({
  id,
  options,
  value,
  onValueChange,
  placeholder,
  emptyText,
  className,
  mono,
  searchable = true,
}: FilterComboboxProps) {
  const selectedOption = options.find((option) => option.value === value) ?? null;

  return (
    <Combobox
      items={options}
      value={selectedOption}
      onValueChange={(option) => {
        if (option) onValueChange(option.value);
      }}
      itemToStringValue={(option) => option.label}
    >
      <ComboboxInput
        id={id}
        dir="rtl"
        placeholder={placeholder}
        readOnly={!searchable}
        className={cn(
          "h-8 w-full bg-background [&_[data-slot=input-group-control]]:min-w-0 [&_[data-slot=input-group-control]]:text-start [&_[data-slot=input-group-control]]:text-xs",
          mono && "[&_[data-slot=input-group-control]]:font-mono",
          className,
        )}
      />
      <ComboboxContent align="start" className="text-start" dir="rtl">
        <ComboboxEmpty>{emptyText}</ComboboxEmpty>
        <ComboboxList>
          {(option) => (
            <ComboboxItem key={option.value} value={option} className={cn("text-start", mono && "font-mono")} dir="rtl">
              <span className="min-w-0 truncate" dir="auto">
                {option.label}
              </span>
            </ComboboxItem>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  );
}
