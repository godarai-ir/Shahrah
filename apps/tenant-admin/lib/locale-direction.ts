export type LocaleDirection = "ltr" | "rtl";

export function getLocaleDirection(locale: string): LocaleDirection {
  const language = locale.trim().toLowerCase().split(/[-_]/, 1)[0];

  return language === "fa" ? "rtl" : "ltr";
}
