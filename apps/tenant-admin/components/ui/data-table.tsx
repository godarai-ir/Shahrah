"use client";

import * as React from "react";
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
  type RowData,
  type RowSelectionState,
  type Table as TanStackTable,
  type Updater,
  type VisibilityState,
} from "@tanstack/react-table";
import { useTranslations } from "next-intl";

import { TableLoadingRows } from "@/components/app-ui/TableStateRows";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  type TableDensity,
} from "@/components/ui/table";
import { useTablePreferences } from "@/hooks/use-table-preferences";
import { cn } from "@/lib/utils";

declare module "@tanstack/react-table" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData extends RowData, TValue> {
    /** Shown in the column menu so users can hide it. */
    optional?: boolean;
    /** Optional columns only: start hidden until the user turns them on. */
    defaultHidden?: boolean;
    /** Column name for the column menu. */
    label?: string;
    headerClassName?: string;
    cellClassName?: string;
  }
}

export const SELECTION_COLUMN_ID = "__select";

export type DataTableColumnOption = { id: string; label: string; visible: boolean };

type DataTableContextValue = {
  density: TableDensity;
  setDensity: (density: TableDensity) => void;
  optionalColumns: DataTableColumnOption[];
  setColumnVisible: (id: string, visible: boolean) => void;
};

const DataTableContext = React.createContext<DataTableContextValue | null>(null);

/** Density and column controls of the enclosing DataTable; null outside one. */
export function useDataTableContext() {
  return React.useContext(DataTableContext);
}

export type DataTableProps<TData, TValue> = {
  tableId: string;
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  dir?: "rtl" | "ltr" | "auto";
  className?: string;

  state?: "ready" | "loading" | "error";
  emptyLabel: string;
  noResultsSlot?: React.ReactNode;
  errorSlot?: React.ReactNode;
  skeletonRows?: number;

  enableRowSelection?: boolean;
  getRowId?: (row: TData) => string;
  onSelectionChange?: (ids: string[]) => void;

  toolbar?: React.ReactNode;
  bulkBar?: (ids: string[], clear: () => void) => React.ReactNode;
  footer?: React.ReactNode;
};

