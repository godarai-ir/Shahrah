import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

export type FormSheetActionsProps = {
  primary: ReactNode;
  secondary?: ReactNode;
  auxiliary?: ReactNode;
  destructive?: ReactNode;
  className?: string;
};

export function FormSheetActions({ primary, secondary, auxiliary, destructive, className }: FormSheetActionsProps) {
  const hasPeripheralActions = Boolean(auxiliary || destructive);

  return (
    <div className={cn("flex w-full flex-col gap-2 sm:flex-row sm:items-center sm:justify-between", className)}>
      <div className={cn("grid min-w-0 flex-1 gap-2 [&>*]:w-full", secondary ? "grid-cols-2" : "grid-cols-1")}>
        {secondary}
        {primary}
      </div>
      {hasPeripheralActions ? (
        <div className="flex shrink-0 flex-wrap items-center gap-2">
          {auxiliary}
          {destructive}
        </div>
      ) : null}
    </div>
  );
}
