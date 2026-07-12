export type DateTimeInput = string | number | Date | null | undefined;

export type FormatDateTimeOptions = {
  locale?: string;
  fallback?: string;
  format?: Intl.DateTimeFormatOptions;
};

const DEFAULT_FORMAT: Intl.DateTimeFormatOptions = { dateStyle: "medium", timeStyle: "short" };

export function parseDateTime(value: DateTimeInput): Date | null {
  if (value === null || value === undefined || value === "") return null;
  const date = value instanceof Date ? value : new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function resolveDateLocale(locale = "en"): string {
  return locale.toLowerCase().startsWith("fa") ? "fa-IR-u-ca-persian" : locale;
}

export function formatDateTime(value: DateTimeInput, { locale = "en", fallback = "-", format = DEFAULT_FORMAT }: FormatDateTimeOptions = {}): string {
  const date = parseDateTime(value);
  if (!date) return fallback;
  return new Intl.DateTimeFormat(resolveDateLocale(locale), format).format(date);
}

export function toDateTimeLocalValue(value: DateTimeInput): string {
  const date = parseDateTime(value);
  if (!date) return "";
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function fromDateTimeLocalValue(value: string): string {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "" : date.toISOString();
}
