"use client";

import { useLocale } from "next-intl";
import { useMemo } from "react";

import { getCalendarLocale, getIntlLocale, resolveLocale, type AppLocale } from "@/lib/locale";

/** Active app locale, narrowed to a supported value. */
export function useAppLocale(): AppLocale {
  return resolveLocale(useLocale());
}

/** BCP 47 tag for number and currency formatting in the active locale. */
export function useIntlLocale(): string {
  return getIntlLocale(useAppLocale());
}

/** BCP 47 tag for date formatting, including the Jalali calendar for Persian. */
export function useCalendarLocale(): string {
  return getCalendarLocale(useAppLocale());
}

export function useNumberFormatter(options?: Intl.NumberFormatOptions): Intl.NumberFormat {
  const intlLocale = useIntlLocale();
  const optionsKey = JSON.stringify(options ?? {});

  return useMemo(
    () => new Intl.NumberFormat(intlLocale, JSON.parse(optionsKey) as Intl.NumberFormatOptions),
    [intlLocale, optionsKey],
  );
}

export function useDateTimeFormatter(options?: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
  const calendarLocale = useCalendarLocale();
  const optionsKey = JSON.stringify(options ?? {});

  return useMemo(
    () => new Intl.DateTimeFormat(calendarLocale, JSON.parse(optionsKey) as Intl.DateTimeFormatOptions),
    [calendarLocale, optionsKey],
  );
}
