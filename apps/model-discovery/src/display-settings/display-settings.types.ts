export type CurrencyUnit = 'USD' | 'IRR' | 'TOMAN';
export type DateCalendar = 'GREGORIAN' | 'SHAMSI';
export type PriceScale = 'PER_MILLION_TOKENS' | 'PER_TOKEN';

export interface DisplaySettings {
  currencyUnit: CurrencyUnit;
  dateCalendar: DateCalendar;
  priceScale: PriceScale;
}

export interface ExchangeRateSnapshot {
  usdToIrr: number;
  fetchedAt: string;
  source: string;
}

export interface DisplaySettingsResponse extends DisplaySettings {
  exchangeRate: ExchangeRateSnapshot;
}
