import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';

export interface ExchangeRatePayload {
  usdToIrr: number;
  fetchedAt: string;
  source: string;
}

@Injectable()
export class MockExchangeRateProvider {
  constructor(private readonly config: ConfigService) {}

  async fetchUsdToIrr(): Promise<ExchangeRatePayload> {
    const now = new Date().toISOString();
    return {
      usdToIrr: 1850000,
      fetchedAt: now,
      source: this.config.get<string>(
        'EXCHANGE_RATE_API_URL',
        'mock://exchange-rate',
      ),
    };
  }
}
