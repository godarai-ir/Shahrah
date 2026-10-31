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
  themeMode: 'SYSTEM',
};

const EMPTY = '-';

export function formatDisplayPrice(
  value: number | null | undefined,
  settings: DisplaySettings,
  labels: DisplayPriceLabels,
  usdToIrr?: number | null,
): string {
  const amount = formatDisplayPriceAmount(value, settings, labels, usdToIrr);
  if (amount === EMPTY) return EMPTY;

  const unitLabel = settings.priceScale === 'PER_TOKEN' ? labels.perToken : labels.perMillionTokens;
  return `${amount}، ${unitLabel}`;
}

export function formatDisplayPriceAmount(
  value: number | null | undefined,
  settings: DisplaySettings,
  labels: DisplayPriceLabels,
  usdToIrr?: number | null,
  perUnit = false,
): string {
  if (value === null || value === undefined) return EMPTY;

  let transformed = value;
  // A per-unit price (image, second) is already final, so the token scale does not apply.
  if (!perUnit && settings.priceScale === 'PER_TOKEN') {
    transformed = transformed / 1_000_000;
  }

  if (settings.currencyUnit !== 'USD') {
    if (!usdToIrr || !Number.isFinite(usdToIrr) || usdToIrr <= 0) return EMPTY;
    transformed = transformed * usdToIrr;
  }

  if (settings.currencyUnit === 'TOMAN') {
    transformed /= 10;
  }

  let fractionDigits = perUnit ? 4 : settings.priceScale === 'PER_TOKEN' ? 8 : 2;
  if (settings.currencyUnit === 'IRR') {
    // Rial is the smallest currency unit and cannot have a fractional part.
    transformed = Math.ceil(transformed);
    fractionDigits = 0;
  } else if (settings.currencyUnit === 'TOMAN') {
    // Round the toman amount up, then round the last three digits (hundreds) up too.
    transformed = Math.ceil(Math.ceil(transformed) / 1000) * 1000;
    fractionDigits = 0;
  }

  const formatted = new Intl.NumberFormat('fa-IR', {
    maximumFractionDigits: fractionDigits,
  }).format(transformed);

  return `${formatted} ${labels.currencies[settings.currencyUnit]}`;
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
