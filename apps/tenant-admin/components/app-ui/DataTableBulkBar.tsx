"use client";

import * as React from "react";
import { X } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export type DataTableBulkAction = {
  key: string;
  label: React.ReactNode;
  onSelect: () => void;
  disabled?: boolean;
  destructive?: boolean;
  icon?: React.ReactNode;
};

export type DataTableBulkBarProps = {
  count: number;
  onClear: () => void;
  /** At most four; a destructive action is rendered last. */
  actions: DataTableBulkAction[];
  className?: string;
};

export function DataTableBulkBar({ count, onClear, actions, className }: DataTableBulkBarProps) {
  const t = useTranslations("common");
  const ordered = [...actions.filter((action) => !action.destructive), ...actions.filter((action) => action.destructive)].slice(0, 4);

  return (
    <div
      data-slot="data-table-bulk-bar"
      role="toolbar"
      aria-label={t("table.bulkActions")}
      className={cn("flex min-h-12 flex-wrap items-center gap-2 rounded-lg border border-primary-subtle-border bg-primary-soft px-2 py-1.5", className)}
    >
      <span className="px-1 text-sm font-medium" aria-live="polite">
        {t("table.selectedCount", { count })}
      </span>
      <Button type="button" variant="ghost" size="sm" onClick={onClear}>
        <X data-icon="inline-start" aria-hidden="true" />
        {t("table.clearSelection")}
      </Button>
      <div className="ms-auto flex flex-wrap items-center gap-2">
        {ordered.map((action) => (
          <Button
            key={action.key}
            type="button"
            size="sm"
            variant={action.destructive ? "destructive" : "outline"}
            disabled={action.disabled}
            onClick={action.onSelect}
          >
            {action.icon}
            {action.label}
          </Button>
        ))}
      </div>
    </div>
  );
}
