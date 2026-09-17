"use client";

import * as React from "react";
import { Columns3, Rows3 } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { useDataTableContext } from "@/components/ui/data-table";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { TableDensity } from "@/components/ui/table";
import { cn } from "@/lib/utils";

export type DataTableToolbarProps = {
  search?: React.ReactNode;
  filters?: React.ReactNode;
  /** Column menu; shown when the enclosing DataTable has optional columns. */
  columns?: boolean;
  /** Density switch. */
  density?: boolean;
  className?: string;
};

/** Filter bar for a DataTable. Order from the start edge: search, filters, column menu, density. */
export function DataTableToolbar({ search, filters, columns = true, density = true, className }: DataTableToolbarProps) {
  const t = useTranslations("common");
  const table = useDataTableContext();
  const optionalColumns = table?.optionalColumns ?? [];

  return (
    <div
      data-slot="data-table-toolbar"
      className={cn("flex min-h-12 flex-wrap items-center gap-2 rounded-lg border border-border bg-card px-2 py-1.5", className)}
    >
      {search ? <div className="min-w-48 flex-1 sm:max-w-md">{search}</div> : null}
      {filters ? <div className="flex flex-wrap items-center gap-2">{filters}</div> : null}
      <div className="ms-auto flex items-center gap-1">
        {table && columns && optionalColumns.length > 0 ? (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button type="button" variant="ghost" size="sm">
                <Columns3 data-icon="inline-start" aria-hidden="true" />
                {t("table.columns")}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>{t("table.optionalColumns")}</DropdownMenuLabel>
              {optionalColumns.map((column) => (
                <DropdownMenuCheckboxItem
                  key={column.id}
                  checked={column.visible}
                  onCheckedChange={(checked) => table.setColumnVisible(column.id, checked === true)}
                  onSelect={(event) => event.preventDefault()}
                >
                  {column.label}
                </DropdownMenuCheckboxItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        ) : null}
        {table && density ? (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button type="button" variant="ghost" size="sm">
                <Rows3 data-icon="inline-start" aria-hidden="true" />
                {t("table.density")}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuRadioGroup value={table.density} onValueChange={(value) => table.setDensity(value as TableDensity)}>
                <DropdownMenuRadioItem value="compact">{t("table.densityCompact")}</DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="relaxed">{t("table.densityRelaxed")}</DropdownMenuRadioItem>
              </DropdownMenuRadioGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : null}
      </div>
    </div>
  );
}
