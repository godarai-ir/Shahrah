export type CurrencyUnit = 'USD' | 'IRR' | 'TOMAN';
export type DateCalendar = 'GREGORIAN' | 'SHAMSI';
export type PriceScale = 'PER_MILLION_TOKENS' | 'PER_TOKEN';
export type ThemeMode = 'LIGHT' | 'DARK' | 'SYSTEM';

export type DisplaySettings = {
  currencyUnit: CurrencyUnit;
  dateCalendar: DateCalendar;
  priceScale: PriceScale;
  themeMode: ThemeMode;
};

export type DisplayPriceLabels = {
  currencies: Record<CurrencyUnit, string>;
  perToken: string;
  perMillionTokens: string;
};

export const DISPLAY_SETTINGS_STORAGE_KEY = 'model_discovery_web_display_settings_v1';

export const DISPLAY_SETTINGS_DEFAULTS: DisplaySettings = {
  dateCalendar: 'GREGORIAN',
  currencyUnit: 'USD',
  priceScale: 'PER_MILLION_TOKENS',
  themeMode: 'LIGHT',
};

const EMPTY = '-';
const DEFAULT_USD_TO_IRR = 650000;

export function formatDisplayPrice(
  value: number | null | undefined,
  settings: DisplaySettings,
  labels: DisplayPriceLabels,
): string {
  if (value === null || value === undefined) return EMPTY;

  let transformed = value;
  if (settings.priceScale === 'PER_TOKEN') {
    transformed = transformed / 1_000_000;
  }

  if (settings.currencyUnit === 'IRR') {
    transformed = transformed * DEFAULT_USD_TO_IRR;
  } else if (settings.currencyUnit === 'TOMAN') {
    transformed = (transformed * DEFAULT_USD_TO_IRR) / 10;
  }

  const fractionDigits = settings.priceScale === 'PER_TOKEN' ? 8 : 2;
  const formatted = new Intl.NumberFormat('fa-IR', {
    maximumFractionDigits: fractionDigits,
  }).format(transformed);

  const unitLabel = settings.priceScale === 'PER_TOKEN' ? labels.perToken : labels.perMillionTokens;
  return `${formatted} ${labels.currencies[settings.currencyUnit]}، ${unitLabel}`;
}

export function formatDisplayDate(
  value: string | null | undefined,
  settings: DisplaySettings,
): string {
  if (!value) return EMPTY;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return EMPTY;

  const locale = settings.dateCalendar === 'SHAMSI' ? 'fa-IR-u-ca-persian' : 'fa-IR-u-ca-gregory';
  return new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}
