import { Injectable } from '@nestjs/common';
import {
  ExchangeRatePayload,
  MockExchangeRateProvider,
} from './exchange-rate.provider';

@Injectable()
export class ExchangeRateService {
  private cache: ExchangeRatePayload | null = null;

  constructor(private readonly provider: MockExchangeRateProvider) {}

  async getCachedOrFetch(): Promise<ExchangeRatePayload> {
    if (this.cache) return this.cache;
    return this.forceRefresh();
  }

  readCached(): ExchangeRatePayload | null {
    return this.cache;
  }

  async forceRefresh(): Promise<ExchangeRatePayload> {
    const value = await this.provider.fetchUsdToIrr();
    this.cache = value;
    return value;
  }
}
