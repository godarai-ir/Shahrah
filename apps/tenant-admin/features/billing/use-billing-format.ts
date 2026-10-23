"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";

import { useAppLocale } from "@/hooks/use-locale-format";
import {
  formatCreditAmount,
  formatDate,
  formatDateOnly,
  formatIRRAmount,
  formatSelectedJalaliMonth,
  formatTomanAmount,
} from "@/features/billing/domain";

export type BillingFormat = {
  /** Rial amount with the localized unit, e.g. "12,000 IRR". */
  irr: (value: number | undefined | null) => string;
  /** Toman amount with the localized unit, converted from rials. */
  toman: (valueIRR: number | undefined | null) => string;
  /** Credit amount with two decimals and the localized unit, e.g. "12.50 credits". */
  credits: (value: string | number | undefined | null) => string;
  /** Rounded rial amount without a unit, for columns that name the unit. */
  amount: (value: number | undefined | null) => string;
  /** Plain localized integer, for counts and pagination. */
  count: (value: number) => string;
  /** Jalali date and time in the workspace time zone. */
  dateTime: (value: string | undefined | null) => string;
  /** Jalali date in the workspace time zone. */
  date: (value: string | undefined | null) => string;
  /** Label for a `YYYY-MM` Jalali month key. */
  jalaliMonth: (monthKey: string) => string;
};

/** Locale-aware billing formatters, including the localized currency units. */
export function useBillingFormat(): BillingFormat {
  const locale = useAppLocale();
  const t = useTranslations("billing");
  const irrUnit = t("units.irr");
  const tomanUnit = t("units.toman");
  const creditUnit = t("units.credits");

  return useMemo<BillingFormat>(
    () => ({
      irr: (value) => `${formatIRRAmount(value, locale)} ${irrUnit}`,
      toman: (valueIRR) => `${formatTomanAmount(valueIRR, locale)} ${tomanUnit}`,
      credits: (value) => `${formatCreditAmount(value, locale)} ${creditUnit}`,
      amount: (value) => formatIRRAmount(value, locale),
      count: (value) => formatIRRAmount(value, locale),
      dateTime: (value) => formatDate(value, locale),
      date: (value) => formatDateOnly(value, locale),
      jalaliMonth: (monthKey) => formatSelectedJalaliMonth(monthKey, locale),
    }),
    [creditUnit, irrUnit, locale, tomanUnit],
  );
}
