import { parsePricingProfiles } from './pricing-parser';

describe('parsePricingProfiles', () => {
  it('preserves region and expands video resolutions', () => {
    const result = parsePricingProfiles([
      {
        output_cost_per_second_540p: 0.75,
        output_cost_per_second_720p: 1.5,
        region: 'us-west-2',
      },
    ]);

    expect(result.diagnostics).toEqual([]);
    expect(result.status).toBe('priced');
    expect(result.profiles).toEqual([
      expect.objectContaining({
        regionSlug: 'us-west-2',
        rates: [
          expect.objectContaining({
            usage: 'output',
            unit: 'second',
            amount: '0.75',
            quantity: 1,
            resolution: '540p',
          }),
          expect.objectContaining({
            usage: 'output',
            unit: 'second',
            amount: '1.5',
            quantity: 1,
            resolution: '720p',
          }),
        ],
      }),
    ]);
  });

  it('keeps a generic per-second rate resolution-free', () => {
    const result = parsePricingProfiles([
      { output_cost_per_second: 0.08, region: '*' },
    ]);

    expect(result.profiles[0].regionSlug).toBe('global');
    expect(result.profiles[0].rates[0]).toEqual(
      expect.objectContaining({
        usage: 'output',
        unit: 'second',
        amount: '0.08',
        resolution: undefined,
      }),
    );
  });

  it('normalizes token prices to one million units', () => {
    const result = parsePricingProfiles([
      {
        input_cost_per_token: 0.000003,
        output_cost_per_token: 0.000015,
        cache_read_input_token_cost: 0.0000003,
      },
    ]);

    expect(result.profiles[0].rates).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ usage: 'input', amount: '3' }),
        expect.objectContaining({ usage: 'output', amount: '15' }),
        expect.objectContaining({ usage: 'cache_read', amount: '0.3' }),
      ]),
    );
  });

  it('parses batch, priority, and cumulative tiers', () => {
    const result = parsePricingProfiles([
      {
        input_cost_per_token: 0.000001,
        input_cost_per_token_batches: 0.0000005,
        priority_pricing: { input_cost_per_token: 0.000002 },
        tiered_pricing: {
          input: [{ cost_per_token: 0.000003, from: 200000 }],
          pricing_mode: 'cumulative',
        },
      },
    ]);

    expect(result.diagnostics).toEqual([]);
    expect(result.profiles[0].rates).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ serviceTier: 'batch', amount: '0.5' }),
        expect.objectContaining({ serviceTier: 'priority', amount: '2' }),
        expect.objectContaining({
          tierFrom: 200000,
          tierMode: 'cumulative',
          amount: '3',
        }),
      ]),
    );
  });

  it('does not infer free from zero or unknown pricing', () => {
    const zero = parsePricingProfiles([{ input_cost_per_token: 0 }]);
    const unknown = parsePricingProfiles([{ future_cost_metric: 1 }]);

    expect(zero.status).toBe('unknown');
    expect(unknown.status).toBe('unknown');
    expect(unknown.diagnostics).toEqual([
      expect.objectContaining({ code: 'unknown_pricing_key' }),
      expect.objectContaining({ code: 'empty_profile' }),
    ]);
  });
});
