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
      className={cn("@container flex flex-col gap-3 text-sm text-muted-foreground @md:flex-row @md:items-center @md:justify-between", className)}
    >
      <p>{summary}</p>
      <nav aria-label={label} className="flex flex-col gap-2 @sm:flex-row @sm:flex-wrap @sm:items-center">
        {pageSizeLabel ? <span>{pageSizeLabel}</span> : null}
        <span className="flex items-center justify-between gap-2 @sm:justify-start">
          {pageSize !== undefined ? <Badge variant="outline">{pageSize}</Badge> : null}
          <Button type="button" variant="outline" size="sm" onClick={onPrevious} disabled={disabled || !canPrevious}>
            {previousLabel}
          </Button>
          {page !== undefined ? <Badge variant="secondary">{page}</Badge> : null}
          <Button type="button" variant="outline" size="sm" onClick={onNext} disabled={disabled || !canNext}>
            {nextLabel}
          </Button>
        </span>
      </nav>
    </div>
  );
}
