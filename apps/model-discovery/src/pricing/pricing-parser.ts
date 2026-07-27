import { Decimal } from '@prisma/client-runtime-utils';
import {
  ParsedPricingProfile,
  ParsedPricingRate,
  PricingDiagnostic,
  PricingParseResult,
  PricingServiceTier,
  PricingUsage,
} from './pricing.types';

type UnknownRecord = Record<string, unknown>;

type RateDescriptor = {
  usage: PricingUsage;
  unit: string;
  quantity: number;
  resolution?: string;
};

const RESERVED_COST_KEYS = new Set([
  'region',
  'priority_pricing',
  'tiered_pricing',
]);

const TOKEN_QUANTITY = 1_000_000;

const explicitRateKeys: Record<string, RateDescriptor> = {
  input_cost_per_token: {
    usage: 'input',
    unit: 'token',
    quantity: TOKEN_QUANTITY,
  },
  output_cost_per_token: {
    usage: 'output',
    unit: 'token',
    quantity: TOKEN_QUANTITY,
  },
  cache_read_input_token_cost: {
    usage: 'cache_read',
    unit: 'token',
    quantity: TOKEN_QUANTITY,
  },
  cache_creation_input_token_cost: {
    usage: 'cache_write',
    unit: 'token',
    quantity: TOKEN_QUANTITY,
  },
  cache_creation_input_token_cost_per_hour: {
    usage: 'cache_write',
    unit: 'token_hour',
    quantity: TOKEN_QUANTITY,
  },
  input_cost_per_token_batches: {
    usage: 'input',
    unit: 'token',
    quantity: TOKEN_QUANTITY,
  },
  output_cost_per_token_batches: {
    usage: 'output',
    unit: 'token',
    quantity: TOKEN_QUANTITY,
  },
  cache_read_input_audio_token_cost: {
    usage: 'cache_read',
    unit: 'audio_token',
    quantity: TOKEN_QUANTITY,
  },
  cache_creation_input_audio_token_cost: {
    usage: 'cache_write',
    unit: 'audio_token',
    quantity: TOKEN_QUANTITY,
  },
};

const unitQuantities: Record<string, number> = {
  token: TOKEN_QUANTITY,
  audio_token: TOKEN_QUANTITY,
  video_token: TOKEN_QUANTITY,
  image_token: TOKEN_QUANTITY,
  second: 1,
  image: 1,
  request: 1,
  character: 1_000,
  query: 1,
  page: 1,
  annotated_page: 1,
};

function asObject(input: unknown): UnknownRecord | undefined {
  if (!input || typeof input !== 'object' || Array.isArray(input)) {
    return undefined;
  }
  return input as UnknownRecord;
}

