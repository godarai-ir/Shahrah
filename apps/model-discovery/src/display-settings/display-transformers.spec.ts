import { convertPriceScale, convertUsdToCurrency, formatDateByCalendar } from './display-transformers';
import { DisplaySettings } from './display-settings.types';

describe('display-transformers', () => {
  const base: DisplaySettings = {
    currencyUnit: 'USD',
    dateCalendar: 'GREGORIAN',
    priceScale: 'PER_MILLION_TOKENS',
  };

  it('converts usd to irr and toman', () => {
    expect(convertUsdToCurrency(2, { ...base, currencyUnit: 'IRR' }, 650000)).toBe(1300000);
    expect(convertUsdToCurrency(2, { ...base, currencyUnit: 'TOMAN' }, 650000)).toBe(130000);
    expect(convertUsdToCurrency(2, base, 650000)).toBe(2);
  });

  it('converts per-million to per-token', () => {
    expect(convertPriceScale(1, { ...base, priceScale: 'PER_TOKEN' })).toBe(0.000001);
    expect(convertPriceScale(1, base)).toBe(1);
  });

  it('formats date using selected calendar', () => {
    const iso = '2026-05-18T10:00:00.000Z';
    const gregorian = formatDateByCalendar(iso, { ...base, dateCalendar: 'GREGORIAN' });
    const shamsi = formatDateByCalendar(iso, { ...base, dateCalendar: 'SHAMSI' });

    expect(gregorian).not.toBe('-');
    expect(shamsi).not.toBe('-');
    expect(shamsi).not.toBe(gregorian);
  });
});
