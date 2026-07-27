export type PricingTier = {
  from?: number;
  to?: number;
  mode?: string;
};

export type PricingRate = {
  key: string;
  usage: "input" | "output" | "cache_read" | "cache_write" | string;
  unit: string;
  amount: string;
  quantity: number;
  service_tier: "standard" | "batch" | "priority" | string;
  resolution?: string;
  tier?: PricingTier;
  qualifiers?: Record<string, unknown>;
};

export type PricingProfile = {
  region: string;
  currency: string;
  effective_from?: string;
  effective_to?: string;
  source_url?: string;
  rates: PricingRate[];
};

export type ModelPrice = {
  status?: "unknown" | "priced" | "free" | string;
  profiles: PricingProfile[];
};

export type PricingDisplayOptions = {
  currency: "USD" | "TOMAN";
  scale: "PER_MILLION_TOKENS" | "PER_TOKEN";
  usdToIrr: number;
};

export type PricingEntry = {
  profile: PricingProfile;
  rate: PricingRate;
};

export function pricingEntries(price: ModelPrice | undefined): PricingEntry[] {
  if (!price || price.status === "unknown") return [];
  return price.profiles.flatMap((profile) =>
    profile.rates.map((rate) => ({ profile, rate })),
  );
}

export function pricingRatePresentation(
  entry: PricingEntry,
  display?: PricingDisplayOptions,
): { label: string; value: string } {
  const { profile, rate } = entry;
  let amount = rate.amount;
  let currency = profile.currency;
  let quantity = rate.quantity;

  if (
    display?.scale === "PER_TOKEN" &&
    rate.unit === "token" &&
    rate.quantity === 1_000_000
  ) {
    const numericAmount = Number(amount);
    if (Number.isFinite(numericAmount)) {
      amount = formatDecimal(numericAmount / 1_000_000);
      quantity = 1;
    }
  }

  if (
    display?.currency === "TOMAN" &&
    profile.currency === "USD" &&
    Number.isFinite(display.usdToIrr) &&
    display.usdToIrr > 0
  ) {
    const numericAmount = Number(amount);
    if (Number.isFinite(numericAmount)) {
      amount = new Intl.NumberFormat("en-US", {
        maximumFractionDigits: quantity === 1 ? 8 : 2,
      }).format((numericAmount * display.usdToIrr) / 10);
      currency = "TOMAN";
    }
  }

  const quantityLabel =
    quantity === 1 ? "" : `${quantity.toLocaleString("en-US")} `;
  const resolution = rate.resolution ? ` · ${rate.resolution}` : "";
  const serviceTier =
    rate.service_tier && rate.service_tier !== "standard"
      ? ` · ${rate.service_tier}`
      : "";
  const region = profile.region && profile.region !== "global"
    ? ` · ${profile.region}`
    : "";

  return {
    label: `${rate.usage} / ${quantityLabel}${rate.unit}${resolution}${serviceTier}${region}`,
    value: `${amount} ${currency}`,
  };
}

export function findPricingRate(
  price: ModelPrice | undefined,
  usage: string,
  unit = "token",
): PricingRate | undefined {
  return price?.profiles
    .flatMap((profile) => profile.rates)
    .find(
      (rate) =>
        rate.usage === usage &&
        rate.unit === unit &&
        (rate.service_tier === "standard" || !rate.service_tier) &&
        !rate.tier,
    );
}

export function primaryPricingRate(
  price: ModelPrice | undefined,
): PricingRate | undefined {
  return (
    findPricingRate(price, "input", "token") ??
    findPricingRate(price, "output", "second") ??
    price?.profiles.flatMap((profile) => profile.rates)[0]
  );
}

export function pricingSummaryText(
  price: ModelPrice | undefined,
  fallback: string,
): string {
  if (!price || price.status === "unknown") return fallback;
  const entry = pricingEntries(price).find(
    ({ rate }) => rate === primaryPricingRate(price),
  );
  if (!entry) return price.status === "free" ? "Free" : fallback;
  const presentation = pricingRatePresentation(entry);
  return `${presentation.value} · ${presentation.label}`;
}

export function pricingRateNumber(rate: PricingRate | undefined): number | undefined {
  if (!rate) return undefined;
  const value = Number(rate.amount);
  return Number.isFinite(value) ? value : undefined;
}

export function clonePricing(price: ModelPrice | undefined): ModelPrice | undefined {
  if (!price) return undefined;
  return {
    ...price,
    profiles: price.profiles.map((profile) => ({
      ...profile,
      rates: profile.rates.map((rate) => ({
        ...rate,
        tier: rate.tier ? { ...rate.tier } : undefined,
        qualifiers: rate.qualifiers ? { ...rate.qualifiers } : undefined,
      })),
    })),
  };
}

export function isPricingResolved(price: ModelPrice | undefined): boolean {
  if (price?.status === "unknown") return false;
  return Boolean(
    price?.profiles.some(
      (profile) =>
        profile.region.trim() &&
        profile.currency.trim() &&
        profile.rates.some(
          (rate) =>
            rate.key.trim() &&
            rate.usage.trim() &&
            rate.unit.trim() &&
            rate.quantity > 0 &&
            rate.amount.trim() !== "" &&
            Number.isFinite(Number(rate.amount)) &&
            Number(rate.amount) >= 0,
        ),
    ),
  );
}

function formatDecimal(value: number): string {
  return value.toFixed(12).replace(/\.?0+$/, "");
}
