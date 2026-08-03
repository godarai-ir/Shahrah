import { NextResponse } from 'next/server.js';
import { parseExchangeRateQuote } from '../../../lib/exchange-rate.ts';

const UPSTREAM_PATH = '/api/v1/exchange-rates/current';
const RESPONSE_CACHE_CONTROL = 'public, max-age=60, stale-while-revalidate=600';

export async function GET() {
  const baseURL =
    process.env.SHAHRAH_CONTROL_PLANE_URL ||
    (process.env.NODE_ENV === 'production' ? 'http://control-plane:4000' : 'http://localhost:4000');

  try {
    const response = await fetch(new URL(UPSTREAM_PATH, ensureTrailingSlash(baseURL)), {
      headers: { accept: 'application/json' },
      next: { revalidate: 60 },
      signal: AbortSignal.timeout(3_000),
    });

    if (!response.ok) {
      return unavailableResponse();
    }

    const quote = parseExchangeRateQuote(await response.json());
    if (!quote) {
      return unavailableResponse();
    }

    return NextResponse.json(quote, {
      headers: { 'cache-control': RESPONSE_CACHE_CONTROL },
    });
  } catch {
    return unavailableResponse();
  }
}

function unavailableResponse() {
  return NextResponse.json(
    { error: 'Current exchange rate is unavailable' },
    { status: 503, headers: { 'cache-control': 'no-store' } },
  );
}

function ensureTrailingSlash(value: string) {
  return value.endsWith('/') ? value : `${value}/`;
}