export function DataTable<TData, TValue>({
  tableId,
  columns,
  data,
  dir,
  className,
  state = "ready",
  emptyLabel,
  noResultsSlot,
  errorSlot,
  skeletonRows = 5,
  enableRowSelection = false,
  getRowId,
  onSelectionChange,
  toolbar,
  bulkBar,
  footer,
}: DataTableProps<TData, TValue>) {
  const t = useTranslations("common");
  const defaultHidden = React.useMemo(
    () => columns.filter((column) => column.meta?.optional && column.meta.defaultHidden).map(getColumnId),
    [columns],
  );
  const { density, setDensity, columnVisibility, setColumnVisibility } = useTablePreferences(tableId, {
    hiddenColumns: defaultHidden,
  });
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});

  // Pages are fetched server-side, so a different set of rows means a new selection.
  const rowIdsKey = getRowId ? data.map(getRowId).join(",") : String(data.length);
  React.useEffect(() => {
    setRowSelection((current) => (Object.keys(current).length ? {} : current));
  }, [rowIdsKey]);

  const tableColumns = React.useMemo<ColumnDef<TData, TValue>[]>(() => {
    if (!enableRowSelection) return columns;
    const selectColumn: ColumnDef<TData, TValue> = {
      id: SELECTION_COLUMN_ID,
      enableHiding: false,
      meta: { headerClassName: "w-10", cellClassName: "w-10" },
      header: ({ table }) => (
        <Checkbox
          aria-label={t("table.selectAll")}
          checked={table.getIsAllRowsSelected() ? true : table.getIsSomeRowsSelected() ? "indeterminate" : false}
          onCheckedChange={(value) => table.toggleAllRowsSelected(value === true)}
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          aria-label={t("table.selectRow")}
          checked={row.getIsSelected()}
          disabled={!row.getCanSelect()}
          onCheckedChange={(value) => row.toggleSelected(value === true)}
        />
      ),
    };
    return [selectColumn, ...columns];
  }, [columns, enableRowSelection, t]);

  const handleVisibilityChange = (updater: Updater<VisibilityState>) => {
    setColumnVisibility(typeof updater === "function" ? updater(columnVisibility) : updater);
  };

  // TanStack Table intentionally returns table helpers that React Compiler cannot memoize safely.
  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data,
    columns: tableColumns,
    getRowId: getRowId ? (row) => getRowId(row) : undefined,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualFiltering: true,
    manualSorting: true,
    enableRowSelection,
    state: { columnVisibility, rowSelection },
    onColumnVisibilityChange: handleVisibilityChange,
    onRowSelectionChange: setRowSelection,
  });

  const selectedIds = React.useMemo(
    () => Object.keys(rowSelection).filter((id) => rowSelection[id]),
    [rowSelection],
  );
  const selectedKey = selectedIds.join(",");
  const onSelectionChangeRef = React.useRef(onSelectionChange);
  React.useEffect(() => {
    onSelectionChangeRef.current = onSelectionChange;
  });
  React.useEffect(() => {
    onSelectionChangeRef.current?.(selectedKey ? selectedKey.split(",") : []);
  }, [selectedKey]);

  const clearSelection = React.useCallback(() => setRowSelection({}), []);

  // The table instance is stable across renders, so column options are rebuilt every render.
  const context: DataTableContextValue = {
    density,
    setDensity,
    optionalColumns: optionalColumnOptions(table),
    setColumnVisible: (id, visible) => setColumnVisibility({ ...columnVisibility, [id]: visible }),
  };

  const colSpan = Math.max(1, table.getVisibleLeafColumns().length);
  const rows = table.getRowModel().rows;
  const showBulkBar = enableRowSelection && bulkBar && selectedIds.length > 0;
  const bar = showBulkBar ? bulkBar(selectedIds, clearSelection) : toolbar;

  return (
    <DataTableContext.Provider value={context}>
      <div data-slot="data-table" className={cn("flex min-w-0 flex-col gap-3", className)}>
        {bar ? <div className="sticky top-16 z-10">{bar}</div> : null}
        <Table className="text-xs lg:text-sm" density={density} dir={dir}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    className={header.column.columnDef.meta?.headerClassName}
                    style={{ width: header.getSize() }}
                  >
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {state === "loading" ? (
              <TableLoadingRows colSpan={colSpan} rowCount={skeletonRows} label={t("loading")} skeletonClassName="h-5" />
            ) : state === "error" ? (
              <TableRow>
                <TableCell colSpan={colSpan} className="py-6">{errorSlot}</TableCell>
              </TableRow>
            ) : rows.length ? (
              rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() ? "selected" : undefined}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id} className={cell.column.columnDef.meta?.cellClassName}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : noResultsSlot ? (
              <TableRow>
                <TableCell colSpan={colSpan} className="py-6">{noResultsSlot}</TableCell>
              </TableRow>
            ) : (
              <TableRow>
                <TableCell colSpan={colSpan} className="py-8 text-center text-muted-foreground" dir={dir}>
                  {emptyLabel}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        {footer}
      </div>
    </DataTableContext.Provider>
  );
}

function getColumnId<TData, TValue>(column: ColumnDef<TData, TValue>): string {
  if (column.id) return column.id;
  if ("accessorKey" in column && column.accessorKey) return String(column.accessorKey).replaceAll(".", "_");
  return "";
}

function optionalColumnOptions<TData>(table: TanStackTable<TData>): DataTableColumnOption[] {
  return table
    .getAllLeafColumns()
    .filter((column) => column.columnDef.meta?.optional)
    .map((column) => ({
      id: column.id,
      label: column.columnDef.meta?.label ?? column.id,
      visible: column.getIsVisible(),
    }));
}
