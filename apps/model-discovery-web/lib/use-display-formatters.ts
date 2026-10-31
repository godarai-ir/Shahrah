"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useDisplaySettings } from "@/components/display-settings-provider";
import {
  formatDisplayDate,
  formatDisplayPrice,
  formatDisplayPriceAmount,
} from "@/lib/display-settings";

export function useDisplayFormatters() {
  const t = useTranslations("Units");
  const { settings, exchangeRate, exchangeRateAvailability } = useDisplaySettings();

  const priceLabels = useMemo(
    () => ({
      currencies: {
        USD: t("usd"),
        IRR: t("irr"),
        TOMAN: t("toman"),
      },
      perToken: t("perToken"),
      perMillionTokens: t("perMillionTokens"),
    }),
    [t],
  );

  return {
    settings,
    exchangeRate,
    exchangeRateAvailability,
    formatDate: (value: string | null | undefined) => formatDisplayDate(value, settings),
    formatPrice: (value: number | null | undefined) =>
      formatDisplayPrice(value, settings, priceLabels, exchangeRate?.rate),
    formatPriceAmount: (value: number | null | undefined) =>
      formatDisplayPriceAmount(value, settings, priceLabels, exchangeRate?.rate),
    formatUnitPriceAmount: (value: number | null | undefined) =>
      formatDisplayPriceAmount(value, settings, priceLabels, exchangeRate?.rate, true),
  };
}
