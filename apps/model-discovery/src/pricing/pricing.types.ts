export type PricingUsage =
  | 'input'
  | 'output'
  | 'cache_read'
  | 'cache_write';

export type PricingServiceTier = 'standard' | 'batch' | 'priority';

export type PricingStatus = 'unknown' | 'priced' | 'free';

export type ParsedPricingRate = {
  rateKey: string;
  usage: PricingUsage;
  unit: string;
  amount: string;
  quantity: number;
  serviceTier: PricingServiceTier;
  resolution?: string;
  tierFrom?: number;
  tierTo?: number;
  tierMode?: string;
  qualifiers?: Record<string, string | number | boolean>;
};

export type ParsedPricingProfile = {
  regionSlug: string;
  currency: string;
  sourceUrl?: string;
  rates: ParsedPricingRate[];
};

export type PricingDiagnostic = {
  severity: 'error' | 'warning';
  code:
    | 'duplicate_rate'
    | 'empty_profile'
    | 'invalid_amount'
    | 'invalid_tier'
    | 'unknown_pricing_key';
  message: string;
  regionSlug: string;
  key?: string;
};

export type PricingParseResult = {
  profiles: ParsedPricingProfile[];
  diagnostics: PricingDiagnostic[];
  status: PricingStatus;
};
