"use client";

import { TokenScale } from "@/components/catalog/token-scale";
import { useDisplayFormatters } from "@/lib/use-display-formatters";
import { cn } from "@/lib/utils";
import { formatRateUnit, type PricingRate } from "@/lib/pricing";

type DisplayPriceProps = {
  value?: number | null;
  rate?: PricingRate | null;
  className?: string;
};

export function DisplayPrice({ value, rate, className }: DisplayPriceProps) {
  const { settings, formatPriceAmount } = useDisplayFormatters();

  const amount = rate ? Number(rate.amount) : value;
  if (amount === null || amount === undefined || !Number.isFinite(amount)) return <>-</>;

  return (
    <span className={cn("inline-flex flex-wrap items-baseline gap-1", className)}>
      <bdi dir="auto">{formatPriceAmount(amount)}</bdi>
      <span aria-hidden="true">،</span>
      {rate ? (
        <bdi dir="ltr">/ {formatRateUnit(rate)}</bdi>
      ) : (
        <TokenScale scale={settings.priceScale} />
      )}
    </span>
  );
}
