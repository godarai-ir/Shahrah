"use client";

import { FilterCombobox } from "@/components/app-ui/FilterCombobox";

type CatalogFilterComboboxOption = {
  value: string;
  label: string;
};

type CatalogFilterComboboxProps = {
  id: string;
  options: CatalogFilterComboboxOption[];
  value: string;
  onValueChange: (value: string) => void;
  placeholder: string;
  emptyText: string;
  searchable?: boolean;
};

export function CatalogFilterCombobox({
  id,
  options,
  value,
  onValueChange,
  placeholder,
  emptyText,
  searchable = true,
}: CatalogFilterComboboxProps) {
  return (
    <FilterCombobox
      id={id}
      options={options}
      value={value}
      onValueChange={onValueChange}
      placeholder={placeholder}
      emptyText={emptyText}
      searchable={searchable}
      className="h-9 bg-background text-foreground transition-colors hover:bg-accent/40 [&_[data-slot=input-group-control]]:text-sm"
    />
  );
}
