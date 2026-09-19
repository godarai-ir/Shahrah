"use client";

import { memo, type Dispatch, type SetStateAction } from "react";
import { flexRender, getCoreRowModel, useReactTable, type ColumnDef } from "@tanstack/react-table";
import { Columns3, Filter } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";

import { COLUMN_CHOICES, type RequestLog } from "./logs-types";

export const LogsTable = memo(function LogsTable({
  rows,
  columns,
  pageStart,
  pageEnd,
  total,
  offset,
  limit,
  loading,
  visibleColumns,
  setVisibleColumns,
  onPreviousPage,
  onNextPage,
  hasNextPage,
  onPageLimitChange,
  t,
}: {
  rows: RequestLog[];
  columns: ColumnDef<RequestLog>[];
  pageStart: number;
  pageEnd: number;
  total: number;
  offset: number;
  limit: number;
  loading: boolean;
  visibleColumns: string[];
  setVisibleColumns: Dispatch<SetStateAction<string[]>>;
  onPreviousPage: () => void;
  onNextPage: () => void;
  hasNextPage: boolean;
  onPageLimitChange: (value: number) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <AppCard
      title={t("table.title")}
      actions={
        <div className="flex items-center gap-2">
          <ColumnMenu visibleColumns={visibleColumns} setVisibleColumns={setVisibleColumns} t={t} />
        </div>
      }
      headerClassName="flex-row items-center gap-2"
      titleContainerClassName="flex-1"
      actionsClassName="flex-nowrap"
    >
      {rows.length ? (
        <LogsDataTable
          columns={columns}
          data={rows}
          emptyLabel={t("empty")}
        />
      ) : (
        <EmptyState title={t("empty")} description={t("emptyDescription")} icon={Filter} />
      )}
      <div className="flex flex-wrap items-center justify-between gap-2 pt-4">
        <div className="text-xs text-muted-foreground">{t("table.description", { start: pageStart, end: pageEnd, total })}</div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" disabled={offset === 0 || loading} onClick={onPreviousPage}>{t("actions.previous")}</Button>
          <AppSelectField
            value={String(limit)}
            onValueChange={(value) => onPageLimitChange(Number(value))}
            options={[25, 50, 100, 200].map((value) => ({ value: String(value), label: value }))}
            ariaLabel={t("table.pageSize")}
            triggerClassName="h-8 min-w-20 text-xs"
          />
          <Button variant="outline" size="sm" disabled={!hasNextPage || loading} onClick={onNextPage}>{t("actions.next")}</Button>
        </div>
      </div>
    </AppCard>
  );
});

function LogsDataTable({
  columns,
  data,
  emptyLabel,
}: {
  columns: ColumnDef<RequestLog>[];
  data: RequestLog[];
  emptyLabel: string;
}) {
  // TanStack Table intentionally returns table helpers that React Compiler cannot memoize safely.
  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div className="@container overflow-x-auto rounded-lg border">
      <Table className="table-fixed text-xs">
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id} className={logsColumnClassName(header.column.id)}>
                  {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.length ? (
            table.getRowModel().rows.map((row) => (
              <TableRow key={row.id} data-state={row.getIsSelected() && "selected"} className={cn(row.original.status_code >= 400 && "bg-destructive/4")}>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id} className={logsColumnClassName(cell.column.id)}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell colSpan={columns.length} className="py-8 text-center text-muted-foreground">
                {emptyLabel}
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
}

function logsColumnClassName(columnID: string): string {
  switch (columnID) {
    case "status":
      return "w-20";
    case "started_at":
      return "w-36";
    case "model":
      return "hidden @sm:table-cell";
    case "provider":
      return "hidden w-28 @lg:table-cell";
    case "latency":
      return "hidden w-24 @2xl:table-cell";
    case "tokens":
      return "hidden w-20 @3xl:table-cell";
    case "cost":
      return "hidden w-28 @4xl:table-cell";
    case "signals":
      return "hidden w-32 @4xl:table-cell";
    case "route":
    case "payload":
      return "hidden @5xl:table-cell";
    case "actions":
      return "w-14";
    default:
      return "hidden @5xl:table-cell";
  }
}

function ColumnMenu({
  visibleColumns,
  setVisibleColumns,
  t,
}: {
  visibleColumns: string[];
  setVisibleColumns: Dispatch<SetStateAction<string[]>>;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-8 px-2 sm:px-3" aria-label={t("actions.columns")} title={t("actions.columns")}>
          <Columns3 className="h-4 w-4" />
          <span className="hidden sm:inline">{t("actions.columns")}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48 text-start">
        <div>
          {COLUMN_CHOICES.map((column) => (
            <DropdownMenuCheckboxItem
              key={column}
              checked={visibleColumns.includes(column)}
              className="text-start"
              onCheckedChange={(checked) => {
                setVisibleColumns((current) => checked ? [...current, column] : current.filter((item) => item !== column));
              }}
            >
              <span className="w-full text-start">{t(`columns.${column}`)}</span>
            </DropdownMenuCheckboxItem>
          ))}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
