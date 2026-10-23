// Contracts for the control-plane credit billing endpoints (packages/contracts/platform/credit_admin.go).
// Credit amounts are decimal strings with two places; prices are toman per credit.

export type CreditSettings = {
  billing_unit: string;
  overage_allowed: boolean;
  overage_enabled: boolean;
  overage_cap_credits: string | null;
};

export type TenantCreditOverview = {
  tenant_id: string;
  billing_unit: string;
  available_credits: string;
  debt_credits: string;
  ledger_sum_credits: string;
  settings: CreditSettings;
};

export type CreditLedgerEntry = {
  entry_id: string;
  kind: string;
  amount_credits: string;
  occurred_at: string;
  lot_id?: string;
  request_id?: string;
  model?: string;
  user_id?: string;
  team_id?: string;
  product?: string;
  billing_mode?: string;
  credit_equivalent?: string;
  reason?: string;
  actor?: string;
};

export type CreditLot = {
  lot_id: string;
  source: string;
  amount_credits: string;
  remaining_credits: string;
  expires_at?: string;
  purchase_price_irr?: number;
  price_per_credit_irr?: string;
  source_ref?: Record<string, unknown>;
  created_at: string;
};

export type CreditLotAllocation = { entry_id: string; kind: string; amount_credits: string; occurred_at: string };

export type CreditPriceRevision = {
  id: string;
  price_toman_per_credit: number;
  previous_price_toman?: number;
  base_usd_toman: string;
  observed_avg_usd_toman?: string;
  band_pct: string;
  status: string;
  effective_from?: string;
  plan_effective_from?: string;
  proposed_by?: string;
  decided_by?: string;
  decided_at?: string;
  decision_reason?: string;
  created_at: string;
};

export type CreditPriceStatus = {
  active: CreditPriceRevision | null;
  proposed: CreditPriceRevision | null;
  seven_day_average_usd_toman: string | null;
  revisions: CreditPriceRevision[];
};

export type CreditModelMultiplier = { model: string; multiplier: string; reason: string; created_by?: string; created_at: string };

export type CreditRateCardVersion = {
  version: number;
  status: string;
  effective_from: string;
  notice_until?: string;
  changelog_fa: string;
  changelog_en: string;
  published_by?: string;
  published_at?: string;
  multipliers: CreditModelMultiplier[];
};

export type CreditPackVersion = {
  slug: string;
  version: number;
  credits: number;
  discount_percent: string;
  validity_months: number;
  status: string;
  sort_order: number;
  created_by?: string;
  created_at: string;
};

const CREDIT_AMOUNT = /^-?\d+(\.\d{1,2})?$/;

/** Validates a signed credit amount with at most two decimals, rejecting zero. */
export function parseCreditAmount(value: string): string | null {
  const trimmed = value.trim();
  if (!CREDIT_AMOUNT.test(trimmed) || Number(trimmed) === 0) return null;
  return trimmed;
}

export function formatCredits(value: string): string {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return value;
  return `${new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(parsed)} cr`;
}

export function formatToman(value: number | string): string {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return String(value);
  return `${new Intl.NumberFormat('en-US').format(Math.round(parsed))} toman`;
}

export type CreditConversionPreview = {
  tenant_id: string;
  tenant_slug: string;
  billing_unit: string;
  price_toman_per_credit: number;
  wallet_irr: number;
  unsettled_irr: number;
  credits: string;
  residue_irr: number;
  rules_to_convert: number;
  plan_slug?: string;
  billing_interval?: string;
  new_plan_price_irr?: number;
  new_plan_monthly_credits?: string;
  warnings: string[];
};

export type CreditConversion = {
  id: string;
  tenant_id: string;
  sequence: number;
  direction: 'to_credit' | 'to_irr';
  price_toman_per_credit: number;
  wallet_irr: number;
  converted_irr: number;
  residue_irr: number;
  credits: string;
  debt: string;
  uncollected_irr: number;
  rules_converted: number;
  plan_slug?: string;
  subscription_id?: string;
  actor: string;
  reason: string;
  created_at: string;
};

export type TenantCreditConversionStatus = { preview: CreditConversionPreview; conversions: CreditConversion[] };

export type CreditFleetConversionResult = {
  converted: CreditConversion[];
  failed: { tenant_id: string; tenant_slug: string; error: string }[];
  remaining: number;
};

const CONVERSION_WARNINGS: Record<string, string> = {
  no_subscription: 'No active subscription: the plan is not moved.',
  no_published_credit_plan: 'The plan has no published credit version: publish it first or the plan is not moved.',
  pending_wallet_topup: 'A wallet top-up is pending; if it completes later, convert the tenant again.',
  pending_bank_transfer: 'A wallet bank transfer awaits review; approving it later needs another conversion.',
  unsettled_irr_usage: 'IRR usage from this month is settled from the wallet first.',
  already_credit: 'Already billed in credits; only a remaining wallet balance can be converted.',
};

export function conversionWarning(code: string): string {
  return CONVERSION_WARNINGS[code] ?? code;
}

/** Warnings that leave the tenant on its current plan instead of moving it. */
export function blocksPlanMove(preview: CreditConversionPreview): boolean {
  return preview.warnings.some((code) => code === 'no_subscription' || code === 'no_published_credit_plan');
}
