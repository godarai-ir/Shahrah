"use client";

import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import type { CompactDisplayValue } from "@/lib/compact-number";
import { cn } from "@/lib/utils";

type CompactValueProps = {
  display: CompactDisplayValue;
  className?: string;
  tooltipClassName?: string;
  dir?: "auto" | "ltr" | "rtl";
};

export function CompactValue({
  display,
  className,
  tooltipClassName,
  dir = "auto",
}: CompactValueProps) {
  const content = (
    <span
      className={cn("inline-flex min-w-0 max-w-full items-center truncate tabular-nums", className)}
      dir={dir}
      title={display.fullValue}
    >
      {display.value}
    </span>
  );

  if (!display.isCompact) {
    return content;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>{content}</TooltipTrigger>
        <TooltipContent className={cn("font-mono tabular-nums", tooltipClassName)} dir="auto">
          {display.fullValue}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
