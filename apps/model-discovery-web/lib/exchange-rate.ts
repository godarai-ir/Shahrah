export type ExchangeRateStatus = 'fresh' | 'stale';

export type ExchangeRateQuote = {
  base: 'USD';
  quote: 'IRR';
  rate: number;
  source: string;
  marketSymbol: string;
  pricingBasis: 'USDT_AS_USD_PROXY';
  fetchedAt: string;
  status: ExchangeRateStatus;
};

export type ExchangeRateAvailability = 'loading' | ExchangeRateStatus | 'unavailable';

export const CURRENT_EXCHANGE_RATE_PATH = '/catalog/api/exchange-rate';
const EXCHANGE_RATE_STALE_AFTER_MS = 15 * 60 * 1000;

export function parseExchangeRateQuote(value: unknown): ExchangeRateQuote | null {
  if (!isRecord(value)) return null;

  const fetchedAt = typeof value.fetchedAt === 'string' ? value.fetchedAt : '';
  const fetchedAtMs = Date.parse(fetchedAt);
  const rate = value.rate;

  if (
    value.base !== 'USD' ||
    value.quote !== 'IRR' ||
    typeof rate !== 'number' ||
    !Number.isFinite(rate) ||
    rate <= 0 ||
    typeof value.source !== 'string' ||
    !value.source.toLowerCase().startsWith('nobitex:') ||
    typeof value.marketSymbol !== 'string' ||
    !value.marketSymbol.startsWith('USDT') ||
    value.pricingBasis !== 'USDT_AS_USD_PROXY' ||
    !Number.isFinite(fetchedAtMs) ||
    (value.status !== 'fresh' && value.status !== 'stale')
  ) {
    return null;
  }

  return {
    base: 'USD',
    quote: 'IRR',
    rate,
    source: value.source,
    marketSymbol: value.marketSymbol,
    pricingBasis: 'USDT_AS_USD_PROXY',
    fetchedAt,
    status: value.status,
  };
}

export function resolveExchangeRateAvailability(
  quote: ExchangeRateQuote | null,
  pending: boolean,
  now = Date.now(),
): ExchangeRateAvailability {
  if (!quote) return pending ? 'loading' : 'unavailable';
  if (quote.status === 'stale') return 'stale';

  return now - Date.parse(quote.fetchedAt) > EXCHANGE_RATE_STALE_AFTER_MS ? 'stale' : 'fresh';
}

export async function fetchCurrentExchangeRate(signal?: AbortSignal): Promise<ExchangeRateQuote> {
  const response = await fetch(CURRENT_EXCHANGE_RATE_PATH, {
    headers: { accept: 'application/json' },
    cache: 'no-store',
    signal,
  });

  if (!response.ok) {
    throw new Error(`Exchange rate API returned ${response.status}`);
  }

  const quote = parseExchangeRateQuote(await response.json());
  if (!quote) throw new Error('Exchange rate API returned an invalid payload');

  return quote;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
