"use client";

import { useTranslations } from "next-intl";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { PriceScale } from "@/lib/display-settings";
import { cn } from "@/lib/utils";

type TokenScaleProps = {
  scale: PriceScale;
  className?: string;
};

export function TokenScale({ scale, className }: TokenScaleProps) {
  const t = useTranslations("Units");
  const isPerToken = scale === "PER_TOKEN";
  const label = t(isPerToken ? "oneTokenShort" : "oneMillionTokensShort");
  const description = t(isPerToken ? "oneToken" : "oneMillionTokens");

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            tabIndex={0}
            aria-label={description}
            className={cn(
              "inline-flex cursor-help items-center whitespace-nowrap rounded-sm underline decoration-dotted underline-offset-4 outline-none focus-visible:ring-2 focus-visible:ring-ring",
              className,
            )}
          >
            {label}
          </span>
        </TooltipTrigger>
        <TooltipContent sideOffset={6}>{description}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

