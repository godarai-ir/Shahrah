"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useDisplaySettings } from "@/components/display-settings-provider";
import { formatDisplayDate, formatDisplayPrice } from "@/lib/display-settings";

export function useDisplayFormatters() {
  const t = useTranslations("Units");
  const { settings } = useDisplaySettings();

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
    formatDate: (value: string | null | undefined) => formatDisplayDate(value, settings),
    formatPrice: (value: number | null | undefined) => formatDisplayPrice(value, settings, priceLabels),
  };
}
