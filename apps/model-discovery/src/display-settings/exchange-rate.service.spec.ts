import { ExchangeRateService } from './exchange-rate.service';
import { MockExchangeRateProvider } from './exchange-rate.provider';

describe('ExchangeRateService', () => {
  it('caches latest response until force refresh', async () => {
    let calls = 0;
    const provider: Pick<MockExchangeRateProvider, 'fetchUsdToIrr'> = {
      fetchUsdToIrr: jest.fn(async () => {
        calls += 1;
        return {
          usdToIrr: 650000 + calls,
          fetchedAt: new Date(`2026-05-18T10:00:0${calls}Z`).toISOString(),
          source: 'mock://exchange-rate',
        };
      }),
    };

    const service = new ExchangeRateService(provider as MockExchangeRateProvider);

    const first = await service.getCachedOrFetch();
    const second = await service.getCachedOrFetch();

    expect(first.usdToIrr).toBe(second.usdToIrr);
    expect(provider.fetchUsdToIrr).toHaveBeenCalledTimes(1);

    const refreshed = await service.forceRefresh();
    expect(refreshed.usdToIrr).not.toBe(first.usdToIrr);
    expect(provider.fetchUsdToIrr).toHaveBeenCalledTimes(2);
  });
});
