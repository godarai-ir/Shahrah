import { plainToInstance } from 'class-transformer';
import { validateSync } from 'class-validator';
import {
  resolvePricingTier,
  UpdateModelPricingDto,
} from '../catalog/dto/upsert-model.dto';
import { serializePricingProfile } from './pricing-response';

describe('pricing response round-trip', () => {
  it('preserves nested tier fields when an admin response is submitted again', () => {
    const pricingProfile = serializePricingProfile({
      region: { slug: 'global' },
      currency: 'USD',
      effectiveFrom: new Date('2026-01-01T00:00:00.000Z'),
      effectiveTo: null,
      sourceUrl: null,
      rates: [
        {
          rateKey: 'input:token:standard:resolution=default:tier=200000',
          usage: 'input',
          unit: 'token',
          amount: '5',
          quantity: 1_000_000,
          serviceTier: 'standard',
          resolution: null,
          tierFrom: 200_000,
          tierTo: 1_000_000,
          tierMode: 'cumulative',
          qualifiers: {},
        },
      ],
    });
    const dto = plainToInstance(UpdateModelPricingDto, {
      pricingStatus: 'priced',
      pricingProfiles: [pricingProfile],
    });

    expect(validateSync(dto)).toEqual([]);
    expect(resolvePricingTier(dto.pricingProfiles![0].rates[0])).toEqual({
      from: 200_000,
      to: 1_000_000,
      mode: 'cumulative',
    });
  });

  it('keeps accepting the legacy flat tier fields', () => {
    const dto = plainToInstance(UpdateModelPricingDto, {
      pricingStatus: 'priced',
      pricingProfiles: [
        {
          region: 'global',
          currency: 'USD',
          rates: [
            {
              usage: 'input',
              unit: 'token',
              amount: '5',
              quantity: 1_000_000,
              tierFrom: 200_000,
              tierTo: 1_000_000,
              tierMode: 'cumulative',
            },
          ],
        },
      ],
    });

    expect(validateSync(dto)).toEqual([]);
    expect(resolvePricingTier(dto.pricingProfiles![0].rates[0])).toEqual({
      from: 200_000,
      to: 1_000_000,
      mode: 'cumulative',
    });
  });
});
