import assert from 'node:assert/strict';
import test from 'node:test';
import { GET } from '../app/api/exchange-rate/route.ts';

const validQuote = {
  base: 'USD',
  quote: 'IRR',
  rate: 1_883_500,
  source: 'nobitex:USDTIRT',
  marketSymbol: 'USDTIRT',
  pricingBasis: 'USDT_AS_USD_PROXY',
  fetchedAt: '2026-08-03T10:00:00.000Z',
  status: 'fresh',
};

test('exchange-rate BFF returns only a validated upstream quote', async (t) => {
  const originalFetch = globalThis.fetch;
  const originalBaseURL = process.env.SHAHRAH_CONTROL_PLANE_URL;
  t.after(() => {
    globalThis.fetch = originalFetch;
    if (originalBaseURL === undefined) {
      delete process.env.SHAHRAH_CONTROL_PLANE_URL;
    } else {
      process.env.SHAHRAH_CONTROL_PLANE_URL = originalBaseURL;
    }
  });

  process.env.SHAHRAH_CONTROL_PLANE_URL = 'http://control-plane.test:4000';
  globalThis.fetch = (async (input) => {
    assert.equal(String(input), 'http://control-plane.test:4000/api/v1/exchange-rates/current');
    return Response.json(validQuote);
  }) as typeof fetch;

  const response = await GET();
  assert.equal(response.status, 200);
  assert.equal(response.headers.get('cache-control'), 'public, max-age=60, stale-while-revalidate=600');
  assert.deepEqual(await response.json(), validQuote);
});

test('exchange-rate BFF fails closed for an invalid upstream payload', async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => {
    globalThis.fetch = originalFetch;
  });

  globalThis.fetch = (async () => Response.json({ rate: 650_000 })) as typeof fetch;

  const response = await GET();
  assert.equal(response.status, 503);
  assert.equal(response.headers.get('cache-control'), 'no-store');
});