function normalizeRegion(input: unknown): string {
  if (typeof input !== 'string') return 'global';
  const value = input.trim().toLowerCase();
  if (!value || value === '*' || value === 'global') return 'global';
  return value
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

function descriptorForKey(key: string): RateDescriptor | undefined {
  const explicit = explicitRateKeys[key];
  if (explicit) return explicit;

  const match = key.match(
    /^(input|output)_cost_per_(token|audio_token|video_token|image_token|second|image|request|character|query|page|annotated_page)(?:_(.+))?$/,
  );
  if (!match) return undefined;

  const usage = match[1] as 'input' | 'output';
  const unit = match[2];
  const suffix = match[3]?.trim().toLowerCase();
  if (suffix && unit !== 'second' && unit !== 'image') return undefined;

  return {
    usage,
    unit,
    quantity: unitQuantities[unit],
    resolution: suffix,
  };
}

function decimalAmount(
  raw: unknown,
  quantity: number,
): string | undefined {
  if (
    (typeof raw !== 'number' && typeof raw !== 'string') ||
    String(raw).trim() === ''
  ) {
    return undefined;
  }

  try {
    const value = new Decimal(String(raw));
    if (!value.isFinite() || value.isNegative()) return undefined;
    return value.mul(quantity).toString();
  } catch {
    return undefined;
  }
}

function rateKey(rate: Omit<ParsedPricingRate, 'rateKey'>): string {
  return [
    rate.usage,
    rate.unit,
    rate.serviceTier,
    `resolution=${rate.resolution ?? 'default'}`,
    `tier=${rate.tierFrom ?? 'base'}`,
  ].join(':');
}

function appendRate(
  rates: ParsedPricingRate[],
  diagnostics: PricingDiagnostic[],
  regionSlug: string,
  key: string,
  rawAmount: unknown,
  descriptor: RateDescriptor,
  serviceTier: PricingServiceTier,
  tier?: {
    from?: number;
    to?: number;
    mode?: string;
  },
): void {
  const amount = decimalAmount(rawAmount, descriptor.quantity);
  if (amount === undefined) {
    diagnostics.push({
      severity: 'error',
      code: 'invalid_amount',
      message: `Pricing key ${key} must be a finite non-negative number.`,
      regionSlug,
      key,
    });
    return;
  }

  const candidateWithoutKey: Omit<ParsedPricingRate, 'rateKey'> = {
    usage: descriptor.usage,
    unit: descriptor.unit,
    amount,
    quantity: descriptor.quantity,
    serviceTier,
    resolution: descriptor.resolution,
    tierFrom: tier?.from,
    tierTo: tier?.to,
    tierMode: tier?.mode,
  };
  const candidate: ParsedPricingRate = {
    ...candidateWithoutKey,
    rateKey: rateKey(candidateWithoutKey),
  };

  if (rates.some((rate) => rate.rateKey === candidate.rateKey)) {
    diagnostics.push({
      severity: 'error',
      code: 'duplicate_rate',
      message: `Duplicate pricing rate ${candidate.rateKey}.`,
      regionSlug,
      key,
    });
    return;
  }

  rates.push(candidate);
}

function parseScalarRates(
  cost: UnknownRecord,
  rates: ParsedPricingRate[],
  diagnostics: PricingDiagnostic[],
  regionSlug: string,
  serviceTier: PricingServiceTier,
): void {
  for (const [key, rawAmount] of Object.entries(cost)) {
    if (RESERVED_COST_KEYS.has(key)) continue;

    const descriptor = descriptorForKey(key);
    if (!descriptor) {
      diagnostics.push({
        severity: 'error',
        code: 'unknown_pricing_key',
        message: `Unknown pricing key ${key}.`,
        regionSlug,
        key,
      });
      continue;
    }

    const effectiveTier =
      key.endsWith('_batches') && serviceTier === 'standard'
        ? 'batch'
        : serviceTier;
    appendRate(
      rates,
      diagnostics,
      regionSlug,
      key,
      rawAmount,
      descriptor,
      effectiveTier,
    );
  }
}

function parseTieredRates(
  rawTieredPricing: unknown,
  rates: ParsedPricingRate[],
  diagnostics: PricingDiagnostic[],
  regionSlug: string,
): void {
  const tieredPricing = asObject(rawTieredPricing);
  if (!tieredPricing) return;

  const tierMode =
    typeof tieredPricing['pricing_mode'] === 'string'
      ? tieredPricing['pricing_mode']
      : undefined;
  const usageByGroup: Record<string, PricingUsage> = {
    input: 'input',
    output: 'output',
    cache_read: 'cache_read',
    cache_write: 'cache_write',
  };

  for (const [group, usage] of Object.entries(usageByGroup)) {
    const rawTiers = tieredPricing[group];
    if (rawTiers === undefined) continue;
    if (!Array.isArray(rawTiers)) {
      diagnostics.push({
        severity: 'error',
        code: 'invalid_tier',
        message: `Tier group ${group} must be an array.`,
        regionSlug,
        key: `tiered_pricing.${group}`,
      });
      continue;
    }

    const tiers = rawTiers
      .map((rawTier) => asObject(rawTier))
      .filter((tier): tier is UnknownRecord => Boolean(tier))
      .sort((left, right) => Number(left['from']) - Number(right['from']));

    for (let index = 0; index < tiers.length; index += 1) {
      const tier = tiers[index];
      const from = Number(tier['from']);
      const nextFrom =
        index + 1 < tiers.length ? Number(tiers[index + 1]['from']) : undefined;
      if (!Number.isInteger(from) || from < 0) {
        diagnostics.push({
          severity: 'error',
          code: 'invalid_tier',
          message: `Tier ${group} must have a non-negative integer from value.`,
          regionSlug,
          key: `tiered_pricing.${group}`,
        });
        continue;
      }

      appendRate(
        rates,
        diagnostics,
        regionSlug,
        `tiered_pricing.${group}.cost_per_token`,
        tier['cost_per_token'],
        { usage, unit: 'token', quantity: TOKEN_QUANTITY },
        'standard',
        {
          from,
          to: Number.isInteger(nextFrom) ? nextFrom : undefined,
          mode: tierMode,
        },
      );
    }
  }

  for (const key of Object.keys(tieredPricing)) {
    if (key !== 'pricing_mode' && !usageByGroup[key]) {
      diagnostics.push({
        severity: 'error',
        code: 'unknown_pricing_key',
        message: `Unknown tiered pricing group ${key}.`,
        regionSlug,
        key: `tiered_pricing.${key}`,
      });
    }
  }
}

export function parsePricingProfiles(
  rawCosts: unknown,
  options: {
    sourceUrl?: string;
    explicitFree?: boolean;
  } = {},
): PricingParseResult {
  const diagnostics: PricingDiagnostic[] = [];
  const profiles: ParsedPricingProfile[] = [];
  const costs = Array.isArray(rawCosts) ? rawCosts : [];

  for (const rawCost of costs) {
    const cost = asObject(rawCost);
    if (!cost) continue;
    const regionSlug = normalizeRegion(cost['region']);
    const rates: ParsedPricingRate[] = [];

    parseScalarRates(cost, rates, diagnostics, regionSlug, 'standard');

    const priorityPricing = asObject(cost['priority_pricing']);
    if (priorityPricing) {
      parseScalarRates(
        priorityPricing,
        rates,
        diagnostics,
        regionSlug,
        'priority',
      );
    }

    parseTieredRates(
      cost['tiered_pricing'],
      rates,
      diagnostics,
      regionSlug,
    );

    if (rates.length === 0) {
      diagnostics.push({
        severity: 'warning',
        code: 'empty_profile',
        message: 'Pricing profile contains no recognized rates.',
        regionSlug,
      });
      continue;
    }

    profiles.push({
      regionSlug,
      currency: 'USD',
      sourceUrl: options.sourceUrl,
      rates,
    });
  }

  const hasPositiveRate = profiles.some((profile) =>
    profile.rates.some((rate) => new Decimal(rate.amount).greaterThan(0)),
  );
  const status = options.explicitFree
    ? 'free'
    : hasPositiveRate
      ? 'priced'
      : 'unknown';

  return { profiles, diagnostics, status };
}
