import assert from 'node:assert/strict';
import test from 'node:test';
import {
  DISPLAY_SETTINGS_DEFAULTS,
  formatDisplayPriceAmount,
} from './display-settings.ts';
import type { DisplayPriceLabels } from './display-settings.ts';
import { parseExchangeRateQuote, resolveExchangeRateAvailability } from './exchange-rate.ts';

const labels: DisplayPriceLabels = {
  currencies: { USD: 'دلار', IRR: 'ریال', TOMAN: 'تومان' },
  perToken: 'هر توکن',
  perMillionTokens: 'یک میلیون توکن',
};

test('dynamic exchange rate only changes IRR and TOMAN prices', () => {
  const usd = { ...DISPLAY_SETTINGS_DEFAULTS, currencyUnit: 'USD' as const };
  const irr = { ...DISPLAY_SETTINGS_DEFAULTS, currencyUnit: 'IRR' as const };
  const toman = { ...DISPLAY_SETTINGS_DEFAULTS, currencyUnit: 'TOMAN' as const };

  assert.equal(formatDisplayPriceAmount(2, usd, labels, 1_800_000), '۲ دلار');
  assert.equal(formatDisplayPriceAmount(2, usd, labels, 1_900_000), '۲ دلار');
  assert.equal(formatDisplayPriceAmount(2, irr, labels, 1_800_000), '۳٬۶۰۰٬۰۰۰ ریال');
  assert.equal(formatDisplayPriceAmount(2, irr, labels, 1_900_000), '۳٬۸۰۰٬۰۰۰ ریال');
  assert.equal(formatDisplayPriceAmount(2, toman, labels, 1_800_000), '۳۶۰٬۰۰۰ تومان');
});

test('IRR has no fractional digits and TOMAN rounds up to the nearest thousand', () => {
  const irr = { ...DISPLAY_SETTINGS_DEFAULTS, currencyUnit: 'IRR' as const };
  const toman = { ...DISPLAY_SETTINGS_DEFAULTS, currencyUnit: 'TOMAN' as const };

  assert.equal(formatDisplayPriceAmount(1.2540345, irr, labels, 1_800_000), '۲٬۲۵۷٬۲۶۳ ریال');
  assert.equal(formatDisplayPriceAmount(1.2540345, toman, labels, 1_800_000), '۲۲۶٬۰۰۰ تومان');
});

test('non-USD prices do not silently use a fixed fallback', () => {
  const irr = { ...DISPLAY_SETTINGS_DEFAULTS, currencyUnit: 'IRR' as const };
  const toman = { ...DISPLAY_SETTINGS_DEFAULTS, currencyUnit: 'TOMAN' as const };

  assert.equal(formatDisplayPriceAmount(2, irr, labels, null), '-');
  assert.equal(formatDisplayPriceAmount(2, toman, labels, undefined), '-');
});

test('exchange-rate payload validation and client-side staleness are explicit', () => {
  const quote = parseExchangeRateQuote({
    base: 'USD',
    quote: 'IRR',
    rate: 1_883_500,
    source: 'nobitex:USDTIRT',
    marketSymbol: 'USDTIRT',
    pricingBasis: 'USDT_AS_USD_PROXY',
    fetchedAt: '2026-08-03T10:00:00.000Z',
    status: 'fresh',
  });

  assert.ok(quote);
  assert.equal(resolveExchangeRateAvailability(quote, false, Date.parse('2026-08-03T10:10:00.000Z')), 'fresh');
  assert.equal(resolveExchangeRateAvailability(quote, false, Date.parse('2026-08-03T10:16:00.000Z')), 'stale');
  assert.equal(parseExchangeRateQuote({ rate: 1_883_500 }), null);
  assert.equal(resolveExchangeRateAvailability(null, false), 'unavailable');
});
