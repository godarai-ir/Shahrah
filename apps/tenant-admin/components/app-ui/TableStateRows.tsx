import * as React from "react";

import { EmptyState, type EmptyStateProps } from "@/components/app-ui/EmptyState";
import { Skeleton } from "@/components/ui/skeleton";
import { TableCell, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";

export type TableLoadingRowsProps = {
  colSpan: number;
  rowCount?: number;
  label?: string;
  className?: string;
  skeletonClassName?: string;
};

export function TableLoadingRows({ colSpan, rowCount = 4, label = "Loading", className, skeletonClassName }: TableLoadingRowsProps) {
  return Array.from({ length: rowCount }, (_, index) => (
    <TableRow key={index} aria-hidden={index > 0 || undefined}>
      <TableCell colSpan={colSpan} className={className}>
        <Skeleton className={cn("h-10 w-full", skeletonClassName)} />
        {index === 0 ? <span className="sr-only">{label}</span> : null}
      </TableCell>
    </TableRow>
  ));
}

export type TableEmptyRowProps = Omit<EmptyStateProps, "variant"> & {
  colSpan: number;
  cellClassName?: string;
};

export function TableEmptyRow({ colSpan, cellClassName, ...emptyStateProps }: TableEmptyRowProps) {
  return (
    <TableRow>
      <TableCell colSpan={colSpan} className={cn("py-6", cellClassName)}>
        <EmptyState variant="table" {...emptyStateProps} />
      </TableCell>
    </TableRow>
  );
}
