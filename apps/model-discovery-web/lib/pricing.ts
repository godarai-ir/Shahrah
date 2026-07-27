export type PricingTier = {
  from?: number | null;
  to?: number | null;
  mode?: string | null;
};

export type PricingRate = {
  key: string;
  usage: string;
  unit: string;
  amount: string;
  quantity: number;
  serviceTier: string;
  resolution?: string | null;
  tier?: PricingTier | null;
  qualifiers?: Record<string, unknown>;
};

export type PricingProfile = {
  region: string;
  currency: string;
  effectiveFrom?: string | null;
  effectiveTo?: string | null;
  sourceUrl?: string | null;
  rates: PricingRate[];
};

export type PricingSummary = {
  usage: string;
  unit: string;
  currency: string;
  quantity: number;
  minimumAmount: string;
  maximumAmount: string;
  regions: string[];
  resolutions: string[];
  rateCount: number;
};

export function allPricingRates(profiles?: PricingProfile[] | null) {
  return (profiles ?? []).flatMap((profile) =>
    profile.rates.map((rate) => ({ profile, rate })),
  );
}

export function findPricingRate(
  profiles: PricingProfile[] | null | undefined,
  usage: string,
  unit = "token",
) {
  return allPricingRates(profiles).find(
    ({ rate }) =>
      rate.usage === usage &&
      rate.unit === unit &&
      rate.serviceTier === "standard" &&
      !rate.tier,
  );
}

export function primaryPricingRate(
  profiles: PricingProfile[] | null | undefined,
) {
  return (
    findPricingRate(profiles, "input", "token")?.rate ??
    findPricingRate(profiles, "output", "second")?.rate ??
    allPricingRates(profiles)[0]?.rate
  );
}

export function formatRateUnit(rate: PricingRate) {
  const quantity = rate.quantity === 1 ? "" : `${rate.quantity.toLocaleString("en-US")} `;
  const resolution = rate.resolution ? ` · ${rate.resolution}` : "";
  return `${quantity}${rate.unit}${resolution}`;
}
