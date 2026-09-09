export const SUPPORTED_LOCALES = ["fa", "en"] as const;

export type AppLocale = (typeof SUPPORTED_LOCALES)[number];

export const DEFAULT_LOCALE: AppLocale = "fa";

/** Cookie inspected by `i18n/request.ts` on every server render. */
export const LOCALE_COOKIE_NAME = "NEXT_LOCALE";

/** One year, so the choice survives normal browser cleanup. */
export const LOCALE_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 365;

export function isSupportedLocale(value: unknown): value is AppLocale {
  return typeof value === "string" && (SUPPORTED_LOCALES as readonly string[]).includes(value);
}

/**
 * Narrow an arbitrary locale-ish value (cookie, header, prop) to a supported
 * locale, matching on the language subtag so `en-US` still resolves to `en`.
 */
export function resolveLocale(value: unknown): AppLocale {
  if (typeof value !== "string") return DEFAULT_LOCALE;
  const language = value.trim().toLowerCase().split(/[-_]/, 1)[0];
  return isSupportedLocale(language) ? language : DEFAULT_LOCALE;
}

/** BCP 47 tag used for `Intl` formatting in the given app locale. */
export function getIntlLocale(locale: string): string {
  return resolveLocale(locale) === "fa" ? "fa-IR" : "en-US";
}

/**
 * `Intl` tag for dates. Persian dates use the Jalali calendar; English keeps
 * the Gregorian calendar so exported and shared values stay unambiguous.
 */
export function getCalendarLocale(locale: string): string {
  return resolveLocale(locale) === "fa" ? "fa-IR-u-ca-persian" : "en-US";
}
