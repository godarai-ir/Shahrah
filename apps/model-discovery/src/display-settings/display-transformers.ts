import { DisplaySettings } from './display-settings.types';

export function convertUsdToCurrency(value: number | null | undefined, settings: DisplaySettings, usdToIrr: number): number | null {
  if (value === null || value === undefined) return null;
  if (settings.currencyUnit === 'USD') return value;
  if (settings.currencyUnit === 'IRR') return value * usdToIrr;
  return (value * usdToIrr) / 10;
}

export function convertPriceScale(value: number | null | undefined, settings: DisplaySettings): number | null {
  if (value === null || value === undefined) return null;
  if (settings.priceScale === 'PER_MILLION_TOKENS') return value;
  return value / 1_000_000;
}

export function formatDateByCalendar(input: string | Date | null | undefined, settings: DisplaySettings): string {
  if (!input) return '-';
  const date = input instanceof Date ? input : new Date(input);
  if (Number.isNaN(date.getTime())) return '-';
  const locale = settings.dateCalendar === 'SHAMSI' ? 'fa-IR-u-ca-persian' : 'en-US';
  return new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeStyle: 'short' }).format(date);
}
