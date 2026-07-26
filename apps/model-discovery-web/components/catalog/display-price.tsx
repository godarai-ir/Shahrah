"use client";

import { TokenScale } from "@/components/catalog/token-scale";
import { useDisplayFormatters } from "@/lib/use-display-formatters";
import { cn } from "@/lib/utils";

type DisplayPriceProps = {
  value?: number | null;
  className?: string;
};

export function DisplayPrice({ value, className }: DisplayPriceProps) {
  const { settings, formatPriceAmount } = useDisplayFormatters();

  if (value === null || value === undefined) return <>-</>;

  return (
    <span className={cn("inline-flex flex-wrap items-baseline gap-1", className)}>
      <bdi dir="auto">{formatPriceAmount(value)}</bdi>
      <span aria-hidden="true">،</span>
      <TokenScale scale={settings.priceScale} />
    </span>
  );
}

