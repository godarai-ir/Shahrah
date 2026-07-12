"use client";

import * as React from "react";
import { FormDialog, type FormDialogSize } from "@/components/app-ui/FormDialog";
import { SearchField } from "@/components/app-ui/SearchField";

export type SearchPickerDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description?: React.ReactNode;
  searchValue: string;
  onSearchValueChange: (value: string) => void;
  searchPlaceholder: string;
  children: React.ReactNode;
  size?: FormDialogSize;
  footer?: React.ReactNode;
};

export function SearchPickerDialog({ open, onOpenChange, title, description, searchValue, onSearchValueChange, searchPlaceholder, children, size = "lg", footer }: SearchPickerDialogProps) {
  return <FormDialog open={open} onOpenChange={onOpenChange} title={title} description={description} size={size} footer={footer}><SearchField value={searchValue} onChange={(event) => onSearchValueChange(event.target.value)} placeholder={searchPlaceholder} />{children}</FormDialog>;
}
