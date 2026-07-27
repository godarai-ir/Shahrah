import { evaluatePublicationReadiness } from './publication-policy';

const profile = (unit: string) => ({
  region: 'global',
  currency: 'USD',
  rates: [
    {
      key: `output:${unit}`,
      usage: 'output',
      unit,
      amount: '1',
      quantity: unit === 'image' ? 1 : 1_000_000,
      serviceTier: 'standard',
    },
  ],
});

describe('evaluatePublicationReadiness', () => {
  it.each(['image', 'image_token', 'token'])(
    'accepts output/%s image pricing',
    (unit) => {
      expect(
        evaluatePublicationReadiness({
          mode: ['image'],
          pricingStatus: 'priced',
          pricingProfiles: [profile(unit)],
        }),
      ).toMatchObject({
        eligible: true,
        strategies: [{ region: 'global', unit }],
      });
    },
  );

  it('rejects unknown and unsupported image pricing', () => {
    const unknown = evaluatePublicationReadiness({
      mode: ['image'],
      pricingStatus: 'unknown',
      pricingProfiles: [],
    });
    expect(unknown.blockers[0].code).toBe('pricing_unknown');

    const unsupported = evaluatePublicationReadiness({
      mode: ['image'],
      pricingStatus: 'priced',
      pricingProfiles: [profile('audio_token')],
    });
    expect(unsupported.blockers).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: 'unsupported_output_strategy' }),
      ]),
    );
  });

  it('uses image before image_token and token', () => {
    const input = profile('token');
    input.rates.push(profile('image_token').rates[0], profile('image').rates[0]);
    const readiness = evaluatePublicationReadiness({
      mode: ['image'],
      pricingStatus: 'priced',
      pricingProfiles: [input],
    });
    expect(readiness.strategies[0].unit).toBe('image');
  });

  it('accepts explicit free image models without rates', () => {
    expect(
      evaluatePublicationReadiness({
        mode: ['image'],
        pricingStatus: 'free',
        pricingProfiles: [],
      }).eligible,
    ).toBe(true);
  });
});
