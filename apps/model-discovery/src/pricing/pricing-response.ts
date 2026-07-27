import { Decimal } from '@prisma/client-runtime-utils';

type PricingRateRecord = {
  rateKey: string;
  usage: string;
  unit: string;
  amount: Decimal | string | number;
  quantity: number;
  serviceTier: string;
  resolution: string | null;
  tierFrom: number | null;
  tierTo: number | null;
  tierMode: string | null;
  qualifiers: unknown;
};

type PricingProfileRecord = {
  currency: string;
  effectiveFrom: Date | null;
  effectiveTo: Date | null;
  sourceUrl: string | null;
  region: { slug: string };
  rates: PricingRateRecord[];
};

export type PricingProfileResponse = ReturnType<typeof serializePricingProfile>;

export function serializePricingProfile(profile: PricingProfileRecord) {
  return {
    region: profile.region.slug,
    currency: profile.currency,
    effectiveFrom: profile.effectiveFrom?.toISOString() ?? null,
    effectiveTo: profile.effectiveTo?.toISOString() ?? null,
    sourceUrl: profile.sourceUrl,
    rates: profile.rates.map((rate) => ({
      key: rate.rateKey,
      usage: rate.usage,
      unit: rate.unit,
      amount: new Decimal(rate.amount).toString(),
      quantity: rate.quantity,
      serviceTier: rate.serviceTier,
      resolution: rate.resolution,
      tier:
        rate.tierFrom !== null ||
        rate.tierTo !== null ||
        rate.tierMode !== null
          ? {
              from: rate.tierFrom,
              to: rate.tierTo,
              mode: rate.tierMode,
            }
          : null,
      qualifiers:
        rate.qualifiers &&
        typeof rate.qualifiers === 'object' &&
        !Array.isArray(rate.qualifiers)
          ? rate.qualifiers
          : {},
    })),
  };
}

export function serializePricingProfiles(profiles: PricingProfileRecord[]) {
  return profiles.map(serializePricingProfile);
}

export function summarizePricing(
  profiles: PricingProfileResponse[],
  modes: string[],
) {
  const rates = profiles.flatMap((profile) =>
    profile.rates
      .filter(
        (rate) =>
          rate.serviceTier === 'standard' &&
          rate.tier === null &&
          profile.currency === 'USD',
      )
      .map((rate) => ({
        ...rate,
        region: profile.region,
        currency: profile.currency,
      })),
  );
  if (rates.length === 0) return null;

  const preferredSelectors = modes.includes('video')
    ? [
        ['output', 'second'],
        ['input', 'second'],
        ['input', 'request'],
        ['output', 'video_token'],
        ['input', 'video_token'],
        ['input', 'token'],
      ]
    : [
        ['input', 'token'],
        ['output', 'token'],
        ['input', 'request'],
        ['output', 'second'],
      ];
  const selector =
    preferredSelectors.find(([usage, unit]) =>
      rates.some((rate) => rate.usage === usage && rate.unit === unit),
    ) ?? [rates[0].usage, rates[0].unit];
  const selected = rates.filter(
    (rate) => rate.usage === selector[0] && rate.unit === selector[1],
  );
  const amounts = selected.map((rate) => new Decimal(rate.amount));
  const minimum = amounts.reduce((left, right) =>
    Decimal.min(left, right),
  );
  const maximum = amounts.reduce((left, right) =>
    Decimal.max(left, right),
  );

  return {
    usage: selector[0],
    unit: selector[1],
    currency: selected[0].currency,
    quantity: selected[0].quantity,
    minimumAmount: minimum.toString(),
    maximumAmount: maximum.toString(),
    regions: Array.from(new Set(selected.map((rate) => rate.region))).sort(),
    resolutions: Array.from(
      new Set(
        selected
          .map((rate) => rate.resolution)
          .filter((resolution): resolution is string => Boolean(resolution)),
      ),
    ).sort(),
    rateCount: selected.length,
  };
}
