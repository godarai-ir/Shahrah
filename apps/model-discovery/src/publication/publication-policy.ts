import { Decimal } from '@prisma/client-runtime-utils';

export type PublicationBlockerCode =
  | 'pricing_unknown'
  | 'free_model_has_positive_rate'
  | 'pricing_profile_required'
  | 'profile_region_required'
  | 'unsupported_currency'
  | 'invalid_effective_interval'
  | 'pricing_rate_required'
  | 'duplicate_rate_key'
  | 'invalid_rate'
  | 'unsupported_output_strategy';

export type PublicationBlocker = {
  code: PublicationBlockerCode;
  message: string;
  region?: string;
  rateKey?: string;
};

export type PublicationStrategy = {
  region: string;
  unit: 'image' | 'image_token' | 'token';
};

export type PublicationReadiness = {
  applicable: boolean;
  eligible: boolean;
  state: 'not_applicable' | 'needs_pricing' | 'ready_to_publish';
  blockers: PublicationBlocker[];
  strategies: PublicationStrategy[];
};

type RateInput = {
  key?: string | null;
  rateKey?: string | null;
  usage: string;
  unit: string;
  amount: string | number | Decimal;
  quantity: number;
  serviceTier?: string | null;
};

type ProfileInput = {
  region?: string | null;
  currency?: string | null;
  effectiveFrom?: string | Date | null;
  effectiveTo?: string | Date | null;
  rates: RateInput[];
};

export type PublicationModelInput = {
  mode: string[];
  pricingStatus: string;
  pricingProfiles: ProfileInput[];
};

const outputUnits = ['image', 'image_token', 'token'] as const;

export function evaluatePublicationReadiness(
  model: PublicationModelInput,
): PublicationReadiness {
  if (!model.mode.map(normalize).includes('image')) {
    return {
      applicable: false,
      eligible: true,
      state: 'not_applicable',
      blockers: [],
      strategies: [],
    };
  }

  const blockers: PublicationBlocker[] = [];
  const strategies: PublicationStrategy[] = [];

  if (model.pricingStatus === 'free') {
    for (const profile of model.pricingProfiles) {
      for (const rate of profile.rates) {
        if (validDecimal(rate.amount)?.greaterThan(0)) {
          blockers.push({
            code: 'free_model_has_positive_rate',
            message: 'A free model cannot contain a positive pricing rate.',
            region: profile.region ?? undefined,
            rateKey: rate.key ?? rate.rateKey ?? undefined,
          });
        }
      }
    }
    return result(blockers, strategies);
  }

  if (model.pricingStatus !== 'priced') {
    blockers.push({
      code: 'pricing_unknown',
      message: 'Pricing status must be priced or explicitly free.',
    });
    return result(blockers, strategies);
  }

  if (model.pricingProfiles.length === 0) {
    blockers.push({
      code: 'pricing_profile_required',
      message: 'At least one pricing profile is required.',
    });
    return result(blockers, strategies);
  }

  for (const profile of model.pricingProfiles) {
    const region = String(profile.region ?? '').trim();
    if (!region) {
      blockers.push({
        code: 'profile_region_required',
        message: 'Pricing profile region is required.',
      });
    }
    if (normalize(profile.currency ?? '') !== 'usd') {
      blockers.push({
        code: 'unsupported_currency',
        message: 'Public image pricing must use USD.',
        region: region || undefined,
      });
    }
    const from = toTime(profile.effectiveFrom);
    const to = toTime(profile.effectiveTo);
    if (from !== undefined && to !== undefined && to <= from) {
      blockers.push({
        code: 'invalid_effective_interval',
        message: 'effectiveTo must be later than effectiveFrom.',
        region: region || undefined,
      });
    }
    if (profile.rates.length === 0) {
      blockers.push({
        code: 'pricing_rate_required',
        message: 'Each pricing profile requires at least one rate.',
        region: region || undefined,
      });
      continue;
    }

    const keys = new Set<string>();
    for (const rate of profile.rates) {
      const key = String(rate.key ?? rate.rateKey ?? '').trim();
      if (!key || keys.has(key)) {
        blockers.push({
          code: 'duplicate_rate_key',
          message: key
            ? `Duplicate rate key ${key}.`
            : 'Every rate requires a stable key.',
          region: region || undefined,
          rateKey: key || undefined,
        });
      }
      keys.add(key);
      if (
        !String(rate.usage).trim() ||
        !String(rate.unit).trim() ||
        !Number.isInteger(rate.quantity) ||
        rate.quantity <= 0 ||
        validDecimal(rate.amount) === undefined
      ) {
        blockers.push({
          code: 'invalid_rate',
          message:
            'Rate usage/unit, non-negative decimal amount and positive quantity are required.',
          region: region || undefined,
          rateKey: key || undefined,
        });
      }
    }

    const strategy = outputUnits.find((unit) =>
      profile.rates.some(
        (rate) =>
          normalize(rate.usage) === 'output' &&
          normalize(rate.unit) === unit &&
          normalize(rate.serviceTier ?? 'standard') === 'standard',
      ),
    );
    if (!strategy) {
      blockers.push({
        code: 'unsupported_output_strategy',
        message:
          'A standard output/image, output/image_token or output/token rate is required.',
        region: region || undefined,
      });
    } else {
      strategies.push({ region: region || 'global', unit: strategy });
    }
  }

  return result(blockers, strategies);
}

function result(
  blockers: PublicationBlocker[],
  strategies: PublicationStrategy[],
): PublicationReadiness {
  const eligible = blockers.length === 0;
  return {
    applicable: true,
    eligible,
    state: eligible ? 'ready_to_publish' : 'needs_pricing',
    blockers,
    strategies,
  };
}

function normalize(value: unknown): string {
  return String(value ?? '').trim().toLowerCase();
}

function validDecimal(value: string | number | Decimal): Decimal | undefined {
  try {
    const decimal = new Decimal(value);
    return decimal.isFinite() && !decimal.isNegative() ? decimal : undefined;
  } catch {
    return undefined;
  }
}

function toTime(value: string | Date | null | undefined): number | undefined {
  if (!value) return undefined;
  const time = value instanceof Date ? value.getTime() : new Date(value).getTime();
  return Number.isFinite(time) ? time : undefined;
}
