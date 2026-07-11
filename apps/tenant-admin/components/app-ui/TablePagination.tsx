"use client";

import * as React from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type TablePaginationProps = {
  label: string;
  summary: React.ReactNode;
  previousLabel: React.ReactNode;
  nextLabel: React.ReactNode;
  onPrevious: () => void;
  onNext: () => void;
  canPrevious: boolean;
  canNext: boolean;
  disabled?: boolean;
  page?: React.ReactNode;
  pageSize?: React.ReactNode;
  pageSizeLabel?: React.ReactNode;
  className?: string;
};

export function TablePagination({
  label,
  summary,
  previousLabel,
  nextLabel,
  onPrevious,
  onNext,
  canPrevious,
  canNext,
  disabled = false,
  page,
  pageSize,
  pageSizeLabel,
  className,
}: TablePaginationProps) {
  return (
    <div
      data-slot="table-pagination"
      className={cn("flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground", className)}
    >
      <p>{summary}</p>
      <nav aria-label={label} className="flex flex-wrap items-center gap-2">
        {pageSizeLabel ? <span>{pageSizeLabel}</span> : null}
        {pageSize !== undefined ? <Badge variant="outline">{pageSize}</Badge> : null}
        <Button type="button" variant="outline" size="sm" onClick={onPrevious} disabled={disabled || !canPrevious}>
          {previousLabel}
        </Button>
        {page !== undefined ? <Badge variant="secondary">{page}</Badge> : null}
        <Button type="button" variant="outline" size="sm" onClick={onNext} disabled={disabled || !canNext}>
          {nextLabel}
        </Button>
      </nav>
    </div>
  );
}
